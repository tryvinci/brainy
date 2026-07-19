package embedding

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

// ProviderConfig configures an OpenAI-compatible /v1/embeddings client.
type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func (c ProviderConfig) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

// ProviderEmbedder calls an OpenAI-compatible embeddings endpoint.
// On failure it soft-degrades to LocalEmbedder so ingest/search keep working.
type ProviderEmbedder struct {
	client   *http.Client
	cfg      ProviderConfig
	fallback LocalEmbedder
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

func (p *ProviderEmbedder) Name() string {
	if !p.cfg.Configured() {
		return p.fallback.Name()
	}
	return "provider-embeddings:" + p.cfg.Model
}

func (p *ProviderEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if !p.cfg.Configured() {
		return p.fallback.Embed(ctx, text)
	}
	values, err := p.embedProvider(ctx, text)
	if err != nil {
		// Soft-degrade: never block ingest/search on embedding provider flakes.
		return p.fallback.Embed(ctx, text)
	}
	if len(values) == 0 {
		return p.fallback.Embed(ctx, text)
	}
	return values, nil
}

func (p *ProviderEmbedder) embedProvider(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(map[string]any{
		"model": p.cfg.Model,
		"input": text,
	})
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
	var payload struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("embedding decode: %w", err)
	}
	if len(payload.Data) == 0 || len(payload.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embedding empty response")
	}
	out := make([]float32, len(payload.Data[0].Embedding))
	for i, v := range payload.Data[0].Embedding {
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
