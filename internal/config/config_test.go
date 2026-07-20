package config

import (
	"os"
	"testing"
)

func TestEntityRankingDefault(t *testing.T) {
	os.Unsetenv("BRAINY_ENTITY_RANKING")
	os.Unsetenv("BRAINY_EMBEDDING_MODEL")
	os.Unsetenv("EMBEDDING_MODEL")
	if entityRankingDefault() {
		t.Fatal("expected OFF with no embedder configured")
	}
	os.Setenv("BRAINY_EMBEDDING_MODEL", "some-embed-model")
	if !entityRankingDefault() {
		t.Fatal("expected ON when embedding model configured")
	}
	os.Setenv("BRAINY_ENTITY_RANKING", "false")
	if entityRankingDefault() {
		t.Fatal("explicit false must override")
	}
	os.Unsetenv("BRAINY_ENTITY_RANKING")
	os.Unsetenv("BRAINY_EMBEDDING_MODEL")
}
