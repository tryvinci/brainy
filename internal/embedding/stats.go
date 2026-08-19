package embedding

import "strings"

// Stats are process-local provider counters. They never include secrets.
type Stats struct {
	Calls     int64 `json:"calls"`
	Failures  int64 `json:"failures"`
	Fallbacks int64 `json:"fallbacks"`
}

// Identity is the non-secret embedder signature used in runtime manifests.
type Identity struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
	Version    string `json:"version"`
	Strict     bool   `json:"strict"`
}

func (i Identity) Signature() string {
	return strings.TrimSpace(i.Provider) + "|" + strings.TrimSpace(i.Model) + "|" + itoa(i.Dimensions)
}

// Record is one embedding write, including provenance for ANN safety.
type Record struct {
	MemoryID   string
	TenantID   string
	SubjectID  string
	Values     []float32
	Provider   string
	Model      string
	Version    string
	Dimensions int
}

func RecordFromEmbedder(e Embedder, memoryID, tenantID, subjectID string, values []float32) Record {
	rec := Record{
		MemoryID:   memoryID,
		TenantID:   tenantID,
		SubjectID:  subjectID,
		Values:     values,
		Dimensions: len(values),
	}
	if ident, ok := e.(interface{ Identity() Identity }); ok {
		id := ident.Identity()
		rec.Provider = id.Provider
		rec.Model = id.Model
		rec.Version = id.Version
		if id.Dimensions > 0 {
			rec.Dimensions = id.Dimensions
		}
	} else if e != nil {
		rec.Provider = e.Name()
		rec.Model = e.Name()
	}
	if rec.Version == "" {
		rec.Version = rec.Provider + "@" + itoa(rec.Dimensions)
	}
	return rec
}

func IdentityOf(e Embedder) Identity {
	if e == nil {
		return Identity{Name: "none", Provider: "none"}
	}
	if ident, ok := e.(interface{ Identity() Identity }); ok {
		return ident.Identity()
	}
	return Identity{Name: e.Name(), Provider: e.Name(), Model: e.Name()}
}

func StatsOf(e Embedder) Stats {
	if e == nil {
		return Stats{}
	}
	if s, ok := e.(interface{ Stats() Stats }); ok {
		return s.Stats()
	}
	return Stats{}
}

// SupportsDimensions is true for OpenAI text-embedding-3-* (and compatible
// proxies that honor the same field). Cloudflare BGE and hash embedders do not.
func SupportsDimensions(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(m, "text-embedding-3")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
