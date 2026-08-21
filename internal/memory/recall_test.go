package memory

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRecallEnumerateDistinctValues(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t1", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Dana: I've been hiking every weekend"},
			{Role: "user", Content: "Dana: I'm a big fan of ceramics"},
			{Role: "user", Content: "Dana: I've been swimming at the lake"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1", Query: "What activities does Dana enjoy?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Items) < 2 {
		t.Fatalf("expected enumerated items, got %#v context=%q", out.Items, out.ContextBlock)
	}
	joined := strings.ToLower(out.ContextBlock + " " + out.Answer)
	for _, need := range []string{"hik", "ceram", "swim"} {
		if !strings.Contains(joined, need) && !itemHas(out.Items, need) {
			t.Fatalf("expected %q in enumerate output %#v", need, out.Items)
		}
	}
}

func TestRecallUsesEvidencePacketAndModeHint(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-pkt", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex currently works as a nurse in Seattle."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-pkt", SubjectID: "u1", Query: "what does Alex currently do?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Explain["reader_source"] != "evidence_packet" {
		t.Fatalf("expected evidence_packet reader, explain=%v", out.Explain)
	}
	tools, _ := out.Explain["tools_executed"].([]string)
	if len(tools) == 0 || tools[0] != "search" {
		t.Fatalf("tools_executed=%v", tools)
	}
	if out.Abstained && out.ContextBlock == "" && out.Answer == "not in memory" {
		// Empty store path is ok only if search truly missed.
		t.Fatalf("unexpected total abstain with ingest: answer=%q context=%q", out.Answer, out.ContextBlock)
	}
}

func TestRecallAbstainsWhenPacketEmpty(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-empty", SubjectID: "u1", Query: "what is the secret code?", Mode: "answer", TopK: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Abstained {
		t.Fatalf("expected abstain on empty evidence, got %#v", out)
	}
}

func TestRecallEnumerateSkipsProvenanceEpisodes(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Yeah, Caroline, Yep, Melanie",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"primitive": PrimitiveEpisode},
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Dana enjoys hiking",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1", Query: "What activities does Dana enjoy?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.ContextBlock + " " + out.Answer)
	for _, it := range out.Items {
		joined += " " + strings.ToLower(it.Value)
	}
	if strings.Contains(joined, "yeah") {
		t.Fatalf("enumerate leaked provenance episode: %#v", out.Items)
	}
	if !strings.Contains(joined, "hik") {
		t.Fatalf("expected fact value, got %#v answer=%q", out.Items, out.Answer)
	}
}

func TestRecallRepresentationOracleIgnoresEpisodes(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["ep"] = MemoryRecord{
		MemoryID: "mem_ep", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Yeah, Caroline is from Sweden, Yep",
		DedupeKey: "ep", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"primitive": PrimitiveEpisode},
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1", Query: "Where is Caroline from?", Mode: "answer",
		OracleMode: "representation", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Explain["oracle_fact_count"] != 0 {
		t.Fatalf("episodes must not count as representation, explain=%v", out.Explain)
	}
	if out.Explain["oracle_episode_count"] != 1 {
		t.Fatalf("expected one episode, explain=%v", out.Explain)
	}
	if out.AnswerStatus != AnswerNotFound {
		t.Fatalf("expected representation_absent, status=%q answer=%q", out.AnswerStatus, out.Answer)
	}

	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Content: "Caroline is from Sweden",
		DedupeKey: "fact", Status: StatusActive, UpdatedAt: now,
	}
	out, err = svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1", Query: "Where is Caroline from?", Mode: "answer",
		OracleMode: "representation", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := out.Explain["oracle_fact_count"].(int); n < 1 {
		t.Fatalf("expected stored fact counted, explain=%v", out.Explain)
	}
	blob, _ := out.Explain["oracle_fact_blob"].(string)
	if !strings.Contains(strings.ToLower(blob), "sweden") {
		t.Fatalf("expected Sweden in fact blob, blob=%q", blob)
	}
}

func itemHas(items []RecallItem, needle string) bool {
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Value), needle) {
			return true
		}
	}
	return false
}

func TestAssembleContextHonorsMaxEvidenceTokens(t *testing.T) {
	pkt := EvidencePacket{
		Contents: []string{
			"Alex lives in Austin Texas now",
			"Alex used to live in New York",
			"Alex likes hiking on weekends",
		},
	}
	wide := assembleContextFromPacket(pkt, 4000)
	narrow := assembleContextFromPacket(pkt, 20)
	if narrow == "" {
		t.Fatal("small budget should still keep the first line")
	}
	if len(narrow) >= len(wide) {
		t.Fatalf("max evidence tokens must truncate, narrow=%d wide=%d", len(narrow), len(wide))
	}
}

func TestRecallUsesMaxEvidenceTokensBudget(t *testing.T) {
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-budget", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{{Role: "user", Content: "Alex lives in Austin and used to live in New York and likes hiking"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-budget", SubjectID: "u1", Query: "where does Alex live",
		Mode: "context", TopK: 10, MaxEvidenceTokens: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Explain["budget_tokens"] != 12 {
		t.Fatalf("expected max_evidence_tokens to win budget, explain=%v", out.Explain)
	}
}

func TestRecallOriginHopAnswersPlaceNotAnaphora(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-origin", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: This necklace is from my home country, Portugal."},
			{Role: "user", Content: "Jordan: I've known these friends for four years, since I moved from my home country."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-origin", SubjectID: "u1",
		Query: "Where did Jordan move from 4 years ago?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.Answer + " " + strings.Join(func() []string {
		vals := make([]string, 0, len(out.Items))
		for _, it := range out.Items {
			vals = append(vals, it.Value)
		}
		return vals
	}(), " "))
	if !strings.Contains(joined, "portugal") {
		t.Fatalf("expected origin place in answer, got answer=%q items=%#v explain_hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(strings.ToLower(out.Answer), "home country") && !strings.Contains(joined, "portugal") {
		t.Fatalf("anaphora leaked without place: %q", out.Answer)
	}
}

func TestRecallCareerEnumeratesFieldAndPopulation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-career", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: I'm looking into counseling and mental health as a career."},
			{Role: "user", Content: "Jordan: I'm thinking of working with elderly patients."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-career", SubjectID: "u1",
		Query: "What career path has Jordan decided to pursue?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.Answer + " " + out.ContextBlock)
	for _, it := range out.Items {
		joined += " " + strings.ToLower(it.Value)
	}
	if !strings.Contains(joined, "counseling") {
		t.Fatalf("expected counseling, items=%#v answer=%q", out.Items, out.Answer)
	}
	if !strings.Contains(joined, "elderly") {
		t.Fatalf("expected population qualifier, items=%#v answer=%q", out.Items, out.Answer)
	}
}

func TestRecallKidsLikesSkipsJunkAndKeepsExhibitNoun(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-kids", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Riley: Yesterday I took the kids to the museum."},
			{Role: "user", Content: "Riley: They were stoked for the fossils exhibit!"},
			{Role: "user", Content: "Riley: The kids love nature."},
			{Role: "user", Content: "Riley: The kids were talking about our last one over summer break."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-kids", SubjectID: "u1",
		Query: "What do Riley's kids like?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.ContextBlock)
	for _, it := range out.Items {
		joined += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(joined, "fossil") {
		t.Fatalf("expected fossils, items=%#v", out.Items)
	}
	if !strings.Contains(joined, "nature") {
		t.Fatalf("expected nature, items=%#v", out.Items)
	}
	if strings.Contains(joined, "summer break") || strings.Contains(joined, "our last one") {
		t.Fatalf("junk preference leaked: %#v", out.Items)
	}
}

func TestRecallResearchEnumeratesTopicNotIdentity(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-research", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Jordan: I am a marine biologist."},
			{Role: "user", Content: "Jordan: Researching scholarship programs — it's been a dream to help."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-research", SubjectID: "u1",
		Query: "What did Jordan research?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.Answer + " " + out.ContextBlock)
	for _, it := range out.Items {
		joined += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(joined, "scholarship") {
		t.Fatalf("expected researched topic, items=%#v answer=%q", out.Items, out.Answer)
	}
	if strings.Contains(joined, "marine biologist") {
		t.Fatalf("identity leaked into research enumerate: %#v", out.Items)
	}
}

func TestSlotValueKeepsQuotedTitleContainingIs(t *testing.T) {
	got, ok := slotValueFromMemoryContent(`Riley read "Life is Elsewhere" (12 July 2022)`)
	if !ok || !strings.EqualFold(got, "Life is Elsewhere") {
		t.Fatalf("quoted title, got %q ok=%v", got, ok)
	}
	if got, ok := slotValueFromMemoryContent("Life is Elsewhere"); ok && strings.EqualFold(got, "Elsewhere") {
		t.Fatalf("must not split title on is, got %q", got)
	}
	got, ok = slotValueFromMemoryContent("Riley is single")
	if !ok || !strings.EqualFold(got, "single") {
		t.Fatalf("identity copula, got %q ok=%v", got, ok)
	}
	got, ok = slotValueFromMemoryContent("Melanie realized that self-care is important")
	if !ok || !strings.Contains(strings.ToLower(got), "self-care") {
		t.Fatalf("realized-that clause, got %q ok=%v", got, ok)
	}
	if strings.EqualFold(strings.TrimSpace(got), "important") {
		t.Fatalf("must not clip clause copula to adjective tail, got %q", got)
	}
	got, ok = slotValueFromMemoryContent("Riley: This book I read last year reminds me. [visible text: LIFE IS ELSEWHERE]")
	if !ok || !strings.Contains(strings.ToLower(got), "life is elsewhere") {
		t.Fatalf("visible-text title, got %q ok=%v", got, ok)
	}
	if strings.EqualFold(got, "Elsewhere") {
		t.Fatalf("visible-text title split on is: %q", got)
	}
	if hasSlotTemplate("nothing is impossible") || hasSlotTemplate("life is elsewhere") {
		t.Fatal("lowercase title dst is not a slot template")
	}
}

func TestRecallBooksRejectOneWordQuoteAndKeepTitleCaseRun(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-books", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: `Riley: I loved reading "The Little Prince" as a kid.`},
			{Role: "user", Content: "Riley: This book I read last year, The Hidden Garden, still stays with me."},
			{Role: "user", Content: "Riley: This book I read last year reminds me to keep going. [visible text: THE QUIET ORCHARD]"},
			{Role: "user", Content: `Riley: I loved reading "Life is Elsewhere" last spring.`},
			{Role: "user", Content: `Riley: I read "Perfect"`},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-books", SubjectID: "u1",
		Query: "What books has Riley read?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.ToLower(out.ContextBlock)
	for _, it := range out.Items {
		joined += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(joined, "little prince") {
		t.Fatalf("expected quoted title, items=%#v", out.Items)
	}
	if !strings.Contains(joined, "hidden garden") {
		t.Fatalf("expected unquoted title run, items=%#v", out.Items)
	}
	if !strings.Contains(joined, "quiet orchard") {
		t.Fatalf("expected deictic visible title, items=%#v answer=%q", out.Items, out.Answer)
	}
	if !strings.Contains(joined, "life is elsewhere") {
		t.Fatalf("title containing is must stay intact, items=%#v answer=%q", out.Items, out.Answer)
	}
	if itemHas(out.Items, "elsewhere") && !itemHas(out.Items, "life is elsewhere") {
		t.Fatalf("must not split title on is, items=%#v", out.Items)
	}
	if strings.Contains(joined, `"perfect"`) || itemHas(out.Items, "Perfect") {
		t.Fatalf("one-word quote should be rejected, items=%#v", out.Items)
	}
}

func TestRecallPrefersStructuredValueOverSlogan(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["slogan"] = MemoryRecord{
		MemoryID: "mem_s", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Melanie after the charity race: We Can Really Accept Who We Are And Be Content",
		DedupeKey: "s", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"primitive": PrimitiveEpisode, "rule": "conversation_episode"},
	}
	store.records["fact"] = MemoryRecord{
		MemoryID: "mem_f", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Melanie realized that self-care is important",
		DedupeKey: "f", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateBelief, "value_norm": "self-care is important"},
		Explain:  map[string]any{"predicate": PredicateBelief, "value_norm": "self-care is important"},
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1",
		Query: "What did Melanie realize after the charity race?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if strings.Contains(got, "we can really accept") {
		t.Fatalf("slogan must not be the answer, got %q", out.Answer)
	}
	if !strings.Contains(got, "self-care") {
		t.Fatalf("expected structured value, got answer=%q abstained=%v", out.Answer, out.Abstained)
	}
}

func TestRecallAbstainsOnEpisodeOnlySlogan(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["slogan"] = MemoryRecord{
		MemoryID: "mem_s", TenantID: "t1", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "We Can Really Accept Who We Are And Be Content",
		DedupeKey: "s", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"primitive": PrimitiveEpisode},
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1",
		Query: "What did Melanie realize after the charity race?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Abstained || out.Answer != "not in memory" {
		t.Fatalf("expected abstain without structured facts, got answer=%q abstained=%v", out.Answer, out.Abstained)
	}
}

func TestRecallEnumerateSkipsOtherPredicates(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["plan"] = MemoryRecord{
		MemoryID: "mem_p", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Caroline researched adoption agencies",
		DedupeKey: "p", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePlan, "value_norm": "adoption agencies"},
		Explain:  map[string]any{"predicate": PredicatePlan, "value_norm": "adoption agencies"},
	}
	store.records["media"] = MemoryRecord{
		MemoryID: "mem_m", TenantID: "t1", SubjectID: "u1",
		Kind:      KindFact,
		Content:   "Caroline read Wicked",
		DedupeKey: "m", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateMediaConsumed, "value_norm": "wicked"},
		Explain:  map[string]any{"predicate": PredicateMediaConsumed, "value_norm": "wicked"},
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t1", SubjectID: "u1",
		Query: "What are Caroline's summer plans?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " " + strings.ToLower(it.Value)
	}
	if strings.Contains(got, "wicked") {
		t.Fatalf("media fact leaked into plans answer: answer=%q items=%#v", out.Answer, out.Items)
	}
	if !strings.Contains(strings.ToLower(out.Answer), "adoption") {
		t.Fatalf("expected plan value, got answer=%q items=%#v", out.Answer, out.Items)
	}
}

func TestPickStructuredAnswerPrefersPredicateMatch(t *testing.T) {
	results := []SearchResult{
		{
			MemoryID: "ep", Content: "We Can Really Accept Who We Are And Be Content",
			Explain: map[string]any{"primitive": PrimitiveEpisode},
		},
		{
			MemoryID: "media", Content: "Caroline read Wicked",
			Explain: map[string]any{"predicate": PredicateMediaConsumed, "value_norm": "wicked"},
		},
		{
			MemoryID: "plan", Content: "Caroline researched adoption agencies",
			Explain: map[string]any{"predicate": PredicatePlan, "value_norm": "adoption agencies"},
		},
	}
	got := pickStructuredAnswer("What are Caroline's summer plans?", results)
	if !strings.Contains(strings.ToLower(got), "adoption") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(strings.ToLower(got), "wicked") {
		t.Fatalf("media leaked: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "we can really") {
		t.Fatalf("slogan leaked: %q", got)
	}
}

func TestRecallIngestPrefersResearchedPlanOverMedia(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-plan", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Caroline researched adoption agencies"},
			{Role: "user", Content: "Caroline read Wicked"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-plan", SubjectID: "u1",
		Query: "What are Caroline's summer plans?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if out.Abstained || strings.Contains(got, "wicked") {
		t.Fatalf("media leaked or abstained: answer=%q items=%#v", out.Answer, out.Items)
	}
	if !strings.Contains(got, "adoption") {
		t.Fatalf("expected researched plan value, got %q", out.Answer)
	}
}

func TestRecallIngestSloganOnlyAbstains(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-slogan", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "We Can Really Accept Who We Are And Be Content"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-slogan", SubjectID: "u1",
		Query: "What did Melanie realize after the charity race?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Abstained || out.Answer != "not in memory" {
		t.Fatalf("expected abstain on slogan-only ingest, got answer=%q abstained=%v", out.Answer, out.Abstained)
	}
}

func TestRecallIngestPointFactFromWorksAs(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-nurse", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Alex currently works as a nurse in Seattle."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-nurse", SubjectID: "u1",
		Query: "what does Alex currently do?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Abstained || strings.TrimSpace(out.Answer) == "" || out.Answer == "not in memory" {
		t.Fatalf("point fact blanked: %#v", out)
	}
	if !strings.Contains(strings.ToLower(out.Answer), "nurse") {
		t.Fatalf("expected occupation, got %q", out.Answer)
	}
}

func TestRecallIngestRealizedClauseNotAdjectiveTail(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-realize", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Melanie after the charity race: We Can Really Accept Who We Are And Be Content"},
			{Role: "user", Content: "Melanie realized that self-care is important"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-realize", SubjectID: "u1",
		Query: "What did Melanie realize after the charity race?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(strings.TrimSpace(out.Answer))
	if out.Abstained || strings.Contains(got, "we can really accept") {
		t.Fatalf("slogan or abstain: answer=%q abstained=%v", out.Answer, out.Abstained)
	}
	if got == "important" || got == "important." {
		t.Fatalf("clipped copula tail: %q", out.Answer)
	}
	if !strings.Contains(got, "self-care") {
		t.Fatalf("expected realized clause, got %q", out.Answer)
	}
}

func TestRecallIngestNamedSubjectNotReporter(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	_, err := svc.Ingest(context.Background(), IngestRequest{
		TenantID: "t-named", SubjectID: "u1", SourceType: "conversation",
		Messages: []Message{
			{Role: "user", Content: "Riley: Casey researched wildfire recovery last spring."},
			{Role: "user", Content: "Riley: I went swimming."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-named", SubjectID: "u1",
		Query: "What did Casey research?", Mode: "answer", TopK: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer + " " + out.ContextBlock)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if out.Abstained || !strings.Contains(got, "wildfire") {
		t.Fatalf("expected named-subject research, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(strings.ToLower(out.Answer), "swimming") {
		t.Fatalf("reporter activity leaked: %q", out.Answer)
	}
}

func TestRecallPreferenceHopUsesStructuredProof(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["nate-pref"] = MemoryRecord{
		MemoryID: "mem_n", TenantID: "t-mh-pref", SubjectID: "u1",
		Kind: KindFact, Content: "Nate enjoys turtles",
		DedupeKey: "n", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "turtles", "subject": "Nate"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "turtles", "subject": "Nate"},
	}
	store.records["joanna-pref"] = MemoryRecord{
		MemoryID: "mem_j", TenantID: "t-mh-pref", SubjectID: "u1",
		Kind: KindFact, Content: "Joanna enjoys turtles",
		DedupeKey: "j", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "turtles", "subject": "Joanna"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "turtles", "subject": "Joanna"},
	}
	store.records["slogan"] = MemoryRecord{
		MemoryID: "mem_s", TenantID: "t-mh-pref", SubjectID: "u1",
		Kind: KindFact, Primitive: PrimitiveEpisode,
		Content:   "Games Are Both Competitive And Chill",
		DedupeKey: "s", Status: StatusActive, UpdatedAt: now,
		Explain: map[string]any{"primitive": PrimitiveEpisode},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePreference, val: "turtles", memID: "mem_n"},
		stubAtom{pred: PredicatePreference, val: "turtles", memID: "mem_j"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-mh-pref", SubjectID: "u1",
		Query: "What animal do both Nate and Joanna like?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer + " " + out.ContextBlock)
	if !strings.Contains(got, "turtle") {
		t.Fatalf("expected turtles in proof/answer, answer=%q context=%q", out.Answer, out.ContextBlock)
	}
	if strings.Contains(strings.ToLower(out.Answer), "competitive and chill") {
		t.Fatalf("slogan crowded out proof: %q", out.Answer)
	}
	cov, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-mh-pref", SubjectID: "u1",
		Query: "What animal do both Nate and Joanna like?", OracleMode: "coverage", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cov.Coverage == nil || cov.Coverage["satisfied"] != true {
		t.Fatalf("coverage should be satisfied when structured preference is in the packet, coverage=%v answer=%q", cov.Coverage, cov.Answer)
	}
}

func TestRecallCoordinatedJoinWithoutBothCue(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["tim-jersey"] = MemoryRecord{
		MemoryID: "mem_tj", TenantID: "t-mh-join", SubjectID: "u1",
		Kind: KindFact, Content: "Tim owns a jersey",
		DedupeKey: "tj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "Tim"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "Tim"},
	}
	store.records["john-jersey"] = MemoryRecord{
		MemoryID: "mem_jj", TenantID: "t-mh-join", SubjectID: "u1",
		Kind: KindFact, Content: "John owns a jersey",
		DedupeKey: "jj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "jersey", "subject": "John"},
	}
	store.records["john-ball"] = MemoryRecord{
		MemoryID: "mem_jb", TenantID: "t-mh-join", SubjectID: "u1",
		Kind: KindFact, Content: "John owns a signed baseball",
		DedupeKey: "jb", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "signed baseball", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "signed baseball", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "jersey", memID: "mem_tj"},
		stubAtom{pred: PredicatePossession, val: "jersey", memID: "mem_jj"},
		stubAtom{pred: PredicatePossession, val: "signed baseball", memID: "mem_jb"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-mh-join", SubjectID: "u1",
		Query: "What similar collectible do Tim and John own?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "jersey") {
		t.Fatalf("expected shared jersey, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "baseball") {
		t.Fatalf("private possession leaked into join: %q", out.Answer)
	}
}

func TestRecallEntityListDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["riley-dog"] = MemoryRecord{
		MemoryID: "mem_rd", TenantID: "t-list", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's dog is named Max",
		DedupeKey: "rd", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "max", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "max", "subject": "Riley"},
	}
	store.records["riley-job"] = MemoryRecord{
		MemoryID: "mem_rj", TenantID: "t-list", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "max", memID: "mem_rd"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-list", SubjectID: "u1",
		Query: "What are Riley's dogs' names?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "max") {
		t.Fatalf("expected pet name, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded the names list: %q", out.Answer)
	}
}

func TestRecallInstrumentListDoesNotDumpHobby(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["jordan-play"] = MemoryRecord{
		MemoryID: "mem_jp", TenantID: "t-list2", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan plays piano",
		DedupeKey: "jp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateSkill, "value_norm": "piano", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateSkill, "value_norm": "piano", "subject": "Jordan"},
	}
	store.records["jordan-hike"] = MemoryRecord{
		MemoryID: "mem_jh", TenantID: "t-list2", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan enjoys hiking",
		DedupeKey: "jh", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Jordan"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateSkill, val: "piano", memID: "mem_jp"},
		stubAtom{pred: PredicatePreference, val: "hiking", memID: "mem_jh"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-list2", SubjectID: "u1",
		Query: "What instruments does Jordan play?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "piano") {
		t.Fatalf("expected instrument, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "hiking") {
		t.Fatalf("hobby crowded the instrument list: %q", out.Answer)
	}
}

func TestRecallPetTrickListDoesNotDumpHobby(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["james-trick"] = MemoryRecord{
		MemoryID: "mem_jt", TenantID: "t-list3", SubjectID: "u1",
		Kind: KindFact, Content: "James's dog knows sit",
		DedupeKey: "jt", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateSkill, "value_norm": "sit", "subject": "James"},
		Explain:  map[string]any{"predicate": PredicateSkill, "value_norm": "sit", "subject": "James"},
	}
	store.records["james-hike"] = MemoryRecord{
		MemoryID: "mem_jh2", TenantID: "t-list3", SubjectID: "u1",
		Kind: KindFact, Content: "James enjoys hiking",
		DedupeKey: "jh2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "James"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "James"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateSkill, val: "sit", memID: "mem_jt"},
		stubAtom{pred: PredicatePreference, val: "hiking", memID: "mem_jh2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-list3", SubjectID: "u1",
		Query: "What kind of tricks do James's pets know?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "sit") {
		t.Fatalf("expected trick, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "hiking") {
		t.Fatalf("hobby crowded the trick list: %q", out.Answer)
	}
}

func TestRecallCountPossessionsDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["cal-car1"] = MemoryRecord{
		MemoryID: "mem_c1", TenantID: "t-count", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a red coupe",
		DedupeKey: "c1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "red coupe", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "red coupe", "subject": "Calvin"},
	}
	store.records["cal-car2"] = MemoryRecord{
		MemoryID: "mem_c2", TenantID: "t-count", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a black sedan",
		DedupeKey: "c2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "black sedan", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "black sedan", "subject": "Calvin"},
	}
	store.records["cal-job"] = MemoryRecord{
		MemoryID: "mem_cj", TenantID: "t-count", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin works as a nurse",
		DedupeKey: "cj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Calvin"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "red coupe", memID: "mem_c1"},
		stubAtom{pred: PredicatePossession, val: "black sedan", memID: "mem_c2"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_cj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-count", SubjectID: "u1",
		Query: "How many cars does Calvin own?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Answer != "2" {
		t.Fatalf("expected count 2, answer=%q items=%#v", out.Answer, out.Items)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded the count: %q", out.Answer)
	}
}

func TestRecallCountTimesUsesEvidenceNotUniqueSlot(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-ank1"] = MemoryRecord{
		MemoryID: "mem_ja1", TenantID: "t-times", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan injured his ankle",
		DedupeKey: "ja1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
	}
	store.records["j-ank2"] = MemoryRecord{
		MemoryID: "mem_ja2", TenantID: "t-times", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan injured his ankle again",
		DedupeKey: "ja2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateHealth, val: "ankle", memID: "mem_ja1"},
		stubAtom{pred: PredicateHealth, val: "ankle", memID: "mem_ja2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-times", SubjectID: "u1",
		Query: "How many times has Jordan injured his ankle?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Answer != "2" {
		t.Fatalf("expected 2 incidents, answer=%q items=%#v", out.Answer, out.Items)
	}
}

func TestRecallPolarYesFromTriedActivity(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-surf"] = MemoryRecord{
		MemoryID: "mem_rs", TenantID: "t-polar", SubjectID: "u1",
		Kind: KindFact, Content: "Riley tried surfing",
		DedupeKey: "rs", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "surfing", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "surfing", "subject": "Riley"},
	}
	store.records["r-job"] = MemoryRecord{
		MemoryID: "mem_rj2", TenantID: "t-polar", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rj2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "surfing", memID: "mem_rs"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rj2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-polar", SubjectID: "u1",
		Query: "Has Riley tried surfing?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(strings.TrimSpace(out.Answer), "Yes") {
		t.Fatalf("expected Yes, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
}

func TestRecallPolarDoesNotYesFromUnrelatedHobby(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-hike"] = MemoryRecord{
		MemoryID: "mem_rh", TenantID: "t-polar2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys hiking",
		DedupeKey: "rh", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePreference, val: "hiking", memID: "mem_rh"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-polar2", SubjectID: "u1",
		Query: "Has Riley tried surfing?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(strings.TrimSpace(out.Answer), "Yes") {
		t.Fatalf("unrelated hobby must not prove polar yes: %q", out.Answer)
	}
}

func TestRecallPracticeLocationListDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-yoga"] = MemoryRecord{
		MemoryID: "mem_ry", TenantID: "t-loc", SubjectID: "u1",
		Kind: KindFact, Content: "Riley practices yoga at the lake",
		DedupeKey: "ry", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "lake", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "lake", "subject": "Riley"},
	}
	store.records["r-job3"] = MemoryRecord{
		MemoryID: "mem_rj3", TenantID: "t-loc", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rj3", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "lake", memID: "mem_ry"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rj3"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-loc", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "lake") {
		t.Fatalf("expected practice location, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded the location list: %q", out.Answer)
	}
}

func TestFilterBesides(t *testing.T) {
	items := []RecallItem{{Value: "hiking"}, {Value: "deadlines at work"}}
	q := "What is the biggest stressor in Andrew's life besides not being able to hike frequently?"
	got := filterBesides(q, items)
	if len(got) != 1 || !strings.Contains(strings.ToLower(got[0].Value), "deadline") {
		t.Fatalf("besides filter=%#v", got)
	}
}

func TestRecallUnwindActivityDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-unw"] = MemoryRecord{
		MemoryID: "mem_ru", TenantID: "t-unw", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys ceramics",
		DedupeKey: "ru", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Riley"},
	}
	store.records["r-job4"] = MemoryRecord{
		MemoryID: "mem_rj4", TenantID: "t-unw", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rj4", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "ceramics", memID: "mem_ru"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rj4"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-unw", SubjectID: "u1",
		Query: "What does Riley do to unwind?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "ceram") {
		t.Fatalf("expected unwind activity, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded unwind list: %q", out.Answer)
	}
}

func TestRecallSuperlativeVisitPicksMostEvidence(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["t-jp1"] = MemoryRecord{
		MemoryID: "mem_jp1", TenantID: "t-vis", SubjectID: "u1",
		Kind: KindFact, Content: "Tim visited Japan",
		DedupeKey: "jp1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "japan", "subject": "Tim"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "japan", "subject": "Tim"},
	}
	store.records["t-jp2"] = MemoryRecord{
		MemoryID: "mem_jp2", TenantID: "t-vis", SubjectID: "u1",
		Kind: KindFact, Content: "Tim visited Japan again",
		DedupeKey: "jp2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "japan", "subject": "Tim"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "japan", "subject": "Tim"},
	}
	store.records["t-fr"] = MemoryRecord{
		MemoryID: "mem_fr", TenantID: "t-vis", SubjectID: "u1",
		Kind: KindFact, Content: "Tim visited France",
		DedupeKey: "fr", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "france", "subject": "Tim"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "france", "subject": "Tim"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "japan", memID: "mem_jp1"},
		stubAtom{pred: PredicateActivity, val: "japan", memID: "mem_jp2"},
		stubAtom{pred: PredicateActivity, val: "france", memID: "mem_fr"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-vis", SubjectID: "u1",
		Query: "which country has Tim visited most frequently in his travels?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out.Answer), "japan") {
		t.Fatalf("expected most-visited country, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(strings.ToLower(out.Answer), "france") {
		t.Fatalf("superlative dumped the union: %q", out.Answer)
	}
}

func TestRecallBesidesDropsExcludedStressor(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["a-hike"] = MemoryRecord{
		MemoryID: "mem_ah", TenantID: "t-bes", SubjectID: "u1",
		Kind: KindFact, Content: "Andrew enjoys hiking",
		DedupeKey: "ah", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Andrew"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Andrew"},
	}
	store.records["a-work"] = MemoryRecord{
		MemoryID: "mem_aw", TenantID: "t-bes", SubjectID: "u1",
		Kind: KindFact, Content: "Andrew enjoys deadlines at work",
		DedupeKey: "aw", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "deadlines at work", "subject": "Andrew"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "deadlines at work", "subject": "Andrew"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_ah"},
		stubAtom{pred: PredicateActivity, val: "deadlines at work", memID: "mem_aw"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-bes", SubjectID: "u1",
		Query: "What is the biggest stressor in Andrew's life besides not being able to hike frequently?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "deadline") {
		t.Fatalf("expected remaining stressor, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("besides clause did not drop excluded item: %q", out.Answer)
	}
}

func TestRecallWhoSupportsFromTypedHop(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["c-sup"] = MemoryRecord{
		MemoryID: "mem_cs", TenantID: "t-who", SubjectID: "u1",
		Kind: KindFact, Content: "Dana supports Calvin in tough times",
		DedupeKey: "cs", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Calvin"},
	}
	store.records["c-job"] = MemoryRecord{
		MemoryID: "mem_cj2", TenantID: "t-who", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin works as a nurse",
		DedupeKey: "cj2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Calvin"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "dana", memID: "mem_cs"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_cj2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-who", SubjectID: "u1",
		Query: "Who supports Calvin in tough times?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "dana") {
		t.Fatalf("expected supporter, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded who-answer: %q", out.Answer)
	}
}

func TestRecallChildhoodItemsNotFamily(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-toy"] = MemoryRecord{
		MemoryID: "mem_jt2", TenantID: "t-ch", SubjectID: "u1",
		Kind: KindFact, Content: "John owned a toy train as a child",
		DedupeKey: "jt2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "toy train", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "toy train", "subject": "John"},
	}
	store.records["j-sis"] = MemoryRecord{
		MemoryID: "mem_js", TenantID: "t-ch", SubjectID: "u1",
		Kind: KindFact, Content: "John's sister is Dana",
		DedupeKey: "js", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "toy train", memID: "mem_jt2"},
		stubAtom{pred: PredicateFamilyMember, val: "dana", memID: "mem_js"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-ch", SubjectID: "u1",
		Query: "What items does John mention having as a child?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "train") {
		t.Fatalf("expected childhood item, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "dana") {
		t.Fatalf("family crowded childhood items: %q", out.Answer)
	}
}

func TestRecallWhenInjuryDateNotEventDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	d2023 := time.Date(2023, 5, 7, 12, 0, 0, 0, time.UTC)
	d2022 := time.Date(2022, 1, 15, 12, 0, 0, 0, time.UTC)
	store.records["j-ank"] = MemoryRecord{
		MemoryID: "mem_ja3", TenantID: "t-when", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan injured his ankle",
		DedupeKey: "ja3", Status: StatusActive, UpdatedAt: now, ObservedAt: &d2023,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "ankle", "subject": "Jordan"},
	}
	store.records["j-wrist"] = MemoryRecord{
		MemoryID: "mem_jw", TenantID: "t-when", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan injured his wrist",
		DedupeKey: "jw", Status: StatusActive, UpdatedAt: now, ObservedAt: &d2022,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "wrist", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "wrist", "subject": "Jordan"},
	}
	store.records["j-job"] = MemoryRecord{
		MemoryID: "mem_jj", TenantID: "t-when", SubjectID: "u1",
		Kind: KindFact, Content: "Jordan works as a nurse",
		DedupeKey: "jj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Jordan"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Jordan"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateHealth, val: "ankle", memID: "mem_ja3"},
		stubAtom{pred: PredicateHealth, val: "wrist", memID: "mem_jw"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_jj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-when", SubjectID: "u1",
		Query: "When did Jordan get an ankle injury in 2023?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "7 may 2023") && !strings.Contains(got, "may 7") {
		t.Fatalf("expected dated injury, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") || strings.Contains(got, "wrist") {
		t.Fatalf("when-query dumped unrelated slots: %q", out.Answer)
	}
}

func TestRecallTransferKeepsRecipientNotJoin(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["e-sug"] = MemoryRecord{
		MemoryID: "mem_es", TenantID: "t-xfer", SubjectID: "u1",
		Kind: KindFact, Content: "Evan suggested quinoa bowls to Sam",
		DedupeKey: "es", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "quinoa bowls", "subject": "Evan"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "quinoa bowls", "subject": "Evan"},
	}
	store.records["e-soda"] = MemoryRecord{
		MemoryID: "mem_eso", TenantID: "t-xfer", SubjectID: "u1",
		Kind: KindFact, Content: "Evan enjoys soda",
		DedupeKey: "eso", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "soda", "subject": "Evan"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "soda", "subject": "Evan"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePreference, val: "quinoa bowls", memID: "mem_es"},
		stubAtom{pred: PredicatePreference, val: "soda", memID: "mem_eso"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-xfer", SubjectID: "u1",
		Query: "What kind of healthy food suggestions has Evan given to Sam?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "quinoa") {
		t.Fatalf("expected transferred suggestion, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "soda") {
		t.Fatalf("giver's unrelated preference crowded transfer: %q", out.Answer)
	}
}

func TestRecallAfterClauseKeepsMatchingMeals(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["s-meal"] = MemoryRecord{
		MemoryID: "mem_sm", TenantID: "t-after", SubjectID: "u1",
		Kind: KindFact, Content: "Sam started eating quinoa bowls after a health scare",
		DedupeKey: "sm", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "quinoa bowls", "subject": "Sam"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "quinoa bowls", "subject": "Sam"},
	}
	store.records["s-candy"] = MemoryRecord{
		MemoryID: "mem_sc", TenantID: "t-after", SubjectID: "u1",
		Kind: KindFact, Content: "Sam enjoys candy",
		DedupeKey: "sc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "candy", "subject": "Sam"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "candy", "subject": "Sam"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePreference, val: "quinoa bowls", memID: "mem_sm"},
		stubAtom{pred: PredicatePreference, val: "candy", memID: "mem_sc"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-after", SubjectID: "u1",
		Query: "What kind of healthy meals did Sam start eating after getting a health scare?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "quinoa") {
		t.Fatalf("expected after-clause meal, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "candy") {
		t.Fatalf("unrelated meal crowded after-clause: %q", out.Answer)
	}
}

func TestRecallCommunityListDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-com"] = MemoryRecord{
		MemoryID: "mem_rc", TenantID: "t-com", SubjectID: "u1",
		Kind: KindFact, Content: "Riley participates in the community garden",
		DedupeKey: "rc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Riley"},
	}
	store.records["r-job3"] = MemoryRecord{
		MemoryID: "mem_rj3", TenantID: "t-com", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rj3", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "community garden", memID: "mem_rc"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rj3"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-com", SubjectID: "u1",
		Query: "In what ways is Riley participating in the community?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "garden") {
		t.Fatalf("expected community activity, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded community list: %q", out.Answer)
	}
}

func TestRecallWhoInjuredInFamily(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["e-fam"] = MemoryRecord{
		MemoryID: "mem_ef", TenantID: "t-famh", SubjectID: "u1",
		Kind: KindFact, Content: "Casey is Evan's sister",
		DedupeKey: "ef", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "casey", "subject": "Evan"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "casey", "subject": "Evan"},
	}
	store.records["c-inj"] = MemoryRecord{
		MemoryID: "mem_ci", TenantID: "t-famh", SubjectID: "u1",
		Kind: KindFact, Content: "Casey injured her wrist",
		DedupeKey: "ci", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "wrist", "subject": "Casey"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "wrist", "subject": "Casey"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "casey", memID: "mem_ef"},
		stubAtom{pred: PredicateHealth, val: "wrist", memID: "mem_ci"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-famh", SubjectID: "u1",
		Query: "Who was injured in Evan's family?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "casey") {
		t.Fatalf("expected injured family member, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "wrist") && !strings.Contains(got, "casey") {
		t.Fatalf("health slot dumped instead of who: %q", out.Answer)
	}
}

func TestRecallOrgBeneficiariesFromAffiliation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-aff"] = MemoryRecord{
		MemoryID: "mem_jaff", TenantID: "t-org", SubjectID: "u1",
		Kind: KindFact, Content: "John's charity tournaments benefit Lakeside Shelter",
		DedupeKey: "jaff", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateAffiliation, "value_norm": "Lakeside Shelter", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateAffiliation, "value_norm": "Lakeside Shelter", "subject": "John"},
	}
	store.records["j-job4"] = MemoryRecord{
		MemoryID: "mem_jj4", TenantID: "t-org", SubjectID: "u1",
		Kind: KindFact, Content: "John works as a nurse",
		DedupeKey: "jj4", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateAffiliation, val: "Lakeside Shelter", memID: "mem_jaff"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_jj4"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-org", SubjectID: "u1",
		Query: "Who or which organizations have been the beneficiaries of John's charity tournaments?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "lakeside") {
		t.Fatalf("expected beneficiary org, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded beneficiary answer: %q", out.Answer)
	}
}

func TestRecallWhereKinshipPlaceNotActivityDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-div"] = MemoryRecord{
		MemoryID: "mem_jd", TenantID: "t-where", SubjectID: "u1",
		Kind: KindFact, Content: "Jolene and her partner found a cool diving spot in Bali",
		DedupeKey: "jd", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "diving", "subject": "Jolene"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "diving", "subject": "Jolene"},
	}
	store.records["j-par"] = MemoryRecord{
		MemoryID: "mem_jp", TenantID: "t-where", SubjectID: "u1",
		Kind: KindFact, Content: "Jolene's partner is Alex",
		DedupeKey: "jp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "alex", "subject": "Jolene"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "alex", "subject": "Jolene"},
	}
	store.records["j-jobw"] = MemoryRecord{
		MemoryID: "mem_jjw", TenantID: "t-where", SubjectID: "u1",
		Kind: KindFact, Content: "Jolene works as a nurse",
		DedupeKey: "jjw", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Jolene"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Jolene"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "diving", memID: "mem_jd"},
		stubAtom{pred: PredicateFamilyMember, val: "alex", memID: "mem_jp"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_jjw"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-where", SubjectID: "u1",
		Query: "Where did Jolene and her partner find a cool diving spot?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "bali") {
		t.Fatalf("expected place from kinship hop, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded where-answer: %q", out.Answer)
	}
}

func TestRecallGroupCompanionActivities(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-hik"] = MemoryRecord{
		MemoryID: "mem_jh", TenantID: "t-grp", SubjectID: "u1",
		Kind: KindFact, Content: "John went hiking with his colleagues",
		DedupeKey: "jh", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
	}
	store.records["j-cer"] = MemoryRecord{
		MemoryID: "mem_jc", TenantID: "t-grp", SubjectID: "u1",
		Kind: KindFact, Content: "John enjoys ceramics",
		DedupeKey: "jc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "John"},
	}
	store.records["j-jobg"] = MemoryRecord{
		MemoryID: "mem_jjg", TenantID: "t-grp", SubjectID: "u1",
		Kind: KindFact, Content: "John works as a nurse",
		DedupeKey: "jjg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_jh"},
		stubAtom{pred: PredicateActivity, val: "ceramics", memID: "mem_jc"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_jjg"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-grp", SubjectID: "u1",
		Query: "What outdoor activities has John done with his colleagues?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "hik") {
		t.Fatalf("expected colleague activity, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "ceram") {
		t.Fatalf("solo hobby crowded colleague list: %q", out.Answer)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded colleague list: %q", out.Answer)
	}
}

func TestRecallEventsPlanningForClause(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["m-bake"] = MemoryRecord{
		MemoryID: "mem_mb", TenantID: "t-ev", SubjectID: "u1",
		Kind: KindFact, Content: "Maria is planning a bake sale for the homeless shelter fundraiser",
		DedupeKey: "mb", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateEvent, "value_norm": "bake sale", "subject": "Maria"},
		Explain:  map[string]any{"predicate": PredicateEvent, "value_norm": "bake sale", "subject": "Maria"},
	}
	store.records["m-job"] = MemoryRecord{
		MemoryID: "mem_mj", TenantID: "t-ev", SubjectID: "u1",
		Kind: KindFact, Content: "Maria works as a nurse",
		DedupeKey: "mj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Maria"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Maria"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateEvent, val: "bake sale", memID: "mem_mb"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_mj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-ev", SubjectID: "u1",
		Query: "What events is Maria planning for the homeless shelter fundraiser?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "bake") {
		t.Fatalf("expected planned event, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded event list: %q", out.Answer)
	}
}

func TestRecallConsequenceHavingNotPossessionDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["a-all"] = MemoryRecord{
		MemoryID: "mem_aa", TenantID: "t-con", SubjectID: "u1",
		Kind: KindFact, Content: "Audrey got allergies from having so many dogs",
		DedupeKey: "aa", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateHealth, "value_norm": "allergies", "subject": "Audrey"},
		Explain:  map[string]any{"predicate": PredicateHealth, "value_norm": "allergies", "subject": "Audrey"},
	}
	store.records["a-dog"] = MemoryRecord{
		MemoryID: "mem_ad", TenantID: "t-con", SubjectID: "u1",
		Kind: KindFact, Content: "Audrey's dog is named Scout",
		DedupeKey: "ad", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "scout", "subject": "Audrey"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "scout", "subject": "Audrey"},
	}
	store.records["a-job"] = MemoryRecord{
		MemoryID: "mem_aj", TenantID: "t-con", SubjectID: "u1",
		Kind: KindFact, Content: "Audrey works as a nurse",
		DedupeKey: "aj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Audrey"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Audrey"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateHealth, val: "allergies", memID: "mem_aa"},
		stubAtom{pred: PredicatePossession, val: "scout", memID: "mem_ad"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_aj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-con", SubjectID: "u1",
		Query: "What did Audrey get with having so many dogs?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "allerg") {
		t.Fatalf("expected health consequence, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "scout") {
		t.Fatalf("possession crowded consequence: %q", out.Answer)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded consequence: %q", out.Answer)
	}
}

func TestRecallHowManyChildrenSkipsPartner(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-son"] = MemoryRecord{
		MemoryID: "mem_rs", TenantID: "t-kids", SubjectID: "u1",
		Kind: KindFact, Content: "Sam is Riley's son",
		DedupeKey: "rs", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "sam", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "sam", "subject": "Riley"},
	}
	store.records["r-dau"] = MemoryRecord{
		MemoryID: "mem_rd", TenantID: "t-kids", SubjectID: "u1",
		Kind: KindFact, Content: "Dana is Riley's daughter",
		DedupeKey: "rd", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Riley"},
	}
	store.records["r-par"] = MemoryRecord{
		MemoryID: "mem_rp", TenantID: "t-kids", SubjectID: "u1",
		Kind: KindFact, Content: "Alex is Riley's partner",
		DedupeKey: "rp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "alex", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "alex", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "sam", memID: "mem_rs"},
		stubAtom{pred: PredicateFamilyMember, val: "dana", memID: "mem_rd"},
		stubAtom{pred: PredicateFamilyMember, val: "alex", memID: "mem_rp"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-kids", SubjectID: "u1",
		Query: "How many children does Riley have?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.Answer) != "2" {
		t.Fatalf("expected child count 2, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
}

func TestRecallKinshipHobbiesUseDestNotSource(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["a-mom"] = MemoryRecord{
		MemoryID: "mem_am", TenantID: "t-kinh", SubjectID: "u1",
		Kind: KindFact, Content: "Dana is Alex's mother",
		DedupeKey: "am", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Alex"},
	}
	store.records["d-pot"] = MemoryRecord{
		MemoryID: "mem_dp", TenantID: "t-kinh", SubjectID: "u1",
		Kind: KindFact, Content: "Dana enjoys pottery",
		DedupeKey: "dp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "Dana"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "Dana"},
	}
	store.records["a-hik"] = MemoryRecord{
		MemoryID: "mem_ah2", TenantID: "t-kinh", SubjectID: "u1",
		Kind: KindFact, Content: "Alex enjoys hiking",
		DedupeKey: "ah2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Alex"},
	}
	store.records["a-jobk"] = MemoryRecord{
		MemoryID: "mem_ajk", TenantID: "t-kinh", SubjectID: "u1",
		Kind: KindFact, Content: "Alex works as a nurse",
		DedupeKey: "ajk", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Alex"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "dana", memID: "mem_am"},
		stubAtom{pred: PredicateActivity, val: "pottery", memID: "mem_dp"},
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_ah2"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_ajk"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-kinh", SubjectID: "u1",
		Query: "What were Alex's mother's hobbies?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "potter") {
		t.Fatalf("expected dest hobbies, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("source hobbies crowded kinship list: %q", out.Answer)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded kinship hobbies: %q", out.Answer)
	}
}

func TestRecallDualCommunityIntersectsNotUnion(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-gard"] = MemoryRecord{
		MemoryID: "mem_rg", TenantID: "t-dual", SubjectID: "u1",
		Kind: KindFact, Content: "Riley participates in the community garden",
		DedupeKey: "rg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Riley"},
	}
	store.records["c-gard"] = MemoryRecord{
		MemoryID: "mem_cg", TenantID: "t-dual", SubjectID: "u1",
		Kind: KindFact, Content: "Casey participates in the community garden",
		DedupeKey: "cg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Casey"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "community garden", "subject": "Casey"},
	}
	store.records["r-cer"] = MemoryRecord{
		MemoryID: "mem_rce", TenantID: "t-dual", SubjectID: "u1",
		Kind: KindFact, Content: "Riley participates in ceramics",
		DedupeKey: "rce", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "community garden", memID: "mem_rg"},
		stubAtom{pred: PredicateActivity, val: "community garden", memID: "mem_cg"},
		stubAtom{pred: PredicateActivity, val: "ceramics", memID: "mem_rce"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-dual", SubjectID: "u1",
		Query: "Which community activities have Riley and Casey participated in?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "garden") {
		t.Fatalf("expected shared activity, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "ceram") {
		t.Fatalf("private activity leaked into dual join: %q", out.Answer)
	}
}

func TestRecallCountSpecificPossessionHead(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["c-f1"] = MemoryRecord{
		MemoryID: "mem_cf1", TenantID: "t-fer", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a Ferrari 458",
		DedupeKey: "cf1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "ferrari 458", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "ferrari 458", "subject": "Calvin"},
	}
	store.records["c-f2"] = MemoryRecord{
		MemoryID: "mem_cf2", TenantID: "t-fer", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a Ferrari California",
		DedupeKey: "cf2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "ferrari california", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "ferrari california", "subject": "Calvin"},
	}
	store.records["c-house"] = MemoryRecord{
		MemoryID: "mem_ch", TenantID: "t-fer", SubjectID: "u1",
		Kind: KindFact, Content: "Calvin owns a cottage",
		DedupeKey: "ch", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "cottage", "subject": "Calvin"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "cottage", "subject": "Calvin"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "ferrari 458", memID: "mem_cf1"},
		stubAtom{pred: PredicatePossession, val: "ferrari california", memID: "mem_cf2"},
		stubAtom{pred: PredicatePossession, val: "cottage", memID: "mem_ch"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-fer", SubjectID: "u1",
		Query: "How many Ferraris does Calvin own?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.Answer) != "2" {
		t.Fatalf("expected Ferrari count 2, answer=%q items=%#v", out.Answer, out.Items)
	}
}

func TestRecallItemsForClauseKeepsMatchingPossessions(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["a-toy"] = MemoryRecord{
		MemoryID: "mem_at", TenantID: "t-ifor", SubjectID: "u1",
		Kind: KindFact, Content: "Audrey bought a puzzle toy for her dogs",
		DedupeKey: "at", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "puzzle toy", "subject": "Audrey"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "puzzle toy", "subject": "Audrey"},
	}
	store.records["a-couch"] = MemoryRecord{
		MemoryID: "mem_ac", TenantID: "t-ifor", SubjectID: "u1",
		Kind: KindFact, Content: "Audrey owns a couch",
		DedupeKey: "ac", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "couch", "subject": "Audrey"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "couch", "subject": "Audrey"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "puzzle toy", memID: "mem_at"},
		stubAtom{pred: PredicatePossession, val: "couch", memID: "mem_ac"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-ifor", SubjectID: "u1",
		Query: "What items has Audrey bought or made for her dogs?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "puzzle") {
		t.Fatalf("expected dog item, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "couch") {
		t.Fatalf("unrelated possession crowded for-clause list: %q", out.Answer)
	}
}

func TestRecallWhoToldAboutFromTypedHop(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["e-told"] = MemoryRecord{
		MemoryID: "mem_et", TenantID: "t-told", SubjectID: "u1",
		Kind: KindFact, Content: "Evan told Dana about his marriage",
		DedupeKey: "et", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Evan"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "dana", "subject": "Evan"},
	}
	store.records["e-job"] = MemoryRecord{
		MemoryID: "mem_ej", TenantID: "t-told", SubjectID: "u1",
		Kind: KindFact, Content: "Evan works as a nurse",
		DedupeKey: "ej", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Evan"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Evan"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "dana", memID: "mem_et"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_ej"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-told", SubjectID: "u1",
		Query: "Who did Evan tell about his marriage?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "dana") {
		t.Fatalf("expected told-to person, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded who-told: %q", out.Answer)
	}
}

func TestRecallPolarYesFromTaughtSkill(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-con"] = MemoryRecord{
		MemoryID: "mem_rcn", TenantID: "t-teach", SubjectID: "u1",
		Kind: KindFact, Content: "Riley taught herself how to play the console",
		DedupeKey: "rcn", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateSkill, "value_norm": "console", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateSkill, "value_norm": "console", "subject": "Riley"},
	}
	store.records["r-hike2"] = MemoryRecord{
		MemoryID: "mem_rh2", TenantID: "t-teach", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys hiking",
		DedupeKey: "rh2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePreference, "value_norm": "hiking", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateSkill, val: "console", memID: "mem_rcn"},
		stubAtom{pred: PredicatePreference, val: "hiking", memID: "mem_rh2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-teach", SubjectID: "u1",
		Query: "Did Riley teach herself how to play the console?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(strings.TrimSpace(out.Answer), "Yes") {
		t.Fatalf("expected Yes from taught skill, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
}

func TestRecallJourneyChangesDoNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-chg"] = MemoryRecord{
		MemoryID: "mem_rchg", TenantID: "t-jny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley faced voice changes during her journey",
		DedupeKey: "rchg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateIdentity, "value_norm": "voice changes", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateIdentity, "value_norm": "voice changes", "subject": "Riley"},
	}
	store.records["r-hikej"] = MemoryRecord{
		MemoryID: "mem_rhj", TenantID: "t-jny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys hiking",
		DedupeKey: "rhj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
	}
	store.records["r-jobj"] = MemoryRecord{
		MemoryID: "mem_rjj", TenantID: "t-jny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rjj", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateIdentity, val: "voice changes", memID: "mem_rchg"},
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_rhj"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rjj"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-jny", SubjectID: "u1",
		Query: "What are some changes Riley has faced during her journey?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "voice") {
		t.Fatalf("expected journey change, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("activity crowded journey changes: %q", out.Answer)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded journey changes: %q", out.Answer)
	}
}

func TestRecallPetNamesListDoesNotDumpOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-pet"] = MemoryRecord{
		MemoryID: "mem_rpn", TenantID: "t-pets", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's pet is named Whiskers",
		DedupeKey: "rpn", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "whiskers", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "whiskers", "subject": "Riley"},
	}
	store.records["r-jobp"] = MemoryRecord{
		MemoryID: "mem_rjp", TenantID: "t-pets", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rjp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "whiskers", memID: "mem_rpn"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rjp"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-pets", SubjectID: "u1",
		Query: "What are Riley's pets' names?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "whiskers") {
		t.Fatalf("expected pet name, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded pet names: %q", out.Answer)
	}
}
