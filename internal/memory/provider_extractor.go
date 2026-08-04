package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const providerExtractionVersion = "provider-v2-additive"

// ProviderConfig configures an OpenAI-compatible chat completions client.
type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func (c ProviderConfig) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

// ProviderExtractor calls an OpenAI-compatible /v1/chat/completions endpoint
// and validates structured JSON memories. It never mutates raw ingest state.
type ProviderExtractor struct {
	client   *http.Client
	cfg      ProviderConfig
	fallback DeterministicExtractor
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

func (p *ProviderExtractor) Extract(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	// Deterministic baseline always runs first (ENG-92).
	baseline, err := p.fallback.Extract(ctx, req)
	if err != nil {
		return nil, err
	}
	if !p.cfg.Configured() {
		return baseline, nil
	}

	providerMemories, err := p.extractProvider(ctx, req)
	if err != nil {
		if len(baseline) > 0 {
			return baseline, nil
		}
		return nil, err
	}
	return mergeProviderAndBaseline(baseline, providerMemories), nil
}

func (p *ProviderExtractor) extractProvider(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	body, err := json.Marshal(map[string]any{
		"model":       p.cfg.Model,
		"temperature": 0,
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

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, fmt.Errorf("provider extract decode: %w", err)
	}
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		return nil, nil
	}

	return parseProviderMemories(completion.Choices[0].Message.Content)
}

// Additive extraction: one standalone atom per fact, attributed to named speakers.
// Examples are synthetic (master-plan anti-MemPalace clause).
const providerSystemPrompt = `You are a Memory Analyzer extracting ADD-only atomic memories from conversation.

Return JSON only:
{"memories":[{"kind":"fact|preference|profile","content":"...","source_text":"...","confidence":0.0,"when":"optional absolute date","duration":"optional"}]}

CRITICAL RULES:
1. content must be a self-contained factual sentence usable months later WITHOUT the original turn.
2. In multi-speaker dialogue ("Name: ..."), attribute facts to that speaker by name
   (e.g. "Jordan is a nurse", "Sam participates in ceramics").
3. Emit ONE memory per distinct attribute, activity, place, titled work, preference, or plan.
   Split compound utterances. Prefer many small atoms over one long paraphrase.
4. Always extract when present:
   - identity / role / relationship status
   - origin / moved from <place> (include country/city names literally)
   - activities and hobbies
   - places tied to activities
   - book/movie titles in quotes (verbatim spans)
   - family members' preferences (e.g. "Sam's kids like astronomy")
   - career plans and research topics
5. Resolve relative time against Observation Date in the user message (yesterday → absolute date).
6. Skip pure greetings/acks ("Thanks!", "Yeah, Name", "Cool").
7. When in doubt, EXTRACT — missed atoms destroy multi-attribute recall.
8. return {"memories":[]} only if nothing durable exists.`

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
	Kind       string  `json:"kind"`
	Content    string  `json:"content"`
	SourceText string  `json:"source_text"`
	Confidence float64 `json:"confidence"`
	When       string  `json:"when"`
	Duration   string  `json:"duration"`
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
		kind := strings.ToLower(strings.TrimSpace(item.Kind))
		switch kind {
		case KindFact, KindPreference, KindProfile:
		default:
			return nil, fmt.Errorf("provider extract: invalid kind %q", item.Kind)
		}
		content := NormalizeText(item.Content)
		if content == "" {
			return nil, fmt.Errorf("provider extract: empty content")
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
			"rule": "provider_extract",
		}
		if when := strings.TrimSpace(item.When); when != "" {
			explain["when"] = when
			if kind == KindFact {
				explain["primitive"] = PrimitiveEpisode
			}
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

func mergeProviderAndBaseline(baseline, provider []ExtractedMemory) []ExtractedMemory {
	if len(provider) == 0 {
		return baseline
	}
	out := make([]ExtractedMemory, 0, len(provider)+len(baseline))
	seen := make(map[string]struct{}, len(provider)+len(baseline))
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
	for _, item := range baseline {
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
