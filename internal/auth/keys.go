package auth

import (
	"strings"
)

// KeyRing maps API key secrets to tenant IDs.
type KeyRing struct {
	byKey map[string]string
}

func ParseKeyRing(raw string) *KeyRing {
	ring := &KeyRing{byKey: map[string]string{}}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		tenantID, key, ok := strings.Cut(entry, ":")
		if !ok || tenantID == "" || key == "" {
			continue
		}
		ring.byKey[key] = tenantID
	}
	return ring
}

func (r *KeyRing) TenantForKey(key string) (string, bool) {
	if r == nil || key == "" {
		return "", false
	}
	tenantID, ok := r.byKey[key]
	return tenantID, ok
}

func (r *KeyRing) Enabled() bool {
	return r != nil && len(r.byKey) > 0
}
