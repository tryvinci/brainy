package config

import (
	"os"
	"testing"
	"time"
)

func TestEntityRankingDefault(t *testing.T) {
	os.Unsetenv("BRAINY_ENTITY_RANKING")
	os.Setenv("BRAINY_EMBEDDING_MODEL", "some-embed-model")
	if entityRankingDefault() {
		t.Fatal("expected OFF by default even with embedding model set")
	}
	os.Setenv("BRAINY_ENTITY_RANKING", "true")
	if !entityRankingDefault() {
		t.Fatal("explicit true must enable")
	}
	os.Setenv("BRAINY_ENTITY_RANKING", "false")
	if entityRankingDefault() {
		t.Fatal("explicit false must stay off")
	}
	os.Unsetenv("BRAINY_ENTITY_RANKING")
	os.Unsetenv("BRAINY_EMBEDDING_MODEL")
}

func TestHTTPHardeningDefaults(t *testing.T) {
	os.Unsetenv("BRAINY_MAX_BODY_BYTES")
	os.Unsetenv("BRAINY_HTTP_READ_HEADER_TIMEOUT")
	os.Unsetenv("BRAINY_HTTP_READ_TIMEOUT")
	os.Unsetenv("BRAINY_HTTP_WRITE_TIMEOUT")
	os.Unsetenv("BRAINY_HTTP_IDLE_TIMEOUT")
	os.Unsetenv("BRAINY_PROVIDER_TIMEOUT")

	cfg := Load()
	if cfg.MaxBodyBytes != 5<<20 {
		t.Fatalf("expected default max body 5MiB, got %d", cfg.MaxBodyBytes)
	}
	if cfg.HTTPReadHeaderTimeout != 10*time.Second {
		t.Fatalf("expected default read-header timeout 10s, got %s", cfg.HTTPReadHeaderTimeout)
	}
	if cfg.HTTPReadTimeout != 30*time.Second {
		t.Fatalf("expected default read timeout 30s, got %s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPWriteTimeout != 105*time.Second {
		t.Fatalf("expected default write timeout provider+60s (105s), got %s", cfg.HTTPWriteTimeout)
	}
	if cfg.HTTPIdleTimeout != 120*time.Second {
		t.Fatalf("expected default idle timeout 120s, got %s", cfg.HTTPIdleTimeout)
	}
}

func TestHTTPHardeningOverrides(t *testing.T) {
	os.Setenv("BRAINY_MAX_BODY_BYTES", "1048576")
	os.Setenv("BRAINY_HTTP_READ_HEADER_TIMEOUT", "5s")
	os.Setenv("BRAINY_HTTP_READ_TIMEOUT", "15s")
	os.Setenv("BRAINY_HTTP_WRITE_TIMEOUT", "45s")
	os.Setenv("BRAINY_HTTP_IDLE_TIMEOUT", "60s")
	t.Cleanup(func() {
		os.Unsetenv("BRAINY_MAX_BODY_BYTES")
		os.Unsetenv("BRAINY_HTTP_READ_HEADER_TIMEOUT")
		os.Unsetenv("BRAINY_HTTP_READ_TIMEOUT")
		os.Unsetenv("BRAINY_HTTP_WRITE_TIMEOUT")
		os.Unsetenv("BRAINY_HTTP_IDLE_TIMEOUT")
	})

	cfg := Load()
	if cfg.MaxBodyBytes != 1048576 {
		t.Fatalf("expected max body 1MiB, got %d", cfg.MaxBodyBytes)
	}
	if cfg.HTTPReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read-header timeout 5s, got %s", cfg.HTTPReadHeaderTimeout)
	}
	if cfg.HTTPReadTimeout != 15*time.Second {
		t.Fatalf("expected read timeout 15s, got %s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPWriteTimeout != 45*time.Second {
		t.Fatalf("expected write timeout 45s, got %s", cfg.HTTPWriteTimeout)
	}
	if cfg.HTTPIdleTimeout != 60*time.Second {
		t.Fatalf("expected idle timeout 60s, got %s", cfg.HTTPIdleTimeout)
	}
}
