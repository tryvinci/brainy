package memory

import (
	"context"
	"strings"
	"testing"
)

func TestHopJoinProvenRequiresTypedIDJoin(t *testing.T) {
	alex := CanonicalEntityID("t1", "u1", "Alex")
	dana := CanonicalEntityID("t1", "u1", "Dana")
	hops := []HopResult{
		{Kind: "resolve_entity", OutputKey: "e1", Value: "Alex", EntityID: alex, Source: "typed_store", ProofKind: "typed_exact"},
		{Kind: "fetch_predicate", OutputKey: "ans", Entity: "Alex", EntityID: alex, Predicate: PredicateOccupation, Value: "nurse", Source: "typed_store", ProofKind: "typed_exact", DependsOn: []string{"e1"}},
	}
	if !hopJoinProven(hops) {
		t.Fatal("matching entity IDs must prove the join")
	}
	hops[1].EntityID = dana
	if hopJoinProven(hops) {
		t.Fatal("mismatched entity IDs must not prove the join")
	}
}

func TestComposeFromHopValuesIntersectsEntities(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Nate", Value: "turtles", Values: []string{"turtles", "nintendo games"}, Source: "typed_store"},
		{Kind: "fetch_predicate", Entity: "Joanna", Value: "turtles", Values: []string{"turtles", "writing"}, Source: "typed_store"},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "turtle") {
		t.Fatalf("expected shared turtles, got %q", ans)
	}
	if strings.Contains(low, "nintendo") || strings.Contains(low, "writing") {
		t.Fatalf("union leaked into join answer: %q", ans)
	}
}

func TestComposeFromHopValuesCapsSingleEntityDump(t *testing.T) {
	vals := make([]string, 0, 12)
	for i := 0; i < 12; i++ {
		vals = append(vals, "item"+strings.Repeat("x", i+1))
	}
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Riley", Values: vals, Source: "typed_store"},
	})
	if strings.Count(ans, ",") > 5 {
		t.Fatalf("single-entity hop compose must bound the dump, got %q", ans)
	}
}

func TestComposeFromHopValuesJoinDoesNotFallBackToUnion(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Nate", Value: "turtles", Values: []string{"turtles"}, Source: "typed_store"},
		{Kind: "fetch_predicate", Entity: "Joanna", Value: "writing", Values: []string{"writing"}, Source: "typed_store"},
	})
	if strings.TrimSpace(ans) != "" {
		t.Fatalf("disjoint join must not dump the union, got %q", ans)
	}
}

func TestIntersectHopValuesByRareSharedToken(t *testing.T) {
	hops := []HopResult{
		{Kind: "fetch_predicate", Entity: "Tim", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"map of middle-earth", "dvd and game collection", "basketball signed by favorite player", "harry potter books"}},
		{Kind: "fetch_predicate", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"lord of the rings dvd collection", "star wars movies", "basketball trophy", "hiking gear"}},
	}
	got := intersectHopValuesByRareSharedToken(hops)
	blob := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(blob, "basketball") {
		t.Fatalf("expected basketball join, got %#v", got)
	}
	if strings.Contains(blob, "collection") {
		t.Fatalf("high-df collection must not win, got %#v", got)
	}
	ans := composeFromHopValues(hops)
	if !strings.Contains(strings.ToLower(ans), "basketball") {
		t.Fatalf("compose should use rare shared basketball, got %q", ans)
	}
}

func TestIntersectHopValuesByRareSharedTokenFromContents(t *testing.T) {
	hops := []HopResult{
		{Kind: "fetch_predicate", Entity: "Tim", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"map of middle-earth", "dvd and game collection", "harry potter books"},
			Contents: []string{
				"Tim owns a prized basketball signed by his favorite NBA player.",
				"Tim owns a dvd and game collection.",
			}},
		{Kind: "fetch_predicate", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"lord of the rings dvd collection", "star wars movies", "hiking gear"},
			Contents: []string{
				"John possesses a basketball that has been signed by his teammates.",
				"John owns lord of the rings dvd collection.",
			}},
	}
	got := intersectHopValuesByRareSharedToken(hops)
	blob := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(blob, "basketball") {
		t.Fatalf("expected basketball join from omitted contents, got %#v", got)
	}
	if strings.Contains(blob, "collection") {
		t.Fatalf("value-covered collection must not win, got %#v", got)
	}
	ans := composeFromHopValues(hops)
	if !strings.Contains(strings.ToLower(ans), "basketball") {
		t.Fatalf("compose should rare-share omitted basketball contents, got %q", ans)
	}
}

func TestIntersectHopValuesByRareSharedTokenIgnoresParaphrasedValues(t *testing.T) {
	hops := []HopResult{
		{Kind: "fetch_predicate", Entity: "Tim", Predicate: PredicatePossession, Source: "typed_store",
			Values:   []string{"dvd and game collection", "basketball signed by favorite player"},
			Contents: []string{"Tim owns a prized basketball signed by his favorite player, kept in a case."}},
		{Kind: "fetch_predicate", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Values:   []string{"lord of the rings dvd collection", "basketball trophy", "photo of basketball court at sunset"},
			Contents: []string{"John took a photo of a basketball court at sunset during a morning workout."}},
	}
	got := intersectHopValuesByRareSharedToken(hops)
	blob := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(blob, "basketball") {
		t.Fatalf("expected basketball join, got %#v", got)
	}
	if strings.Contains(blob, "sunset") || strings.Contains(blob, "workout") || strings.Contains(blob, "photo") {
		t.Fatalf("paraphrased court photo must not rare-share, got %#v", got)
	}
	ans := composeFromHopValues(hops)
	low := strings.ToLower(ans)
	if !strings.Contains(low, "basketball") {
		t.Fatalf("compose should join basketball, got %q", ans)
	}
	if strings.Contains(low, "sunset") || strings.Contains(low, "workout") {
		t.Fatalf("compose leaked court-photo paraphrase, got %q", ans)
	}
}

func TestHopComposeUsableKeepsTitleCasedBasketballJoin(t *testing.T) {
	hops := []HopResult{
		{Kind: "fetch_predicate", Entity: "Tim", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"map of middle-earth", "dvd and game collection", "basketball signed by favorite player"}},
		{Kind: "fetch_predicate", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Values: []string{"lord of the rings dvd collection", "basketball trophy", "hiking gear"}},
	}
	ans := composeFromHopValues(hops)
	if !strings.Contains(strings.ToLower(ans), "basketball") {
		t.Fatalf("expected basketball join, got %q", ans)
	}
	if !hopComposeUsable(ans, hops) {
		t.Fatalf("title-cased possession join must stay usable, dump=%v ans=%q", typedAnswerIsHopDump(ans), ans)
	}
	activity := []HopResult{
		{Kind: "follow_relation", Entity: "Andrew", Predicate: PredicateActivity, Source: "typed_store",
			Values: []string{"bike ride with girlfriend", "trying sushi"}},
		{Kind: "follow_relation", Entity: "Buddy", Predicate: PredicateActivity, Source: "search_fallback",
			Values: []string{"bike ride with girlfriend"}},
	}
	if hopComposeUsable("Way, Road Trip, McGee's Bar, Playing Cyberpunk 2077, Notebook, First, Simple Dishes, Tried Cyberpunk 2077", activity) {
		t.Fatal("activity dumps must not become usable hop compose")
	}
	identity := []HopResult{
		{Kind: "follow_relation", Entity: "Maria", Predicate: PredicateIdentity, Source: "typed_store",
			Values: []string{"inspiration", "family", "team"}},
	}
	if hopComposeUsable("Inspiration, Family, Team", identity) {
		t.Fatal("identity dumps must not become usable hop compose")
	}
}

func TestComposeFromHopValuesDoesNotRareJoinActivityDumps(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "follow_relation", Entity: "Andrew", Predicate: PredicateActivity, Source: "typed_store",
			Values: []string{"bike ride with girlfriend", "trying sushi", "walks with buddy"}},
		{Kind: "follow_relation", Entity: "Buddy", Predicate: PredicateActivity, Source: "search_fallback",
			Values: []string{"bike ride with girlfriend", "discover new places to eat"}},
	})
	low := strings.ToLower(ans)
	if strings.Contains(low, "bike") || strings.Contains(low, "sushi") {
		t.Fatalf("activity dump must not rare-join, got %q", ans)
	}
}

func TestComposeFromHopValuesContainmentAndPartner(t *testing.T) {
	got := intersectHopValuesByContainment([]HopResult{
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity, Values: []string{"yoga", "relaxing", "pottery"}, Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Casey", Predicate: PredicateActivity, Values: []string{"organized yoga", "riley's running group", "ceramics"}, Source: "search_fallback"},
	})
	blob := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(blob, "yoga") {
		t.Fatalf("expected yoga containment join across fallback, got %#v", got)
	}
	if strings.Contains(blob, "relax") || strings.Contains(blob, "ceram") || strings.Contains(blob, "potter") {
		t.Fatalf("private or unwind value leaked into containment join: %#v", got)
	}
}

func TestHopValuesMentioningPartner(t *testing.T) {
	got := hopValuesMentioningPartner([]HopResult{
		{Kind: "follow_relation", Entity: "Riley", Values: []string{"yoga"}, Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Casey", Values: []string{"riley's running group", "ceramics"}, Source: "search_fallback"},
	})
	blob := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(blob, "run") {
		t.Fatalf("expected partner running group, got %#v", got)
	}
	if strings.Contains(blob, "ceram") {
		t.Fatalf("private activity leaked into partner mention: %#v", got)
	}
}

func TestComposeFromHopValuesDoesNotKeepPartnerPreferences(t *testing.T) {
	ans := composeFromHopValues([]HopResult{
		{Kind: "fetch_predicate", Entity: "Nate", Values: []string{"turtles", "nintendo games"}, Source: "typed_store"},
		{Kind: "fetch_predicate", Entity: "Joanna", Values: []string{"turtles", "likes nate's cooking"}, Source: "typed_store"},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "turtle") {
		t.Fatalf("expected shared turtles, got %q", ans)
	}
	if strings.Contains(low, "cook") || strings.Contains(low, "nintendo") {
		t.Fatalf("partner mention or union leaked into animal join: %q", ans)
	}
}

func TestComposeFromHopContentsIntersectsEntities(t *testing.T) {
	ans := composeFromHopContents([]HopResult{
		{
			Kind: "fetch_predicate", Entity: "Tim", Source: "search_fallback",
			Contents: []string{"Tim owns a jersey.", "Tim owns a signed baseball."},
		},
		{
			Kind: "fetch_predicate", Entity: "John", Source: "search_fallback",
			Contents: []string{"John owns a jersey.", "John owns vintage posters."},
		},
	})
	low := strings.ToLower(ans)
	if !strings.Contains(low, "jersey") {
		t.Fatalf("expected shared jersey, got %q", ans)
	}
	if strings.Contains(low, "baseball") || strings.Contains(low, "poster") {
		t.Fatalf("union leaked into join contents: %q", ans)
	}
}

func TestHopSlotValuesIgnoreSearchFallback(t *testing.T) {
	vals := hopSlotValues([]HopResult{
		{Kind: "fetch_predicate", Value: "watching pets play", Values: []string{"watching pets play"}, Source: "search_fallback"},
		{Kind: "fetch_predicate", Value: "turtles", Source: "typed_store"},
	})
	if len(vals) != 1 || !strings.EqualFold(vals[0], "turtles") {
		t.Fatalf("typed hop value only, got %#v", vals)
	}
}

func TestHopJoinProvenSearchFallbackIsNotTypedExact(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", OutputKey: "e1", Value: "Alex", Entity: "Alex", Source: "typed_store"},
		{Kind: "fetch_predicate", OutputKey: "ans", Entity: "Alex", Value: "nurse", Source: "search_fallback", ProofKind: "context", DependsOn: []string{"e1"}},
	}
	if hopJoinProven(hops) {
		t.Fatal("search_fallback must not yield hop_join_proven")
	}
}

func TestUnscopedPredicateIsNotTypedExact(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_ = store.UpsertCurrentState(context.Background(), "t1", "u1", PredicateOccupation, "m-dana", "nurse", "replace")
	_, _ = store.UpsertMemory(context.Background(), MemoryRecord{
		MemoryID: "m-dana", TenantID: "t1", SubjectID: "u1",
		Content: "Dana works as a nurse", DedupeKey: "d1", Status: StatusActive,
		Explain:  map[string]any{"subject": "Dana", "predicate": PredicateOccupation, "value_norm": "nurse"},
		Metadata: map[string]any{"subject": "Dana", "predicate": PredicateOccupation, "value_norm": "nurse", "entity_id": CanonicalEntityID("t1", "u1", "Dana")},
	})
	res := HopResult{Entity: "Alex", EntityID: CanonicalEntityID("t1", "u1", "Alex"), Predicate: PredicateOccupation, Source: "unresolved"}
	svc.fetchPredicateHop(context.Background(), "t1", "u1", "", false, HopStep{
		Kind: "fetch_predicate", Entity: "Alex", Predicate: PredicateOccupation,
	}, &res, 10)
	if res.Source == "typed_store" || res.ProofKind == "typed_exact" {
		t.Fatalf("unscoped occupation must not be typed_exact, got source=%s proof=%s value=%q", res.Source, res.ProofKind, res.Value)
	}
}

func TestContainsEntityMentionDoesNotStealFirstNamePrefix(t *testing.T) {
	if containsEntityMention("John Doe works as a nurse", "John Smith") {
		t.Fatal("full-name mention must not match a different John")
	}
	if !containsEntityMention("John Doe works as a nurse", "John Doe") {
		t.Fatal("full-name mention should match")
	}
}

func TestKinDestMentionRewritesBareRole(t *testing.T) {
	got := kinDestMention(HopResult{
		Entity:   "Alex",
		Value:    "mother",
		Contents: []string{"Alex's mother is her mom."},
	})
	if !strings.EqualFold(got, "Alex's mother") {
		t.Fatalf("bare role must rewrite to dest mention, got %q", got)
	}
	gloss := kinDestMention(HopResult{
		Entity:   "Alex",
		Value:    "her mom",
		Contents: []string{"Alex's mother is her mom."},
	})
	if !strings.EqualFold(gloss, "Alex's mother") {
		t.Fatalf("copula gloss must rewrite to dest mention, got %q", gloss)
	}
	named := kinDestMention(HopResult{
		Entity:   "Alex",
		Value:    "dana",
		Contents: []string{"Dana is Alex's mother"},
	})
	if !strings.EqualFold(named, "dana") {
		t.Fatalf("named dest must stay, got %q", named)
	}
}

func TestApplyHopDependencyClearsSourceIDForDest(t *testing.T) {
	srcID := CanonicalEntityID("t1", "u1", "Alex")
	res := HopResult{}
	applyHopDependency(&res, HopResult{
		Entity:   "Alex",
		EntityID: srcID,
		Value:    "mother",
		Contents: []string{"Alex's mother is her mom."},
	})
	if !strings.EqualFold(res.Entity, "Alex's mother") {
		t.Fatalf("expected dest entity, got %q", res.Entity)
	}
	if res.EntityID != "" {
		t.Fatalf("source entity id must not ride onto dest hop, got %q", res.EntityID)
	}
	same := HopResult{}
	applyHopDependency(&same, HopResult{Entity: "Alex", EntityID: srcID, Value: "Alex"})
	if same.EntityID != srcID {
		t.Fatalf("same-person dest must keep source id, got %q", same.EntityID)
	}
}

func TestRecordMatchesHopEntityBareRoleNeedsDestSubject(t *testing.T) {
	dest := MemoryRecord{
		Content:  "Alex's mother enjoyed reading",
		Metadata: map[string]any{"subject": "Alex's mother"},
	}
	src := MemoryRecord{
		Content:  "Alex visited mother's old house last year",
		Metadata: map[string]any{"subject": "Alex"},
	}
	if !recordMatchesHopEntity(dest, "mother", "") {
		t.Fatal("dest-subject record must match bare kin role")
	}
	if recordMatchesHopEntity(src, "mother", "") {
		t.Fatal("source possessive mention must not match bare kin role")
	}
	if !recordMatchesHopEntity(dest, "Alex's mother", "") {
		t.Fatal("rewritten dest mention must match dest record")
	}
	if recordMatchesHopEntity(src, "Alex's mother", "") {
		t.Fatal("rewritten dest mention must not match source visit")
	}
}

func TestAttitudeObjectSlotFromDestFacts(t *testing.T) {
	got, ok := attitudeObjectSlot("Alex's mother had reading as one of her hobbies, often sitting with a book.")
	if !ok || !strings.EqualFold(got, "reading") {
		t.Fatalf("expected reading, got %q ok=%v", got, ok)
	}
	got, ok = attitudeObjectSlot("Alex's mother was passionate about travel.")
	if !ok || !strings.Contains(strings.ToLower(got), "travel") {
		t.Fatalf("expected travel, got %q ok=%v", got, ok)
	}
	got, ok = attitudeObjectSlot("Alex's mother was interested in art.")
	if !ok || !strings.EqualFold(got, "art") {
		t.Fatalf("expected art, got %q ok=%v", got, ok)
	}
	if _, ok := attitudeObjectSlot("Alex visited mother's old house last year."); ok {
		t.Fatal("source visit must not yield an attitude slot")
	}
}

func TestPlayPracticeObjectSlots(t *testing.T) {
	got := playPracticeObjectSlots("Riley plays the clarinet.")
	if !containsFold(got, "clarinet") {
		t.Fatalf("expected clarinet, got %#v", got)
	}
	got = playPracticeObjectSlots("Riley does daily violin practice after work.")
	if !containsFold(got, "violin") {
		t.Fatalf("expected violin from practice, got %#v", got)
	}
	if got := playPracticeObjectSlots("Riley enjoys hiking"); len(got) != 0 {
		t.Fatalf("hobby must not yield play objects: %#v", got)
	}
}

func TestUnwindActivitySlots(t *testing.T) {
	got := unwindActivitySlots("Riley runs to destress")
	if !containsFold(got, "runs") {
		t.Fatalf("expected runs, got %#v", got)
	}
	got = unwindActivitySlots("Riley finds making pottery calming")
	if !containsFold(got, "pottery") {
		t.Fatalf("expected pottery, got %#v", got)
	}
}

func TestTrickObjectSlots(t *testing.T) {
	got := trickObjectSlots("James: They can do tricks like sit, stay, paw, and rollover")
	if !containsFold(got, "sit") || !containsFold(got, "rollover") {
		t.Fatalf("expected trick list, got %#v", got)
	}
	if got := trickObjectSlots("James coded in c++"); len(got) != 0 {
		t.Fatalf("non-trick content must not yield trick slots: %#v", got)
	}
}

func TestCompositionalPracticePlace(t *testing.T) {
	got := compositionalPracticePlace("Riley recommends the yoga studio nearby.", []string{"yoga"})
	if !strings.Contains(strings.ToLower(got), "studio") {
		t.Fatalf("expected yoga studio, got %q", got)
	}
	if p := compositionalPracticePlace("Riley does yoga on the beach.", []string{"yoga"}); p != "" {
		t.Fatalf("prep after practice object must not become a place: %q", p)
	}
	if p := compositionalPracticePlace("Riley met Alex at yoga in the park", []string{"yoga"}); p != "" {
		t.Fatalf("bare practice object plus next token is not a definite place: %q", p)
	}
	if p := compositionalPracticePlace("Yoga walking helped Riley", []string{"yoga"}); p != "" {
		t.Fatalf("gerund after practice object is not a place: %q", p)
	}
}

func TestHopSlotValuesDropsAttendedAndForeignPossessive(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Deborah", Value: "Deborah", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "mother", Predicate: PredicateFamilyMember, Source: "typed_store",
			Value: "mother", Values: []string{"mother"}},
		{Kind: "follow_relation", Entity: "mother", Predicate: PredicateActivity, Source: "typed_store",
			Value:  "reading, travel, art, cooking, attended deborah's yoga classes",
			Values: []string{"reading", "travel", "art", "cooking", "attended deborah's yoga classes"}},
	}
	vals := hopSlotValues(hops)
	joined := strings.ToLower(strings.Join(vals, ","))
	if !strings.Contains(joined, "reading") || !strings.Contains(joined, "cooking") {
		t.Fatalf("expected dest hobbies, got %#v", vals)
	}
	if strings.Contains(joined, "yoga") || strings.Contains(joined, "attended") {
		t.Fatalf("foreign attended event leaked into dest slots: %#v", vals)
	}
}

func TestHopSlotValuesKeepsVisitPossessiveDestination(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Alex", Value: "Alex", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Dana", Value: "Dana", Source: "typed_store"},
		{Kind: "follow_relation", Entity: "Alex", Predicate: PredicatePlan, Source: "typed_store",
			Value:  "write songs, visit dana's studio, attended dana's yoga class",
			Values: []string{"write songs", "visit dana's studio", "attended dana's yoga class"}},
	}
	vals := hopSlotValues(hops)
	joined := strings.ToLower(strings.Join(vals, ","))
	if !strings.Contains(joined, "studio") {
		t.Fatalf("visit destination dropped as foreign possessive: %#v", vals)
	}
	if strings.Contains(joined, "yoga") || strings.Contains(joined, "attended") {
		t.Fatalf("attended foreign event leaked into dest slots: %#v", vals)
	}
}

func TestPreferCoParticipantVisitDestination(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Alex", Value: "Alex"},
		{Kind: "resolve_entity", Entity: "Dana", Value: "Dana"},
		{Kind: "follow_relation", Entity: "Alex", Predicate: PredicatePlan,
			Values: []string{"write songs", "visit dana's studio", "travel to boston"}},
	}
	q := "What plans do Alex and Dana have for when Alex visits Boston?"
	got := preferCoParticipantVisitDestination(q, hops, "Dana will show Alex classic cars and meet at a bench.")
	if !strings.Contains(strings.ToLower(got), "studio") {
		t.Fatalf("expected visit destination over related-activity paraphrase, got %q", got)
	}
	kept := preferCoParticipantVisitDestination(q, hops, "They will visit Dana's studio while Alex is in Boston.")
	if !strings.Contains(strings.ToLower(kept), "studio") || strings.EqualFold(kept, "visit dana's studio") && !strings.Contains(strings.ToLower(kept), "boston") {
		// keep the richer hybrid that already names the place
		if !strings.Contains(strings.ToLower(kept), "studio") {
			t.Fatalf("hybrid that already names the place must be kept, got %q", kept)
		}
	}
	if preferCoParticipantVisitDestination("When did Alex adopt Ned?", hops, "first week of April 2022") != "first week of April 2022" {
		t.Fatal("non-visit queries must not be rewritten")
	}
}
