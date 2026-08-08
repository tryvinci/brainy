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
)

// PacketItem is one structured evidence unit for hybrid reading.
type PacketItem struct {
	EvidenceID string   `json:"evidence_id,omitempty"`
	MemoryID   string   `json:"memory_id,omitempty"`
	Content    string   `json:"content"`
	Predicate  string   `json:"predicate,omitempty"`
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

func (s *Service) synthesizeHybridAnswer(ctx context.Context, query string, plan QueryPlan, pkt EvidencePacket) (answer string, ok bool) {
	cfg := s.hybridReader
	if !cfg.Configured() {
		// Fall back to provider env if recall LLM not wired.
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
	}
	if !cfg.Configured() || !hybridReaderEnabled() {
		return "", false
	}
	// Deterministic paths win for simple temporal / scalar when already resolved.
	if plan.NeedsTemporal && pkt.TemporalAnswer != "" && !plan.NeedsMultiHop {
		return "", false
	}
	if !(plan.NeedsMultiHop || plan.NeedsEnumeration || plan.PrimaryIntent == IntentMultiHop) {
		// Still allow hybrid when deterministic packet is weak.
		if packetCoverageSatisfied(plan, pkt) && len(pkt.Contents) <= 2 {
			return "", false
		}
	}

	type outPayload struct {
		Answer            string   `json:"answer"`
		SupportingIDs     []string `json:"supporting_memory_ids"`
		UnresolvedTargets []string `json:"unresolved_targets"`
		Abstain           bool     `json:"abstain"`
	}
	system := `You answer using ONLY the provided evidence packet JSON.
Return JSON: {"answer":"...","supporting_memory_ids":["..."],"unresolved_targets":["..."],"abstain":false}
Rules: cite only packet memory_ids; abstain=true if required targets are missing; do not invent facts; be concise.`
	packetJSON, _ := json.Marshal(map[string]any{
		"query":            query,
		"plan":             plan,
		"temporal_answer":  pkt.TemporalAnswer,
		"items":            pkt.Items,
		"contents":         pkt.Contents,
		"coverage_targets": plan.CoverageTargets,
	})
	body, err := json.Marshal(map[string]any{
		"model":       cfg.Model,
		"temperature": 0,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": "Evidence packet:\n" + string(packetJSON)},
		},
	})
	if err != nil {
		return "", false
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	endpoint := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", false
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		httpReq.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false
	}
	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &completion); err != nil || len(completion.Choices) == 0 {
		return "", false
	}
	raw := strings.TrimSpace(completion.Choices[0].Message.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	var out outPayload
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", false
	}
	if out.Abstain || strings.TrimSpace(out.Answer) == "" {
		return "", false
	}
	return strings.TrimSpace(out.Answer), true
}