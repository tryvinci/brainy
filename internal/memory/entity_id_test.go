package memory

import (
	"context"
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

func TestExtractDialogueAliases(t *testing.T) {
	pairs := ExtractDialogueAliases("call me Al", "Alex")
	if len(pairs) != 1 || !strings.EqualFold(pairs[0][0], "Alex") || !strings.EqualFold(pairs[0][1], "Al") {
		t.Fatalf("call-me pairs=%v", pairs)
	}
	pairs = ExtractDialogueAliases("Jordan is also known as Jo", "")
	if len(pairs) != 1 || !strings.EqualFold(pairs[0][1], "Jo") {
		t.Fatalf("aka pairs=%v", pairs)
	}
}

func TestDialogueAliasResolvesNickname(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-alias", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex: friends call me Al when we hike."},
			{Role: "user", Content: "Alex: I enjoy hiking every weekend."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ent := ResolveCanonicalEntity(context.Background(), store, "t-alias", "u1", "Al")
	if ent.EntityID == "" {
		t.Fatal("expected nickname to resolve")
	}
	alex := ResolveCanonicalEntity(context.Background(), store, "t-alias", "u1", "Alex")
	if alex.EntityID == "" || alex.EntityID != ent.EntityID {
		t.Fatalf("Al and Alex must share an ID, al=%+v alex=%+v", ent, alex)
	}
}
