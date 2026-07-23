package config

import (
	"os"
	"testing"
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
