package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

// PacketItem is one structured evidence unit for hybrid reading and typed packets.
type PacketItem struct {
	EvidenceID string   `json:"evidence_id,omitempty"`
	MemoryID   string   `json:"memory_id,omitempty"`
	FactID     string   `json:"fact_id,omitempty"`
	Content    string   `json:"content"`
	Predicate  string   `json:"predicate,omitempty"`
	Subject    string   `json:"subject,omitempty"`
	Value      string   `json:"value,omitempty"`
	EntityID   string   `json:"entity_id,omitempty"`
	Span       string   `json:"source_span,omitempty"`
	Role       string   `json:"role,omitempty"` // direct | bridge | temporal | context
	Score      float64  `json:"score,omitempty"`
	Targets    []string `json:"targets,omitempty"`
}

// HybridReaderConfig enables a bounded LLM reader over EvidencePacket only.
type HybridReaderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func (c HybridReaderConfig) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

func hybridReaderEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("BRAINY_RECALL_LLM")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func (s *Service) WithHybridReader(cfg HybridReaderConfig) *Service {
	s.hybridReader = cfg
	return s
}

// hybridReaderResult carries answer text plus observability for explain.
type hybridReaderResult struct {
	Answer            string
	OK                bool
	Attempted         bool
	Reason            string
	SupportingIDs     []string
	UnresolvedTargets []string
	Abstain           bool
	ParseMode         string // json | freeform | ""
	RawSnippet        string
}

func (s *Service) resolveHybridReaderConfig() HybridReaderConfig {
	cfg := s.hybridReader
	if cfg.Configured() {
		return cfg
	}
	cfg = HybridReaderConfig{
		BaseURL: strings.TrimSpace(os.Getenv("BRAINY_PROVIDER_BASE_URL")),
		APIKey:  strings.TrimSpace(os.Getenv("BRAINY_PROVIDER_API_KEY")),
		Model:   strings.TrimSpace(os.Getenv("BRAINY_PROVIDER_MODEL")),
		Timeout: 45 * time.Second,
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	}
	if cfg.APIKey == "" {
		cfg.APIKey = strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	}
	if cfg.Model == "" {
		cfg.Model = strings.TrimSpace(os.Getenv("LLM_MODEL"))
	}
	return cfg
}

func shouldAttemptHybrid(query string, plan QueryPlan, pkt EvidencePacket) (bool, string) {
	// Calendar temporal_answer is not the last word on "when …" questions —
	// hybrid may keep a weekday-relative phrasing the typed date collapsed.
	if plan.NeedsTemporal && pkt.TemporalAnswer != "" && !plan.NeedsMultiHop && !looksWhenEventQuery(query) {
		return false, "temporal_resolved"
	}
	if plan.NeedsMultiHop || plan.NeedsEnumeration || plan.PrimaryIntent == IntentMultiHop || looksWhenEventQuery(query) {
		if len(pkt.Items) == 0 && len(pkt.Contents) == 0 {
			return false, "empty_packet"
		}
		return true, "composition_needed"
	}
	// Point-fact / open-domain: keep deterministic packet join when coverage is
	// already satisfied on a small packet (routing invariant).
	if packetCoverageSatisfied(plan, pkt) && len(pkt.Contents) <= 2 {
		return false, "point_fact_packet_ok"
	}
	if len(pkt.Items) == 0 && len(pkt.Contents) == 0 {
		return false, "empty_packet"
	}
	return true, "weak_packet"
}

func formatHybridMemoryLines(pkt EvidencePacket) []string {
	return formatHybridMemoryLinesForQuery("", pkt)
}

func formatHybridMemoryLinesForQuery(query string, pkt EvidencePacket) []string {
	lines := make([]string, 0, len(pkt.Items)+len(pkt.Contents)+8)
	seen := map[string]struct{}{}
	var hops []HopResult
	if raw, ok := pkt.Coverage["hop_results"]; ok {
		hops, _ = raw.([]HopResult)
	}
	skipSlots := skipUnrelatedHopSlots(query, hops, pkt)
	coverToks := leftoverNonEntityQueryTokens(query, hops)
	add := func(content, memoryID string) {
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		if strings.HasSuffix(content, "?") && !strings.ContainsAny(content, "0123456789") {
			return
		}
		if skipSlots && len(coverToks) > 0 && !contentCoversAnyQueryToken(content, coverToks) {
			// Identity dumps: keep only leftover-covering memories.
			// Activity/event dumps: drop crowded comma lists, keep specific
			// facts whose gold is a synonym of the leftover token (UK / country).
			if hopsAreIdentityOnly(hops) || looksCrowdedHopDump(content) {
				return
			}
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		if memoryID != "" {
			lines = append(lines, "- ["+memoryID+"] "+content)
		} else {
			lines = append(lines, "- "+content)
		}
	}
	if raw, ok := pkt.Coverage["hop_results"]; ok {
		if hops, ok := raw.([]HopResult); ok && len(hops) > 0 {
			skipSlots := skipUnrelatedHopSlots(query, hops, pkt)
			if vals := hopSlotValues(hops); len(vals) > 0 && !skipSlots {
				lines = append(lines, "Structured:")
				for _, h := range hops {
					if h.Kind == "resolve_entity" || h.Source == "unresolved" {
						continue
					}
					slot := firstNonEmpty(h.Predicate, h.OutputKey, h.Kind)
					val := firstNonEmpty(h.Value, "")
					if val == "" && len(h.Values) > 0 {
						val = strings.Join(h.Values, ", ")
					}
					if strings.TrimSpace(val) == "" {
						continue
					}
					lines = append(lines, "- "+slot+" = "+val)
				}
			}
			if !hopDumpsUnproven(hops) && !skipSlots {
				lines = append(lines, "Hop chain:")
				for _, h := range hops {
					dep := ""
					if len(h.DependsOn) > 0 {
						dep = " depends_on=" + strings.Join(h.DependsOn, ",")
					}
					val := firstNonEmpty(h.Value, "")
					if val == "" && len(h.Contents) > 0 {
						val = h.Contents[0]
					}
					lines = append(lines,
						"- hop "+itoa(h.HopIndex)+": "+h.Kind+
							" target="+firstNonEmpty(h.OutputKey, h.Entity)+
							dep+
							" source="+h.Source+
							" result="+truncateRunes(val, 120),
					)
					for i, c := range h.Contents {
						id := ""
						if i < len(h.MemoryIDs) {
							id = h.MemoryIDs[i]
						}
						add(c, id)
					}
				}
			}
		}
	}
	for _, it := range pkt.ProofChain {
		add(it.Content, it.MemoryID)
	}
	// Broad search context after structured proof so slot values are not crowded out.
	for _, it := range pkt.ContextEvidence {
		add(it.Content, it.MemoryID)
	}
	if len(pkt.ContextEvidence) == 0 {
		for _, c := range pkt.Contents {
			add(c, "")
		}
	}
	for _, it := range pkt.Items {
		add(it.Content, it.MemoryID)
	}
	if pkt.TemporalAnswer != "" {
		add(pkt.TemporalAnswer, "")
	}
	return prioritizeHybridMemoryLines(query, hops, lines)
}

// hybridMemoryLineLimit caps the hybrid prompt so activity dumps cannot
// crowd leftover-covering facts or leak chain-of-thought as the answer.
const hybridMemoryLineLimit = 28

// hybridCoverLineLimit keeps leftover-token matches from filling the cap
// with generic "countries he wants to visit" lines and dropping the specific
// place/name fact (United Kingdom, gym, Max).
const hybridCoverLineLimit = 12

func prioritizeHybridMemoryLines(query string, hops []HopResult, lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	coverToks := leftoverNonEntityQueryTokens(query, hops)
	head := make([]string, 0, 8)
	cover := make([]string, 0, len(lines))
	specific := make([]string, 0, len(lines))
	thin := make([]string, 0, len(lines))
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "Structured:" || trim == "Hop chain:" || strings.HasPrefix(trim, "- hop ") {
			head = append(head, line)
			continue
		}
		if len(coverToks) > 0 && contentCoversAnyQueryToken(line, coverToks) {
			cover = append(cover, line)
			continue
		}
		if looksThinPacketLine(line) {
			thin = append(thin, line)
			continue
		}
		specific = append(specific, line)
	}
	cover = rankLinesByRareCover(cover, coverToks)
	if len(cover) > hybridCoverLineLimit {
		cover = cover[:hybridCoverLineLimit]
	}
	out := make([]string, 0, hybridMemoryLineLimit)
	appendCap := func(src []string) {
		for _, line := range src {
			if len(out) >= hybridMemoryLineLimit {
				return
			}
			out = append(out, line)
		}
	}
	appendCap(head)
	appendCap(cover)
	appendCap(specific)
	appendCap(thin)
	return out
}

func rankLinesByRareCover(lines []string, toks []string) []string {
	if len(lines) <= 1 || len(toks) == 0 {
		return lines
	}
	df := make(map[string]int, len(toks))
	for _, tok := range toks {
		for _, line := range lines {
			if contentCoversQueryToken(line, tok) {
				df[tok]++
			}
		}
	}
	type scored struct {
		line string
		n    int
		rare int
	}
	items := make([]scored, 0, len(lines))
	for _, line := range lines {
		n, rare := 0, 1<<20
		for _, tok := range toks {
			if !contentCoversQueryToken(line, tok) {
				continue
			}
			n++
			if d := df[tok]; d < rare {
				rare = d
			}
		}
		if n == 0 {
			rare = 1 << 20
		}
		items = append(items, scored{line: line, n: n, rare: rare})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].n > items[i].n || (items[j].n == items[i].n && items[j].rare < items[i].rare) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.line
	}
	return out
}

func hybridLineBody(line string) string {
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "- [") {
		if i := strings.Index(trim, "] "); i >= 0 {
			return strings.TrimSpace(trim[i+2:])
		}
	}
	return strings.TrimPrefix(trim, "- ")
}

func looksThinPacketLine(line string) bool {
	body := hybridLineBody(line)
	lower := strings.ToLower(body)
	if strings.Contains(lower, "participates in ") || strings.Contains(lower, "unwinds via ") {
		return true
	}
	return len(strings.Fields(body)) <= 4
}

func (s *Service) synthesizeHybridAnswer(ctx context.Context, query string, plan QueryPlan, pkt EvidencePacket) hybridReaderResult {
	cfg := s.resolveHybridReaderConfig()
	if !hybridReaderEnabled() {
		return hybridReaderResult{Reason: "recall_llm_disabled"}
	}
	if !cfg.Configured() {
		return hybridReaderResult{Reason: "reader_unconfigured"}
	}
	if ok, reason := shouldAttemptHybrid(query, plan, pkt); !ok {
		return hybridReaderResult{Reason: reason}
	}

	lines := formatHybridMemoryLinesForQuery(query, pkt)
	if len(lines) == 0 {
		return hybridReaderResult{Attempted: false, Reason: "no_memory_lines"}
	}

	type outPayload struct {
		Answer            string   `json:"answer"`
		SupportingIDs     []string `json:"supporting_memory_ids"`
		UnresolvedTargets []string `json:"unresolved_targets"`
		Abstain           bool     `json:"abstain"`
	}

	// Proven harness semantics (_llm_answer): answer from memories; keep structured
	// fields for observability but treat answer text as primary (soft grounding).
	system := `Answer the question using only the memories below.
Prefer Structured slot values (places, names, lists) over anaphoric phrases like "home country".
If Structured slots do not mention the question's distinctive terms, ignore those slots and answer from the memory lines.
Prefer a concrete short answer (dates, names, places, lists).
If memories place an event on a weekday or weekend relative to a calendar date, keep that relative phrasing. Do not substitute a different calendar day.
When several memories support different parts of the answer, combine every distinct supported item — do not stop at the first memory.
Never invent facts that are not in the memories.
If memories truly lack the answer, set abstain=true and answer to an empty string.
Return JSON: {"answer":"...","supporting_memory_ids":["..."],"unresolved_targets":["..."],"abstain":false}
Include supporting_memory_ids from the bracketed ids when possible; missing ids alone must not block a correct answer.`

	user := "Memories:\n" + strings.Join(lines, "\n") + "\n\nQuestion: " + strings.TrimSpace(query)
	body, err := json.Marshal(map[string]any{
		"model":           cfg.Model,
		"temperature":     0,
		"max_tokens":      2048,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	})
	if err != nil {
		return hybridReaderResult{Attempted: true, Reason: "request_marshal_error"}
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	endpoint := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return hybridReaderResult{Attempted: true, Reason: "request_build_error"}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "brainy-hybrid-reader")
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return hybridReaderResult{Attempted: true, Reason: "llm_http_error"}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return hybridReaderResult{Attempted: true, Reason: "llm_read_error"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return hybridReaderResult{Attempted: true, Reason: "llm_http_status"}
	}
	var completion struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content          string `json:"content"`
				Reasoning        string `json:"reasoning"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &completion); err != nil || len(completion.Choices) == 0 {
		return hybridReaderResult{Attempted: true, Reason: "llm_parse_error"}
	}
	raw := strings.TrimSpace(completion.Choices[0].Message.Content)
	if raw == "" {
		raw = strings.TrimSpace(completion.Choices[0].Message.Reasoning)
	}
	if raw == "" {
		raw = strings.TrimSpace(completion.Choices[0].Message.ReasoningContent)
	}
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	raw = stripThinkTags(raw)
	if obj := extractJSONObject(raw); obj != "" {
		raw = obj
	}

	var out outPayload
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		if ans := softFreeformAnswer(raw); ans != "" {
			return hybridReaderResult{
				Answer:    ans,
				OK:        true,
				Attempted: true,
				Reason:    "freeform_accepted",
				ParseMode: "freeform",
			}
		}
		if ans := salvageJSONAnswer(raw); ans != "" {
			return hybridReaderResult{
				Answer:    ans,
				OK:        true,
				Attempted: true,
				Reason:    "json_salvaged",
				ParseMode: "json_salvage",
			}
		}
		snippet := raw
		if len(snippet) > 160 {
			snippet = snippet[:160]
		}
		return hybridReaderResult{Attempted: true, Reason: "json_parse_error", RawSnippet: snippet}
	}

	ans := strings.TrimSpace(out.Answer)
	if out.Abstain && ans == "" {
		return hybridReaderResult{
			Attempted:         true,
			Reason:            "abstain",
			Abstain:           true,
			SupportingIDs:     out.SupportingIDs,
			UnresolvedTargets: out.UnresolvedTargets,
			ParseMode:         "json",
		}
	}
	if ans == "" || isHybridGarbageAnswer(ans) {
		return hybridReaderResult{
			Attempted:         true,
			Reason:            "empty_or_garbage_answer",
			Abstain:           out.Abstain,
			SupportingIDs:     out.SupportingIDs,
			UnresolvedTargets: out.UnresolvedTargets,
			ParseMode:         "json",
		}
	}
	// Soft grounding: prefer answers with support IDs but never reject solely for missing IDs.
	reason := "ok"
	if len(out.SupportingIDs) == 0 {
		reason = "ok_without_support_ids"
	}
	return hybridReaderResult{
		Answer:            ans,
		OK:                true,
		Attempted:         true,
		Reason:            reason,
		SupportingIDs:     out.SupportingIDs,
		UnresolvedTargets: out.UnresolvedTargets,
		Abstain:           out.Abstain,
		ParseMode:         "json",
	}
}

func softFreeformAnswer(raw string) string {
	ans := strings.TrimSpace(raw)
	if ans == "" || strings.HasPrefix(ans, "{") {
		return ""
	}
	if isHybridGarbageAnswer(ans) {
		return ""
	}
	return ans
}

func stripThinkTags(raw string) string {
	s := raw
	for _, tag := range []string{"think", "analysis", "reasoning"} {
		open := "<" + tag + ">"
		close := "</" + tag + ">"
		for {
			start := strings.Index(strings.ToLower(s), open)
			end := strings.Index(strings.ToLower(s), close)
			if start < 0 || end < 0 || end < start {
				break
			}
			s = strings.TrimSpace(s[:start] + s[end+len(close):])
		}
	}
	return strings.TrimSpace(s)
}

func extractJSONObject(raw string) string {
	start := strings.Index(raw, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inStr {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(raw[start : i+1])
			}
		}
	}
	return ""
}

func salvageJSONAnswer(raw string) string {
	// Best-effort extract of "answer":"..." when full JSON unmarshal fails.
	lower := strings.ToLower(raw)
	idx := strings.Index(lower, `"answer"`)
	if idx < 0 {
		return ""
	}
	rest := raw[idx+len(`"answer"`):]
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, ":") {
		return ""
	}
	rest = strings.TrimSpace(rest[1:])
	if !strings.HasPrefix(rest, `"`) {
		return ""
	}
	rest = rest[1:]
	var b strings.Builder
	escaped := false
	for _, r := range rest {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '"' {
			break
		}
		b.WriteRune(r)
	}
	ans := strings.TrimSpace(b.String())
	if isHybridGarbageAnswer(ans) {
		return ""
	}
	return ans
}

func isHybridGarbageAnswer(ans string) bool {
	lower := strings.ToLower(strings.TrimSpace(ans))
	switch lower {
	case "", "none", "n/a", "na", "null", "undefined", "not mentioned", "not in memory", "i don't know", "i do not know":
		return true
	}
	if strings.Contains(lower, "we need to answer") ||
		strings.Contains(lower, "search memories:") ||
		strings.Contains(lower, "look for mentions") ||
		strings.Count(lower, "[mem_") >= 3 {
		return true
	}
	if utf8Len(ans) > 1200 {
		return true
	}
	letters, total := 0, 0
	counts := map[rune]int{}
	for _, r := range ans {
		if unicode.IsSpace(r) {
			continue
		}
		total++
		counts[r]++
		if unicode.IsLetter(r) {
			letters++
		}
	}
	if total == 0 || letters == 0 {
		return true
	}
	if letters*5 < total {
		return true
	}
	maxc := 0
	for _, c := range counts {
		if c > maxc {
			maxc = c
		}
	}
	return total >= 8 && maxc*2 >= total
}

func packetEvidenceBlob(pkt EvidencePacket) string {
	var b strings.Builder
	for _, c := range pkt.Contents {
		b.WriteString(c)
		b.WriteByte(' ')
	}
	writeItems := func(items []PacketItem) {
		for _, it := range items {
			b.WriteString(it.Content)
			b.WriteByte(' ')
		}
	}
	writeItems(pkt.Items)
	writeItems(pkt.ProofChain)
	writeItems(pkt.ContextEvidence)
	return b.String()
}

// skipUnrelatedHopSlots drops hop-slot dumps from the hybrid prompt when
// distinctive query tokens are in packet memories but not in those slots.
// Skill/possession/preference joins keep Structured. Dual-entity activity
// dumps do not — two names in the query is not a typed join.
func skipUnrelatedHopSlots(query string, hops []HopResult, pkt EvidencePacket) bool {
	if hopDumpsUnproven(hops) {
		return false
	}
	if strings.TrimSpace(query) == "" {
		return false
	}
	if hopsKeepTypedJoin(hops) {
		return false
	}
	slotBlob := strings.Join(hopSlotValues(hops), " ")
	leftover := make([]string, 0, 4)
	for _, tok := range distinctiveQueryTokens(tokenize(query)) {
		if contentCoversQueryToken(slotBlob, tok) {
			continue
		}
		leftover = append(leftover, tok)
	}
	if len(leftover) == 0 {
		return false
	}
	memBlob := packetEvidenceBlob(pkt)
	for _, tok := range leftover {
		if contentCoversQueryToken(memBlob, tok) {
			return true
		}
	}
	return false
}

func hopsKeepTypedJoin(hops []HopResult) bool {
	for _, h := range hops {
		if !hopResultTypedExact(h) {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(h.Predicate)) {
		case PredicateSkill, PredicatePossession, PredicatePreference:
			return true
		}
	}
	return false
}

func looksCrowdedHopDump(content string) bool {
	if looksTitleCaseSlogan(content) {
		return true
	}
	n := 0
	for _, part := range strings.Split(content, ",") {
		if strings.TrimSpace(part) != "" {
			n++
		}
	}
	return n >= 3
}

// typedAnswerIsHopDump is true for slogan/activity dumps, false for short
// typed joins (clarinet, violin / Oliver, Luna, Bailey).
func typedAnswerIsHopDump(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "not in memory") {
		return false
	}
	parts := make([]string, 0, 8)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	n := len(parts)
	if n <= 2 {
		return looksTitleCaseSlogan(s)
	}
	long, slogans := 0, 0
	for _, p := range parts {
		if utf8Len(p) > 32 {
			long++
		}
		if looksTitleCaseSlogan(p) {
			slogans++
		}
	}
	if n <= 6 && long == 0 && slogans == 0 {
		return false
	}
	if n >= 8 || slogans >= 2 || long >= 2 {
		return true
	}
	if utf8Len(s) > 90 && n >= 4 {
		return true
	}
	return looksTitleCaseSlogan(s)
}

func hopsAreIdentityOnly(hops []HopResult) bool {
	typed, identity := 0, 0
	for _, h := range hops {
		switch h.Kind {
		case "follow_relation", "fetch_predicate", "answer_slot":
			if h.Source == "unresolved" || h.Source == "search_fallback" {
				continue
			}
			typed++
			if strings.EqualFold(strings.TrimSpace(h.Predicate), PredicateIdentity) {
				identity++
			}
		}
	}
	return typed > 0 && identity == typed
}

func leftoverNonEntityQueryTokens(query string, hops []HopResult) []string {
	ents := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		ents[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	slotBlob := strings.Join(hopSlotValues(hops), " ")
	out := make([]string, 0, 4)
	for _, tok := range distinctiveQueryTokens(tokenize(query)) {
		if _, ok := ents[tok]; ok {
			continue
		}
		if contentCoversQueryToken(slotBlob, tok) {
			continue
		}
		out = append(out, tok)
	}
	return out
}

func contentCoversAnyQueryToken(content string, toks []string) bool {
	for _, tok := range toks {
		if contentCoversQueryToken(content, tok) {
			return true
		}
	}
	return false
}
