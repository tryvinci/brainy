package pack

import (
	"path/filepath"
	"testing"
)

func TestLoadMarketingPack(t *testing.T) {
	root := filepath.Join("..", "..", "packs")
	reg, err := LoadRegistryFromDir(root)
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	p, ok := reg.Get("marketing")
	if !ok {
		t.Fatal("marketing pack not registered")
	}
	label, primitive, ok := p.LabelForKind("preference")
	if !ok || label != "voice_profile" || primitive != "identity_prior" {
		t.Fatalf("LabelForKind(preference) = %q, %q, %v", label, primitive, ok)
	}
	if w := p.PrimitiveWeight("principle"); w <= 0 {
		t.Fatalf("expected principle weight, got %v", w)
	}
}
