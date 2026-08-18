package memory

import (
	"strings"
	"testing"
)

func TestCanonicalEntityIDScopesAndQualifiers(t *testing.T) {
	johnSmith := CanonicalEntityID("t1", "u1", "John Smith")
	johnDoe := CanonicalEntityID("t1", "u1", "John Doe")
	johnSmithOtherSubject := CanonicalEntityID("t1", "u2", "John Smith")
	if johnSmith == "" || !strings.HasPrefix(johnSmith, "ent:") {
		t.Fatalf("id=%q", johnSmith)
	}
	if johnSmith == johnDoe {
		t.Fatal("two Johns with different labels must not share an ID")
	}
	if johnSmith == johnSmithOtherSubject {
		t.Fatal("same name in another subject must not share an ID")
	}
	if CanonicalEntityID("t1", "u1", "john smith") != johnSmith {
		t.Fatal("normalization must be stable")
	}
	if CanonicalEntityID("t1", "u1", johnSmith) != johnSmith {
		t.Fatal("already-canonical IDs must pass through")
	}
}

func TestRankEntityResolutionUniqueFirstName(t *testing.T) {
	smith := MemoryEntity{EntityID: "ent:smith", CanonicalLabel: "John Smith", Aliases: EntityAliases("John Smith")}
	doe := MemoryEntity{EntityID: "ent:doe", CanonicalLabel: "John Doe", Aliases: EntityAliases("John Doe")}
	got, ok := RankEntityResolution([]MemoryEntity{smith}, "John")
	if !ok || got.EntityID != "ent:smith" {
		t.Fatalf("unique first name, got %+v ok=%v", got, ok)
	}
	if _, ok := RankEntityResolution([]MemoryEntity{smith, doe}, "John"); ok {
		t.Fatal("two Johns must not collapse on a first-name mention")
	}
	got, ok = RankEntityResolution([]MemoryEntity{smith, doe}, "John Doe")
	if !ok || got.EntityID != "ent:doe" {
		t.Fatalf("full label, got %+v ok=%v", got, ok)
	}
}

func TestEntityAliasesIncludeFirstToken(t *testing.T) {
	got := EntityAliases("John Smith")
	if len(got) < 2 || got[0] != "john smith" || got[1] != "john" {
		t.Fatalf("aliases=%v", got)
	}
}
