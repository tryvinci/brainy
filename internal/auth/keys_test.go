package auth

import "testing"

func TestParseKeyRing(t *testing.T) {
	ring := ParseKeyRing("demo:sk_demo, acme:sk_acme")
	if !ring.Enabled() {
		t.Fatal("expected key ring enabled")
	}
	if tenant, ok := ring.TenantForKey("sk_demo"); !ok || tenant != "demo" {
		t.Fatalf("demo key: tenant=%q ok=%v", tenant, ok)
	}
	if tenant, ok := ring.TenantForKey("sk_acme"); !ok || tenant != "acme" {
		t.Fatalf("acme key: tenant=%q ok=%v", tenant, ok)
	}
	if _, ok := ring.TenantForKey("missing"); ok {
		t.Fatal("expected missing key to fail")
	}
}
