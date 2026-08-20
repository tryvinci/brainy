package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const providerExtractionVersion = "provider-v4-ops"

// gpt-oss (and similar reasoning models) spend hundreds of tokens on a
// reasoning channel before emitting JSON. Cloudflare's default max_tokens is
// 256, which finishes with content=null and "empty completion".
const providerMaxCompletionTokens = 4096

// ProviderConfig configures an OpenAI-compatible chat completions client.
type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
	Strict  bool
}

func (c ProviderConfig) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

// ProviderExtractor calls an OpenAI-compatible /v1/chat/completions endpoint
// and validates structured JSON memories. It never mutates raw ingest state.
type ExtractorStats struct {
	Calls     int64 `json:"calls"`
	Failures  int64 `json:"failures"`
	Fallbacks int64 `json:"fallbacks"`
}

type ExtractorIdentity struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Strict   bool   `json:"strict"`
}

func (i ExtractorIdentity) Signature() string {
	return strings.TrimSpace(i.Provider) + "|" + strings.TrimSpace(i.Model)
}

// ProviderExtractor calls an OpenAI-compatible /v1/chat/completions endpoint
// and validates structured JSON memories. It never mutates raw ingest state.
type ProviderExtractor struct {
	client    *http.Client
	cfg       ProviderConfig
	fallback  DeterministicExtractor
	calls     atomic.Int64
	failures  atomic.Int64
	fallbacks atomic.Int64
}

func NewProviderExtractor(cfg ProviderConfig, client *http.Client) *ProviderExtractor {
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 45 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	return &ProviderExtractor{
		client:   client,
		cfg:      cfg,
		fallback: NewDeterministicExtractor(),
	}
}

func (p *ProviderExtractor) WithStrict(strict bool) *ProviderExtractor {
	p.cfg.Strict = strict
	return p
}

func (p *ProviderExtractor) Identity() ExtractorIdentity {
	if !p.cfg.Configured() {
		return ExtractorIdentity{
			Name:     "deterministic-v1",
			Provider: "deterministic",
			Model:    "deterministic-v1",
			Strict:   p.cfg.Strict,
		}
	}
	return ExtractorIdentity{
		Name:     "provider-extract:" + p.cfg.Model,
		Provider: "openai-compatible",
		Model:    strings.TrimSpace(p.cfg.Model),
		Strict:   p.cfg.Strict,
	}
}

func (p *ProviderExtractor) Stats() ExtractorStats {
	return ExtractorStats{
		Calls:     p.calls.Load(),
		Failures:  p.failures.Load(),
		Fallbacks: p.fallbacks.Load(),
	}
}

func (p *ProviderExtractor) Extract(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	p.calls.Add(1)
	req.Messages = EnrichImageText(ctx, req.Messages)
	// Deterministic baseline always runs first (ENG-92).
	baseline, err := p.fallback.Extract(ctx, req)
	if err != nil {
		p.failures.Add(1)
		return nil, err
	}
	if !p.cfg.Configured() {
		if p.cfg.Strict {
			p.failures.Add(1)
			return nil, fmt.Errorf("extraction provider required in strict mode")
		}
		p.fallbacks.Add(1)
		return baseline, nil
	}

	providerMemories, err := p.extractProvider(ctx, req)
	if err != nil {
		p.failures.Add(1)
		if p.cfg.Strict {
			return nil, err
		}
		if len(baseline) > 0 {
			p.fallbacks.Add(1)
			return baseline, nil
		}
		return nil, err
	}
	return filterAssistantRecallEpisodes(req, mergeProviderAndBaseline(baseline, providerMemories, WriteMutationModeOf(req))), nil
}

func (p *ProviderExtractor) extractProvider(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	body, err := json.Marshal(map[string]any{
		"model":           p.cfg.Model,
		"temperature":     0,
		"max_tokens":      providerMaxCompletionTokens,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": providerSystemPrompt},
			{"role": "user", "content": buildProviderUserPrompt(req)},
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(p.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(p.cfg.APIKey); key != "" {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("provider extract request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("provider extract read: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider extract status %d: %s", resp.StatusCode, truncate(string(respBody), 240))
	}

	text, err := completionText(respBody)
	if err != nil {
		return nil, err
	}
	return parseProviderMemories(text)
}

type chatCompletionMessage struct {
	Content          json.RawMessage `json:"content"`
	Reasoning        string          `json:"reasoning"`
	ReasoningContent string          `json:"reasoning_content"`
}

func completionText(respBody []byte) (string, error) {
	var completion struct {
		Choices []struct {
			Message chatCompletionMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return "", fmt.Errorf("provider extract decode: %w", err)
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("provider extract empty completion")
	}
	msg := completion.Choices[0].Message
	text := decodeChatContent(msg.Content)
	if strings.TrimSpace(text) == "" {
		text = strings.TrimSpace(msg.Reasoning)
	}
	if text == "" {
		text = strings.TrimSpace(msg.ReasoningContent)
	}
	if text == "" {
		return "", fmt.Errorf("provider extract empty completion")
	}
	if extracted := firstJSONObject(text); extracted != "" {
		return extracted, nil
	}
	return text, nil
}

func decodeChatContent(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, part := range parts {
			b.WriteString(part.Text)
		}
		return b.String()
	}
	return strings.TrimSpace(string(raw))
}

func firstJSONObject(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "{") {
		return raw
	}
	start := strings.Index(raw, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}
	return ""
}

// Typed extraction v4: ADD/UPDATE/DELETE/NONE ops adapted from Mem0's classic
// extract merge decisions (Apache-2.0 semantics; not a verbatim prompt copy).
// Flat content remains retrieval-primary; typed slots still flow Explain → Metadata.
const providerSystemPrompt = `You are a Memory Analyzer extracting atomic memories and deciding how each relates to Prior context.

Return JSON only:
{"memories":[{"event":"ADD|UPDATE|DELETE|NONE","target_memory_id":"optional id from Prior context","kind":"fact|preference|profile","content":"...","source_text":"...","confidence":0.0,"subject":"optional entity","predicate":"optional taxonomy slot","value":"optional typed value","assertion_kind":"explicit|observed|inferred|corrective|negative","when":"optional absolute date","duration":"optional"}]}

Event rules (Mem0-style):
- ADD: new durable fact not already covered by Prior context.
- UPDATE: revises an existing prior memory; set target_memory_id to that [id]; content is the replacement fact; assertion_kind=corrective when contradicting.
- DELETE: prior memory is obsolete/wrong with no replacement; set target_memory_id; content may briefly state why.
- NONE: duplicate or non-durable; skip storing (still may list for traceability).

Predicates (use when a typed slot fits; otherwise omit predicate/value):
identity, relationship_status, origin, residence, occupation, education,
family_member, activity, activity_purpose, event, media_consumed, preference,
possession, health, plan, belief, skill, affiliation, contact_fact, metric.

CRITICAL RULES:
1. content must be a self-contained factual sentence usable months later WITHOUT the original turn.
2. In multi-speaker dialogue ("Name: ..."), first-person facts ("I am…") belong to that speaker.
   When a turn reports a fact about another named person (Name researched / works as / lives in / realized that / is a),
   attribute the memory to that person, not the reporter.
   In two-party dialogue, second-person facts ("you are…", "you researched…") belong to the addressee.
3. Emit ONE memory per distinct attribute, activity, place, titled work, preference, or plan.
   Split compound utterances. Prefer many small atoms over one long paraphrase.
4. When possible set subject + predicate + value (normalized short value, not the full sentence).
5. Always extract when present:
   - identity / role / relationship status
   - origin / originally from / moved from <place> (include country/city names literally)
   - when the speaker says "home country" and Prior context already has a country, bind that country
   - activities, hobbies, ways people unwind, and places tied to those activities (one memory per place)
   - book/movie titles in quotes OR capitalized titles after read/reading (verbatim spans)
   - visible text in [visible text: ...] after "this book I read" / "this novel"
   - family members' preferences (e.g. "Sam's kids like astronomy")
   - after a kids/children mention, bind later "they were excited/stoked for X" / "they love X" to those kids
   - career plans, fields of study, certifications
   - topics a speaker researched or is researching (the topic, not the motive)
   - collections / possessions ("collects classic children's books")
   - durations ("for 4 years")
6. Resolve relative time against Observation Date (yesterday, last Fri, last week, this month).
   Put the absolute date in "when" AND in content. Also include the session-relative phrase
   when the utterance is last-<weekday> or last week (e.g. "the Monday before 3 March 2024").
7. Skip pure greetings/acks ("Thanks!", "Yeah, Name", "Cool").
8. When Prior context lists [memory_id] lines: prefer UPDATE/DELETE over duplicate ADD for the same subject/predicate; keep subject stable.
9. When in doubt, EXTRACT as ADD — missed atoms destroy multi-attribute recall.
10. return {"memories":[]} only if nothing durable exists.
11. ADD-only for new durable facts. Use UPDATE only when Prior context already has the same subject+predicate with a different value; set when to the observation or stated date.
12. When a speaker or assistant reports a fact about another named person (Name lives in / works as / researched / realized that / is a), subject is that person, never the reporter or the assistant.
13. Durable assistant-role statements about a named person are facts about that person. Do not store the assistant's recap as an assistant-owned memory.`

func buildProviderUserPrompt(req IngestRequest) string {
	var b strings.Builder
	b.WriteString("source_type: ")
	b.WriteString(req.SourceType)
	b.WriteString("\n")
	if v := strings.TrimSpace(req.Vertical); v != "" {
		b.WriteString("vertical: ")
		b.WriteString(v)
		b.WriteString("\n")
	}
	if req.Metadata != nil {
		if raw, ok := req.Metadata["observed_at"]; ok && raw != nil {
			b.WriteString("Observation Date: ")
			b.WriteString(fmt.Sprint(raw))
			b.WriteString("\n")
			b.WriteString("Resolve ALL relative time phrases against Observation Date only.\n")
		}
		if ctxBlock, ok := req.Metadata["extract_context"].(string); ok && strings.TrimSpace(ctxBlock) != "" {
			b.WriteString("\nPrior context (do not repeat verbatim; use to link, update, or contradict):\n")
			b.WriteString(strings.TrimSpace(ctxBlock))
			b.WriteString("\n")
		}
	}
	b.WriteString("New Messages:\n")
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "" {
			role = "user"
		}
		b.WriteString("- ")
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	return b.String()
}

type providerMemoryPayload struct {
	Memories []providerMemoryItem `json:"memories"`
}

type providerMemoryItem struct {
	Event          string     `json:"event"`
	TargetMemoryID string     `json:"target_memory_id"`
	Kind           string     `json:"kind"`
	Content        string     `json:"content"`
	SourceText     string     `json:"source_text"`
	Confidence     float64    `json:"confidence"`
	Subject        string     `json:"subject"`
	Predicate      string     `json:"predicate"`
	Value          flexString `json:"value"`
	AssertionKind  string     `json:"assertion_kind"`
	When           string     `json:"when"`
	Duration       string     `json:"duration"`
}

// flexString accepts JSON strings or numbers (gpt-oss emits value: 4).
type flexString string

func (s *flexString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*s = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = flexString(str)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*s = flexString(n.String())
		return nil
	}
	var bo bool
	if err := json.Unmarshal(b, &bo); err == nil {
		*s = flexString(strconv.FormatBool(bo))
		return nil
	}
	return fmt.Errorf("flexString: unsupported JSON %s", truncate(string(b), 40))
}

func parseProviderMemories(raw string) ([]ExtractedMemory, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	var payload providerMemoryPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("provider extract json: %w", err)
	}

	out := make([]ExtractedMemory, 0, len(payload.Memories))
	for _, item := range payload.Memories {
		event := normalizeMemoryEvent(item.Event)
		if event == "" {
			if strings.EqualFold(strings.TrimSpace(item.AssertionKind), "corrective") && strings.TrimSpace(item.TargetMemoryID) != "" {
				event = MemoryEventUpdate
			} else {
				event = MemoryEventAdd
			}
		}
		kind, pred := coerceProviderKind(item.Kind, item.Predicate)
		if pred != "" {
			item.Predicate = pred
		}
		content := NormalizeText(item.Content)
		if content == "" {
			if event == MemoryEventDelete || event == MemoryEventNone {
				content = strings.TrimSpace(item.TargetMemoryID)
			}
			if content == "" && event != MemoryEventDelete && event != MemoryEventNone {
				return nil, fmt.Errorf("provider extract: empty content")
			}
			if content == "" {
				content = "memory_op:" + event
			}
		}
		source := NormalizeText(item.SourceText)
		if source == "" {
			source = content
		}
		confidence := item.Confidence
		if confidence <= 0 || confidence > 1 {
			confidence = 0.8
		}
		explain := map[string]any{
			"rule":         "provider_extract",
			"memory_event": event,
		}
		if tid := strings.TrimSpace(item.TargetMemoryID); tid != "" {
			explain["target_memory_id"] = tid
		}
		if subj := strings.TrimSpace(item.Subject); subj != "" {
			explain["subject"] = subj
		}
		if pred := normalizeProviderPredicate(item.Predicate); pred != "" {
			explain["predicate"] = pred
			val := strings.TrimSpace(string(item.Value))
			if val == "" {
				val = content
			}
			explain["value_norm"] = strings.ToLower(NormalizeText(val))
		}
		if ak := strings.ToLower(strings.TrimSpace(item.AssertionKind)); ak != "" {
			explain["assertion_kind"] = ak
		}
		if when := strings.TrimSpace(item.When); when != "" {
			explain["when"] = when
			explain["event_start"] = when
		}
		if duration := strings.TrimSpace(item.Duration); duration != "" {
			explain["duration"] = duration
		}
		out = append(out, ExtractedMemory{
			Kind:       kind,
			Content:    content,
			SourceText: source,
			Confidence: confidence,
			Explain:    explain,
			When:       strings.TrimSpace(item.When),
			Duration:   strings.TrimSpace(item.Duration),
		})
	}
	return out, nil
}

func mergeProviderAndBaseline(baseline, provider []ExtractedMemory, mode WriteMutationMode) []ExtractedMemory {
	if len(provider) == 0 {
		return baseline
	}
	filteredBaseline := baseline
	if mode == WriteModeGoverned {
		// Provider NONE/DELETE/UPDATE are authoritative for their semantic key /
		// source span — drop matching baseline items before content-key merge.
		suppressKeys, suppressSources, suppressContents := providerSuppressSets(provider)
		filteredBaseline = make([]ExtractedMemory, 0, len(baseline))
		for _, item := range baseline {
			if baselineSuppressedByProvider(item, suppressKeys, suppressSources, suppressContents) {
				continue
			}
			filteredBaseline = append(filteredBaseline, item)
		}
	}

	out := make([]ExtractedMemory, 0, len(provider)+len(filteredBaseline))
	seen := make(map[string]struct{}, len(provider)+len(filteredBaseline))
	for _, item := range provider {
		key := NormalizeText(item.Content)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	for _, item := range filteredBaseline {
		key := NormalizeText(item.Content)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func providerSuppressSets(provider []ExtractedMemory) (keys, sources, contents map[string]struct{}) {
	keys = make(map[string]struct{})
	sources = make(map[string]struct{})
	contents = make(map[string]struct{})
	for _, item := range provider {
		switch MemoryEventOf(item) {
		case MemoryEventNone, MemoryEventDelete, MemoryEventUpdate:
		default:
			continue
		}
		if sk := semanticSlotKey(item); sk != "" {
			keys[sk] = struct{}{}
		}
		if src := NormalizeText(item.SourceText); src != "" {
			sources[src] = struct{}{}
		}
		if c := NormalizeText(item.Content); c != "" {
			contents[c] = struct{}{}
		}
	}
	return keys, sources, contents
}

func semanticSlotKey(item ExtractedMemory) string {
	subj, _ := item.Explain["subject"].(string)
	pred, _ := item.Explain["predicate"].(string)
	subj = strings.ToLower(NormalizeText(subj))
	pred = strings.ToLower(NormalizeText(pred))
	if subj == "" || pred == "" {
		return ""
	}
	return subj + "|" + pred
}

func baselineSuppressedByProvider(item ExtractedMemory, keys, sources, contents map[string]struct{}) bool {
	if sk := semanticSlotKey(item); sk != "" {
		if _, ok := keys[sk]; ok {
			return true
		}
	}
	src := NormalizeText(item.SourceText)
	content := NormalizeText(item.Content)
	if src != "" {
		if _, ok := sources[src]; ok {
			return true
		}
		for psrc := range sources {
			if sourceSpanOverlap(src, psrc) {
				return true
			}
		}
	}
	if content != "" {
		if _, ok := contents[content]; ok {
			return true
		}
		// UPDATE/NONE paraphrases: baseline content overlapping provider content.
		for pc := range contents {
			if sourceSpanOverlap(content, pc) {
				return true
			}
		}
	}
	return false
}

// sourceSpanOverlap is true when one normalized span contains the other
// (or they share a long enough token run) so paraphrases of the same utterance match.
func sourceSpanOverlap(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	aToks := strings.Fields(a)
	bToks := strings.Fields(b)
	if len(aToks) < 3 || len(bToks) < 3 {
		return false
	}
	set := make(map[string]struct{}, len(bToks))
	for _, t := range bToks {
		set[t] = struct{}{}
	}
	shared := 0
	for _, t := range aToks {
		if _, ok := set[t]; ok {
			shared++
		}
	}
	minLen := len(aToks)
	if len(bToks) < minLen {
		minLen = len(bToks)
	}
	return shared*2 >= minLen && shared >= 3 // >= ~50% of shorter, at least 3 tokens
}

func coerceProviderKind(kind, predicate string) (string, string) {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case KindFact, KindPreference, KindProfile:
		return k, predicate
	case "":
		return KindFact, predicate
	}
	if pred := normalizeProviderPredicate(k); pred != "" && strings.TrimSpace(predicate) == "" {
		return KindFact, pred
	}
	return KindFact, predicate
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func normalizeProviderPredicate(raw string) string {
	p := strings.ToLower(strings.TrimSpace(raw))
	p = strings.ReplaceAll(p, "-", "_")
	p = strings.ReplaceAll(p, " ", "_")
	switch p {
	case PredicateIdentity, PredicateRelationshipStatus, PredicateOrigin, PredicateResidence,
		PredicateOccupation, PredicateEducation, PredicateFamilyMember, PredicateActivity,
		PredicateActivityPurpose, PredicateEvent, PredicateMediaConsumed, PredicatePreference,
		PredicatePossession, PredicateHealth, PredicatePlan, PredicateBelief, PredicateSkill,
		PredicateAffiliation, PredicateContactFact, PredicateMetric:
		return p
	default:
		return ""
	}
}
