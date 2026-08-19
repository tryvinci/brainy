package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// ProviderConfig configures an OpenAI-compatible /v1/embeddings client.
type ProviderConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	Dimensions int
	Strict     bool
}

func (c ProviderConfig) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

// ProviderEmbedder calls an OpenAI-compatible embeddings endpoint.
// Product default soft-degrades to LocalEmbedder on failure. Strict mode
// returns the provider error instead of substituting the 128-d hash.
type ProviderEmbedder struct {
	client      *http.Client
	cfg         ProviderConfig
	fallback    LocalEmbedder
	calls       atomic.Int64
	failures    atomic.Int64
	fallbacks   atomic.Int64
	observedDim atomic.Int64
}

func NewProviderEmbedder(cfg ProviderConfig, client *http.Client) *ProviderEmbedder {
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	return &ProviderEmbedder{
		client:   client,
		cfg:      cfg,
		fallback: NewLocalEmbedder(),
	}
}

func (p *ProviderEmbedder) WithStrict(strict bool) *ProviderEmbedder {
	p.cfg.Strict = strict
	return p
}

func (p *ProviderEmbedder) Name() string {
	if !p.cfg.Configured() {
		return p.fallback.Name()
	}
	return "provider-embeddings:" + p.cfg.Model
}

func (p *ProviderEmbedder) Identity() Identity {
	if !p.cfg.Configured() {
		id := p.fallback.Identity()
		id.Strict = p.cfg.Strict
		return id
	}
	dim := p.cfg.Dimensions
	if dim <= 0 {
		if obs := int(p.observedDim.Load()); obs > 0 {
			dim = obs
		} else {
			dim = ProviderDim
		}
	}
	return Identity{
		Name:       p.Name(),
		Provider:   "openai-compatible",
		Model:      strings.TrimSpace(p.cfg.Model),
		Dimensions: dim,
		Version:    strings.TrimSpace(p.cfg.Model) + "@" + itoa(dim),
		Strict:     p.cfg.Strict,
	}
}

func (p *ProviderEmbedder) Stats() Stats {
	return Stats{
		Calls:     p.calls.Load(),
		Failures:  p.failures.Load(),
		Fallbacks: p.fallbacks.Load(),
	}
}

func (p *ProviderEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	p.calls.Add(1)
	if !p.cfg.Configured() {
		if p.cfg.Strict {
			p.failures.Add(1)
			return nil, fmt.Errorf("embedding provider required in strict mode")
		}
		p.fallbacks.Add(1)
		return p.fallback.Embed(ctx, text)
	}
	values, err := p.embedProvider(ctx, text)
	if err != nil {
		p.failures.Add(1)
		if p.cfg.Strict {
			return nil, err
		}
		p.fallbacks.Add(1)
		return p.fallback.Embed(ctx, text)
	}
	if len(values) == 0 {
		p.failures.Add(1)
		err := fmt.Errorf("embedding empty response")
		if p.cfg.Strict {
			return nil, err
		}
		p.fallbacks.Add(1)
		return p.fallback.Embed(ctx, text)
	}
	p.observedDim.Store(int64(len(values)))
	return values, nil
}

func (p *ProviderEmbedder) embedProvider(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]any{
		"model": p.cfg.Model,
		"input": text,
	}
	if p.cfg.Dimensions > 0 && SupportsDimensions(p.cfg.Model) {
		payload["dimensions"] = p.cfg.Dimensions
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(p.cfg.BaseURL, "/") + "/embeddings"
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
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("embedding read: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	var decoded struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return nil, fmt.Errorf("embedding decode: %w", err)
	}
	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding empty response")
	}
	out := make([]float32, len(decoded.Data[0].Embedding))
	for i, v := range decoded.Data[0].Embedding {
		out[i] = float32(v)
	}
	return out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
