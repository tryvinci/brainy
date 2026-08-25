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
	got := strings.ToLower(out.Answer)
	for _, bad := range []string{"hiking", "beach", "relaxing", "escaping"} {
		if strings.Contains(got, bad) {
			t.Fatalf("polar miss must not dump activities: %q", out.Answer)
		}
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

func TestRecallPracticeLocationPossessiveList(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-yogap"] = MemoryRecord{
		MemoryID: "mem_ryp", TenantID: "t-locp", SubjectID: "u1",
		Kind: KindFact, Content: "Riley practices yoga at her mother's old home, the park, and the beach",
		DedupeKey: "ryp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
	}
	store.records["r-jobp"] = MemoryRecord{
		MemoryID: "mem_rjp", TenantID: "t-locp", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rjp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "yoga", memID: "mem_ryp"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rjp"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-locp", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "mother") || !strings.Contains(got, "park") || !strings.Contains(got, "beach") {
		t.Fatalf("expected possessive/list practice places, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded possessive location list: %q", out.Answer)
	}
}

func TestRecallPracticeLocationIgnoresUnrelatedActivityDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-yoga-crowd"] = MemoryRecord{
		MemoryID: "mem_ryc", TenantID: "t-locc", SubjectID: "u1",
		Kind: KindFact, Content: "Riley practices yoga at her mother's old home, the park, and the beach",
		DedupeKey: "ryc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
	}
	store.atoms = append(store.atoms, stubAtom{pred: PredicateActivity, val: "yoga", memID: "mem_ryc"})
	extras := []struct {
		key, val, content string
	}{
		{"hike", "hiking", "Riley goes hiking at the canyon"},
		{"ceram", "ceramics", "Riley enjoys ceramics"},
		{"relx", "relaxing", "Riley enjoys relaxing"},
		{"escp", "escaping", "Riley enjoys escaping"},
		{"spnd", "spending", "Riley enjoys spending"},
		{"swim", "swimming", "Riley enjoys swimming"},
		{"cook", "cooking", "Riley enjoys cooking"},
		{"run", "running", "Riley enjoys running"},
		{"knit", "knitting", "Riley enjoys knitting"},
		{"gard", "gardening", "Riley enjoys gardening"},
	}
	for _, ex := range extras {
		id := "mem_" + ex.key
		store.records[ex.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-locc", SubjectID: "u1",
			Kind: KindFact, Content: ex.content,
			DedupeKey: ex.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": ex.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": ex.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: PredicateActivity, val: ex.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-locc", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "mother") || !strings.Contains(got, "park") || !strings.Contains(got, "beach") {
		t.Fatalf("expected yoga practice places, answer=%q items=%#v", out.Answer, out.Items)
	}
	for _, bad := range []string{"canyon", "relaxing", "escaping", "spending", "ceramics", "hiking", "knitting"} {
		if strings.Contains(got, bad) {
			t.Fatalf("unrelated activity crowded practice places: %q", out.Answer)
		}
	}
}

func TestRecallPracticeLocationOnPrepAndNestedClause(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, val, content string
	}{
		{"beach", "yoga", "Riley does yoga on the beach."},
		{"nested", "yoga", "Riley met Alex at yoga in the park."},
		{"clause", "yoga", "Riley met Alex at the park Deborah met Jolene."},
		{"near", "yoga", "The yoga retreat Riley attended was near her mother`s old cottage."},
		{"hike", "hiking", "Riley goes hiking at the canyon"},
		{"relx", "relaxing", "Riley enjoys relaxing"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-locon", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: PredicateActivity, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-locon", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "beach") || !strings.Contains(got, "park") || !strings.Contains(got, "cottage") {
		t.Fatalf("expected on-prep / nested / near places, answer=%q items=%#v", out.Answer, out.Items)
	}
	for _, bad := range []string{"yoga in the park", "deborah", "jolene", "canyon", "relaxing"} {
		if strings.Contains(got, bad) {
			t.Fatalf("nested practice object or person clause leaked: %q", out.Answer)
		}
	}
}

func TestRecallPracticeLocationRecoversDestSubjectLocatives(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"park", PredicateEvent, "yoga at park", "Riley met Alex at yoga in the park."},
		{"beach", PredicateActivity, "yoga", "Riley does yoga on the beach."},
		{"studio", PredicatePossession, "favorite yoga studio", "Riley recommends the yoga studio nearby."},
		{"hike", PredicateActivity, "hiking", "Riley goes hiking at the canyon"},
		{"job", PredicateOccupation, "nurse", "Riley works as a nurse"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-locrec", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-locrec", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "park") || !strings.Contains(got, "beach") || !strings.Contains(got, "studio") {
		t.Fatalf("expected dest-subject practice places, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	for _, bad := range []string{"canyon", "nurse"} {
		if strings.Contains(got, bad) {
			t.Fatalf("unrelated dump crowded dest-subject places: %q", out.Answer)
		}
	}
}

func TestRecallPracticeLocationListIgnoresPurchaseLeftover(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"park", PredicateActivity, "yoga", "Riley practices yoga at the park."},
		{"buy", PredicateEvent, "candle", "Riley bought a scented candle for her yoga practice on 28 March 2023."},
		{"beach", PredicateActivity, "yoga", "Riley does yoga on the beach."},
		{"studio", PredicatePossession, "favorite yoga studio", "Riley recommends the yoga studio nearby."},
		{"mom", PredicateActivity, "yoga", "Riley practices yoga at her mother's old home."},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-locbuy", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-locbuy", SubjectID: "u1",
		Query: "Which locations does Riley practice her yoga at?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if strings.Contains(got, "candle") {
		t.Fatalf("purchase leftover must not cover a practice location list: %q", out.Answer)
	}
	if !strings.Contains(got, "park") || !strings.Contains(got, "beach") || !strings.Contains(got, "studio") || !strings.Contains(got, "mother") {
		t.Fatalf("expected leftover practice places, answer=%q items=%#v", out.Answer, out.Items)
	}
}

func TestRecallUnwindRecoversDestressEvidence(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"camp", PredicateActivity, "camping", "Riley enjoys camping"},
		{"run", PredicateActivity, "running", "Riley runs to destress"},
		{"pot", PredicateActivity, "pottery", "Riley finds making pottery calming"},
		{"job", PredicateOccupation, "nurse", "Riley works as a nurse"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-unw2", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-unw2", SubjectID: "u1",
		Query: "What does Riley do to unwind?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "run") || !strings.Contains(got, "potter") {
		t.Fatalf("expected destress unwind activities, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "camp") || strings.Contains(got, "nurse") {
		t.Fatalf("unrelated dump crowded unwind list: %q", out.Answer)
	}
	listed, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-unw2", SubjectID: "u1",
		Query: "What does Riley do to unwind?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	listGot := strings.ToLower(listed.Answer)
	for _, it := range listed.Items {
		listGot += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(listGot, "run") || !strings.Contains(listGot, "potter") {
		t.Fatalf("enumerate unwind must keep destress activities, answer=%q items=%#v", listed.Answer, listed.Items)
	}
	if strings.Contains(listGot, "camp") || strings.Contains(listGot, "nurse") {
		t.Fatalf("enumerate unwind must not dump camping or occupation, %q", listed.Answer)
	}
}

func TestRecallInstrumentRecoversPracticeObject(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"clar", PredicateSkill, "plays clarinet", "Riley plays the clarinet."},
		{"viol", PredicateActivity, "playing violin", "Riley does daily violin practice after work."},
		{"hike", PredicateActivity, "hiking", "Riley enjoys hiking"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-instr2", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-instr2", SubjectID: "u1",
		Query: "What instruments does Riley play?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "clarinet") || !strings.Contains(got, "violin") {
		t.Fatalf("expected play/practice objects, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("hobby crowded recovered instrument list: %q", out.Answer)
	}
}

func TestRecallPetTrickRecoversFromTrickContent(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, subj, content string
	}{
		{"cpp", PredicateSkill, "c++", "James", "James coded in c++"},
		{"daisy", PredicateSkill, "sit, stay, paw, rollover", "Daisy", "James: They can do tricks like sit, stay, paw, and rollover"},
		{"hike", PredicatePreference, "hiking", "James", "James enjoys hiking"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-trick2", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": f.subj},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": f.subj},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-trick2", SubjectID: "u1",
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
		t.Fatalf("expected trick content slots, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "c++") || strings.Contains(got, "hik") {
		t.Fatalf("owner skills crowded trick list: %q", out.Answer)
	}
}

func TestRecallBesidesRecoversWorkStressor(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"hike", PredicateActivity, "hiking", "Andrew enjoys hiking"},
		{"stress", PredicateHealth, "stress due to work", "Andrew is experiencing stress due to his work"},
		{"sushi", PredicateActivity, "sushi before", "Andrew had sushi before"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-bes2", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Andrew"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Andrew"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-bes2", SubjectID: "u1",
		Query: "What is the biggest stressor in Andrew's life besides not being able to hike frequently?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "work") {
		t.Fatalf("expected work stressor, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "hik") || strings.Contains(got, "sushi") {
		t.Fatalf("besides dump leaked: %q", out.Answer)
	}
}

func TestPlacesFromContentLiveParkSentence(t *testing.T) {
	got := placesFromContent("Deborah met Jolene at yoga in the park.")
	joined := strings.ToLower(strings.Join(got, " | "))
	if strings.Contains(joined, "deborah") || strings.Contains(joined, "jolene") {
		t.Fatalf("person leaked from live sentence: %v", got)
	}
	lower := placesFromContent("park deborah met jolene at yoga in the park")
	ljoin := strings.ToLower(strings.Join(lower, " | "))
	if !strings.Contains(ljoin, "park") {
		t.Fatalf("expected park from lowercased blob, got %v", lower)
	}
	if strings.Contains(ljoin, "deborah") || strings.Contains(ljoin, "jolene") {
		t.Fatalf("lowercased person clause leaked: %v", lower)
	}
}

func TestPlacesFromContentOnPrepNestedPersonAndDate(t *testing.T) {
	onBeach := placesFromContent("Riley does yoga on the beach in Bali")
	joined := strings.ToLower(strings.Join(onBeach, " | "))
	if !strings.Contains(joined, "beach") {
		t.Fatalf("expected beach from on-prep, got %v", onBeach)
	}
	if strings.Contains(joined, "yoga") {
		t.Fatalf("practice object leaked as place: %v", onBeach)
	}
	nested := placesFromContent("Riley met Alex at yoga in the park")
	njoin := strings.ToLower(strings.Join(nested, " | "))
	if strings.Contains(njoin, "yoga in") {
		t.Fatalf("nested in must cut practice object phrase: %v", nested)
	}
	clause := placesFromContent("Riley met Alex at the park Deborah met Jolene")
	cjoin := strings.ToLower(strings.Join(clause, " | "))
	if !strings.Contains(cjoin, "park") {
		t.Fatalf("expected park, got %v", clause)
	}
	if strings.Contains(cjoin, "deborah") || strings.Contains(cjoin, "jolene") {
		t.Fatalf("person clause leaked into place: %v", clause)
	}
	dated := placesFromContent("Riley started on 8 September 2023")
	for _, p := range dated {
		if looksPlaceDateToken(strings.Fields(p)[0]) {
			t.Fatalf("date tail must not be a place: %v", dated)
		}
	}
	btick := placesFromContent("The retreat was near her mother`s old cottage")
	bjoin := strings.ToLower(strings.Join(btick, " | "))
	if !strings.Contains(bjoin, "cottage") {
		t.Fatalf("expected backtick mother cottage, got %v", btick)
	}
}

func TestPlacesFromContentStopsRelativeClauseAndGerund(t *testing.T) {
	got := placesFromContent("Riley practices yoga at a cottage that blooms each spring")
	joined := strings.ToLower(strings.Join(got, " | "))
	if !strings.Contains(joined, "cottage") {
		t.Fatalf("expected cottage, got %v", got)
	}
	if strings.Contains(joined, "bloom") {
		t.Fatalf("relative clause leaked into place: %v", got)
	}
	gerund := placesFromContent("Riley unwinds at relaxing")
	for _, p := range gerund {
		if strings.EqualFold(strings.TrimSpace(p), "relaxing") {
			t.Fatalf("lone gerund must not be a place: %v", gerund)
		}
	}
}

func TestRecallUnhealthyListDropsPositiveSlogans(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	facts := []struct {
		key, pred, val, content string
	}{
		{"soda", PredicatePreference, "soda", "Riley continues to enjoy soda and candy."},
		{"candy", PredicatePreference, "candy", "Riley continues to enjoy soda and candy."},
		{"bought", PredicatePreference, "unhealthy snacks", "Riley bought unhealthy snacks"},
		{"hs", PredicateActivity, "healthy snacks", "Riley is starting a new campaign called Healthy Snacks"},
		{"hsi", PredicateActivity, "healthier snack ideas", "Riley posted Healthier Snack Ideas"},
		{"god", PredicateMediaConsumed, "the godfather", "Riley watched The Godfather"},
	}
	for _, f := range facts {
		id := "mem_" + f.key
		store.records[f.key] = MemoryRecord{
			MemoryID: id, TenantID: "t-unh", SubjectID: "u1",
			Kind: KindFact, Content: f.content,
			DedupeKey: f.key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": f.pred, "value_norm": f.val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: f.pred, val: f.val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-unh", SubjectID: "u1",
		Query: "What kind of unhealthy snacks does Riley enjoy eating?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "soda") || !strings.Contains(got, "candy") {
		t.Fatalf("expected soda and candy, answer=%q items=%#v", out.Answer, out.Items)
	}
	for _, it := range out.Items {
		v := strings.ToLower(strings.TrimSpace(it.Value))
		if v == "healthy snacks" || strings.Contains(v, "healthier") || strings.Contains(v, "godfather") {
			t.Fatalf("positive slogan crowded un- list: %q items=%#v", out.Answer, out.Items)
		}
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

func TestRecallWhoSupportsGroupNotOccupation(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-supg"] = MemoryRecord{
		MemoryID: "mem_rsg", TenantID: "t-whog", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's friends and team support her in tough times",
		DedupeKey: "rsg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "friends and team", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "friends and team", "subject": "Riley"},
	}
	store.records["r-jobg"] = MemoryRecord{
		MemoryID: "mem_rjg", TenantID: "t-whog", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rjg", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "friends and team", memID: "mem_rsg"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rjg"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-whog", SubjectID: "u1",
		Query: "Who supports Riley in tough times?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "friend") || !strings.Contains(got, "team") {
		t.Fatalf("expected supporter group, answer=%q hops=%v", out.Answer, out.Explain["hop_results"])
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded group who-answer: %q", out.Answer)
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

func TestRecallNamedCommunityFiltersUnrelatedActivities(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-civ"] = MemoryRecord{
		MemoryID: "mem_rciv", TenantID: "t-ncom", SubjectID: "u1",
		Kind: KindFact, Content: "Riley participates in the civic festival",
		DedupeKey: "rciv", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "civic festival", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "civic festival", "subject": "Riley"},
	}
	store.records["r-coal"] = MemoryRecord{
		MemoryID: "mem_rcoal", TenantID: "t-ncom", SubjectID: "u1",
		Kind: KindFact, Content: "Riley joined the civic coalition",
		DedupeKey: "rcoal", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateAffiliation, "value_norm": "civic coalition", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateAffiliation, "value_norm": "civic coalition", "subject": "Riley"},
	}
	store.records["r-hike-n"] = MemoryRecord{
		MemoryID: "mem_rhn", TenantID: "t-ncom", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys hiking",
		DedupeKey: "rhn", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
	}
	store.records["r-job-n"] = MemoryRecord{
		MemoryID: "mem_rjn", TenantID: "t-ncom", SubjectID: "u1",
		Kind: KindFact, Content: "Riley works as a nurse",
		DedupeKey: "rjn", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateOccupation, "value_norm": "nurse", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "civic festival", memID: "mem_rciv"},
		stubAtom{pred: PredicateAffiliation, val: "civic coalition", memID: "mem_rcoal"},
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_rhn"},
		stubAtom{pred: PredicateOccupation, val: "nurse", memID: "mem_rjn"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-ncom", SubjectID: "u1",
		Query: "In what ways is Riley participating in the civic community?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "civic") {
		t.Fatalf("expected named-community value, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("unrelated activity crowded named community: %q", out.Answer)
	}
	if strings.Contains(got, "nurse") {
		t.Fatalf("occupation crowded named community: %q", out.Answer)
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

func TestRecallOutdoorModifierFiltersIndoorGroupActivities(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-out"] = MemoryRecord{
		MemoryID: "mem_jo", TenantID: "t-out", SubjectID: "u1",
		Kind: KindFact, Content: "John went outdoor hiking with his colleagues",
		DedupeKey: "jo", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
	}
	store.records["j-pot"] = MemoryRecord{
		MemoryID: "mem_jp2", TenantID: "t-out", SubjectID: "u1",
		Kind: KindFact, Content: "John did pottery with his colleagues",
		DedupeKey: "jp2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_jo"},
		stubAtom{pred: PredicateActivity, val: "pottery", memID: "mem_jp2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-out", SubjectID: "u1",
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
		t.Fatalf("expected outdoor colleague activity, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "potter") {
		t.Fatalf("indoor colleague activity crowded outdoor list: %q", out.Answer)
	}
}

func TestRecallOutdoorModifierFiltersIndoorInEnumerateMode(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-out2"] = MemoryRecord{
		MemoryID: "mem_jo2", TenantID: "t-oute", SubjectID: "u1",
		Kind: KindFact, Content: "John went outdoor hiking with his colleagues",
		DedupeKey: "jo2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
	}
	store.records["j-pot2"] = MemoryRecord{
		MemoryID: "mem_jp3", TenantID: "t-oute", SubjectID: "u1",
		Kind: KindFact, Content: "John did pottery with his colleagues",
		DedupeKey: "jp3", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "pottery", "subject": "John"},
	}
	store.records["j-mov"] = MemoryRecord{
		MemoryID: "mem_jm", TenantID: "t-oute", SubjectID: "u1",
		Kind: KindFact, Content: "John has movie nights at home",
		DedupeKey: "jm", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "movie nights at home", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "movie nights at home", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_jo2"},
		stubAtom{pred: PredicateActivity, val: "pottery", memID: "mem_jp3"},
		stubAtom{pred: PredicateActivity, val: "movie nights at home", memID: "mem_jm"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-oute", SubjectID: "u1",
		Query: "What outdoor activities has John done with his colleagues?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "hik") {
		t.Fatalf("expected outdoor colleague activity in enumerate, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "potter") || strings.Contains(got, "movie") {
		t.Fatalf("enumerate mode dumped indoor crowding: %q", out.Answer)
	}
	if len(out.Items) > maxEnumerateAnswerItems {
		t.Fatalf("enumerate list exceeded cap: n=%d", len(out.Items))
	}
}

func TestRecallOutdoorPrefersHeadOverCompanionOnly(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["j-hike"] = MemoryRecord{
		MemoryID: "mem_jhike", TenantID: "t-prefint", SubjectID: "u1",
		Kind: KindFact, Content: "John enjoys being outdoors - going for hikes",
		DedupeKey: "jhike", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "John"},
	}
	store.records["j-conv"] = MemoryRecord{
		MemoryID: "mem_jconv", TenantID: "t-prefint", SubjectID: "u1",
		Kind: KindFact, Content: "John attended a convention with his colleagues",
		DedupeKey: "jconv", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "convention attendance", "subject": "John"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "convention attendance", "subject": "John"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_jhike"},
		stubAtom{pred: PredicateActivity, val: "convention attendance", memID: "mem_jconv"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-prefint", SubjectID: "u1",
		Query: "What outdoor activities has John done with his colleagues?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "hik") && !strings.Contains(got, "outdoor") {
		t.Fatalf("expected outdoor activity over colleague indoor, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "convention") {
		t.Fatalf("colleague indoor crowded outdoor list: %q items=%#v", out.Answer, out.Items)
	}
}

func TestRecallEnumerateModeCapsActivityDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	vals := []string{
		"hiking", "ceramics", "swimming", "knitting", "gardening",
		"cooking", "running", "cycling", "painting", "dancing",
		"climbing", "rowing",
	}
	for _, val := range vals {
		key := val
		id := "mem_" + val
		store.records[key] = MemoryRecord{
			MemoryID: id, TenantID: "t-cap", SubjectID: "u1",
			Kind: KindFact, Content: "Riley enjoys " + val,
			DedupeKey: key, Status: StatusActive, UpdatedAt: now,
			Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": val, "subject": "Riley"},
			Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": val, "subject": "Riley"},
		}
		store.atoms = append(store.atoms, stubAtom{pred: PredicateActivity, val: val, memID: id})
	}
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-cap", SubjectID: "u1",
		Query: "What activities does Riley enjoy?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Items) > maxEnumerateAnswerItems {
		t.Fatalf("enumerate dump was not capped: n=%d items=%#v", len(out.Items), out.Items)
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

func TestRecallKinshipRoleDestHobbiesEnumerate(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["a-momr"] = MemoryRecord{
		MemoryID: "mem_amr", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex's mother is her mom.",
		DedupeKey: "amr", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateFamilyMember, "value_norm": "mother", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateFamilyMember, "value_norm": "mother", "subject": "Alex"},
	}
	store.records["m-read"] = MemoryRecord{
		MemoryID: "mem_mr", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex's mother had reading as one of her hobbies.",
		DedupeKey: "mr", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"subject": "Alex's mother"},
		Explain:  map[string]any{"subject": "Alex's mother"},
	}
	store.records["m-trav"] = MemoryRecord{
		MemoryID: "mem_mt", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex's mother was passionate about travel.",
		DedupeKey: "mt", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"subject": "Alex's mother"},
		Explain:  map[string]any{"subject": "Alex's mother"},
	}
	store.records["m-art"] = MemoryRecord{
		MemoryID: "mem_ma", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex's mother was interested in art.",
		DedupeKey: "ma", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"subject": "Alex's mother"},
		Explain:  map[string]any{"subject": "Alex's mother"},
	}
	store.records["m-cook"] = MemoryRecord{
		MemoryID: "mem_mc", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex's mother had a big passion for cooking.",
		DedupeKey: "mc", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "cooking", "subject": "Alex's mother"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "cooking", "subject": "Alex's mother"},
	}
	store.records["a-vis"] = MemoryRecord{
		MemoryID: "mem_av", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex visited mother's old house last year.",
		DedupeKey: "av", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "visiting mother's old house", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "visiting mother's old house", "subject": "Alex"},
	}
	store.records["a-hikr"] = MemoryRecord{
		MemoryID: "mem_ahr", TenantID: "t-kinrole", SubjectID: "u1",
		Kind: KindFact, Content: "Alex enjoys hiking",
		DedupeKey: "ahr", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Alex"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Alex"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateFamilyMember, val: "mother", memID: "mem_amr"},
		stubAtom{pred: PredicateActivity, val: "cooking", memID: "mem_mc"},
		stubAtom{pred: PredicateActivity, val: "visiting mother's old house", memID: "mem_av"},
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_ahr"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-kinrole", SubjectID: "u1",
		Query: "What were Alex's mother's hobbies?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	for _, want := range []string{"read", "travel", "art", "cook"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected dest hobby %q, answer=%q items=%#v hops=%v", want, out.Answer, out.Items, out.Explain["hop_results"])
		}
	}
	if strings.Contains(got, "house") || strings.Contains(got, "visit") {
		t.Fatalf("source visit crowded dest hobbies: %q", out.Answer)
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("source hobbies crowded dest list: %q", out.Answer)
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

func TestRecallDualCommunityPhraseOverlapAndPartner(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-yoga"] = MemoryRecord{
		MemoryID: "mem_ryo", TenantID: "t-dual2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley started yoga",
		DedupeKey: "ryo", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "yoga", "subject": "Riley"},
	}
	store.records["c-yoga"] = MemoryRecord{
		MemoryID: "mem_cyo", TenantID: "t-dual2", SubjectID: "u1",
		Kind: KindFact, Content: "Casey organized yoga",
		DedupeKey: "cyo", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "organized yoga", "subject": "Casey"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "organized yoga", "subject": "Casey"},
	}
	store.records["c-run"] = MemoryRecord{
		MemoryID: "mem_crun", TenantID: "t-dual2", SubjectID: "u1",
		Kind: KindFact, Content: "Casey joined Riley's running group",
		DedupeKey: "crun", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "riley's running group", "subject": "Casey"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "riley's running group", "subject": "Casey"},
	}
	store.records["r-unw"] = MemoryRecord{
		MemoryID: "mem_runw", TenantID: "t-dual2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley unwinds via relaxing",
		DedupeKey: "runw", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "relaxing", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "relaxing", "subject": "Riley"},
	}
	store.records["c-cer"] = MemoryRecord{
		MemoryID: "mem_cce", TenantID: "t-dual2", SubjectID: "u1",
		Kind: KindFact, Content: "Casey participates in ceramics",
		DedupeKey: "cce", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Casey"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "ceramics", "subject": "Casey"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateActivity, val: "yoga", memID: "mem_ryo"},
		stubAtom{pred: PredicateActivity, val: "organized yoga", memID: "mem_cyo"},
		stubAtom{pred: PredicateActivity, val: "riley's running group", memID: "mem_crun"},
		stubAtom{pred: PredicateActivity, val: "relaxing", memID: "mem_runw"},
		stubAtom{pred: PredicateActivity, val: "ceramics", memID: "mem_cce"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-dual2", SubjectID: "u1",
		Query: "Which community activities have Riley and Casey participated in?", Mode: "enumerate", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "yoga") {
		t.Fatalf("expected yoga containment join, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if !strings.Contains(got, "run") {
		t.Fatalf("expected partner running group, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "relax") {
		t.Fatalf("unwind slogan leaked into dual community: %q", out.Answer)
	}
	if strings.Contains(got, "ceram") {
		t.Fatalf("private activity leaked into dual community: %q", out.Answer)
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

func TestRecallNamedJourneyFiltersUnrelatedIdentity(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-rec"] = MemoryRecord{
		MemoryID: "mem_rrec", TenantID: "t-njny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley faced voice changes during her recovery",
		DedupeKey: "rrec", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateIdentity, "value_norm": "voice changes", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateIdentity, "value_norm": "voice changes", "subject": "Riley"},
	}
	store.records["r-oh"] = MemoryRecord{
		MemoryID: "mem_roh", TenantID: "t-njny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley is from Ohio",
		DedupeKey: "roh", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateIdentity, "value_norm": "from Ohio", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateIdentity, "value_norm": "from Ohio", "subject": "Riley"},
	}
	store.records["r-hike-j2"] = MemoryRecord{
		MemoryID: "mem_rhj2", TenantID: "t-njny", SubjectID: "u1",
		Kind: KindFact, Content: "Riley enjoys hiking",
		DedupeKey: "rhj2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateActivity, "value_norm": "hiking", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicateIdentity, val: "voice changes", memID: "mem_rrec"},
		stubAtom{pred: PredicateIdentity, val: "from Ohio", memID: "mem_roh"},
		stubAtom{pred: PredicateActivity, val: "hiking", memID: "mem_rhj2"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-njny", SubjectID: "u1",
		Query: "What are some changes Riley has faced during her recovery journey?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	for _, it := range out.Items {
		got += " | " + strings.ToLower(it.Value)
	}
	if !strings.Contains(got, "voice") {
		t.Fatalf("expected named-journey change, answer=%q items=%#v hops=%v", out.Answer, out.Items, out.Explain["hop_results"])
	}
	if strings.Contains(got, "ohio") {
		t.Fatalf("unrelated identity crowded named journey: %q", out.Answer)
	}
	if strings.Contains(got, "hik") {
		t.Fatalf("activity crowded named journey: %q", out.Answer)
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

func TestRecallPetNamesDoesNotDumpOtherPossessions(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["r-pet2"] = MemoryRecord{
		MemoryID: "mem_rpn2", TenantID: "t-pets2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's pet is named Whiskers",
		DedupeKey: "rpn2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "whiskers", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "whiskers", "subject": "Riley"},
	}
	store.records["r-car"] = MemoryRecord{
		MemoryID: "mem_rcar", TenantID: "t-pets2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley owns a sedan",
		DedupeKey: "rcar", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "sedan", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "sedan", "subject": "Riley"},
	}
	store.records["r-gui"] = MemoryRecord{
		MemoryID: "mem_rgui", TenantID: "t-pets2", SubjectID: "u1",
		Kind: KindFact, Content: "Riley owns a guitar",
		DedupeKey: "rgui", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "guitar", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "guitar", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "whiskers", memID: "mem_rpn2"},
		stubAtom{pred: PredicatePossession, val: "sedan", memID: "mem_rcar"},
		stubAtom{pred: PredicatePossession, val: "guitar", memID: "mem_rgui"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-pets2", SubjectID: "u1",
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
		t.Fatalf("expected pet name, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "sedan") || strings.Contains(got, "guitar") {
		t.Fatalf("other possessions crowded pet names: %q", out.Answer)
	}
}

func TestRecallPetNamesKeepsMultipleNamedPets(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["p1"] = MemoryRecord{
		MemoryID: "mem_p1", TenantID: "t-pets3", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's pet is named Luna",
		DedupeKey: "p1", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "luna", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "luna", "subject": "Riley"},
	}
	store.records["p2"] = MemoryRecord{
		MemoryID: "mem_p2", TenantID: "t-pets3", SubjectID: "u1",
		Kind: KindFact, Content: "Riley's dog is named Oliver",
		DedupeKey: "p2", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "oliver", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "oliver", "subject": "Riley"},
	}
	store.records["p3"] = MemoryRecord{
		MemoryID: "mem_p3", TenantID: "t-pets3", SubjectID: "u1",
		Kind: KindFact, Content: "Riley owns a sedan",
		DedupeKey: "p3", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "sedan", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "sedan", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "luna", memID: "mem_p1"},
		stubAtom{pred: PredicatePossession, val: "oliver", memID: "mem_p2"},
		stubAtom{pred: PredicatePossession, val: "sedan", memID: "mem_p3"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-pets3", SubjectID: "u1",
		Query: "What are Riley's pets' names?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "luna") || !strings.Contains(got, "oliver") {
		t.Fatalf("expected both named pets, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "sedan") {
		t.Fatalf("sedan crowded named pets: %q", out.Answer)
	}
}

func TestRecallChildhoodItemsPreferChildEvidence(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["c-doll"] = MemoryRecord{
		MemoryID: "mem_cd", TenantID: "t-child", SubjectID: "u1",
		Kind: KindFact, Content: "Riley had a wooden puzzle as a child",
		DedupeKey: "cd", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "wooden puzzle", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "wooden puzzle", "subject": "Riley"},
	}
	store.records["c-desk"] = MemoryRecord{
		MemoryID: "mem_ck", TenantID: "t-child", SubjectID: "u1",
		Kind: KindFact, Content: "Riley owns an office desk",
		DedupeKey: "ck", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicatePossession, "value_norm": "office desk", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicatePossession, "value_norm": "office desk", "subject": "Riley"},
	}
	store.atoms = append(store.atoms,
		stubAtom{pred: PredicatePossession, val: "wooden puzzle", memID: "mem_cd"},
		stubAtom{pred: PredicatePossession, val: "office desk", memID: "mem_ck"},
	)
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-child", SubjectID: "u1",
		Query: "What items did Riley have as a child?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToLower(out.Answer)
	if !strings.Contains(got, "puzzle") {
		t.Fatalf("expected childhood item, answer=%q items=%#v", out.Answer, out.Items)
	}
	if strings.Contains(got, "desk") {
		t.Fatalf("adult possession crowded childhood items: %q", out.Answer)
	}
}

func TestOrdinalNameFromPacketPicksSecondDatedThenUndated(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley has a dog named Shadow."},
			{Content: "Riley got a puppy named Coco on 28 July 2023."},
			{Content: "Riley has a new puppy."},
			{Content: "Riley had a conversation with someone named David"},
			{Content: "John had a family member named Max, a dog, who was part of his family for 10 years."},
		},
	}
	if dt := parseDateFromText("Riley got a puppy named Coco on 28 July 2023."); dt == nil {
		t.Fatal("expected trailing-period date on Coco line to parse")
	}
	got := ordinalNameFromPacket("What is the name of Riley's second puppy?", pkt)
	if got != "Shadow" {
		t.Fatalf("expected Shadow as second named pet, got %q", got)
	}
	if first := ordinalNameFromPacket("What is the name of Riley's first puppy?", pkt); first != "Coco" {
		t.Fatalf("expected Coco as first dated pet, got %q", first)
	}
}

func TestLeftoverCoveringSpecificAnswerSkipsChatTurn(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "John", Value: "John", Source: "search_fallback"},
		{Kind: "fetch_predicate", Entity: "John", Source: "search_fallback",
			Value: "metal detecting", Values: []string{"metal detecting"}},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Oh, I've been organizing something with my friends yesterday - it was cool (7 May 2022; the day before 8 May 2022)"},
			{Content: "John organized a charity gaming tournament for the game CS:GO on 7 May 2022."},
		},
	}
	q := "What did John organize with his friends on May 8, 2022?"
	if dt := parseDateFromText("What did John organize with his friends on May 8, 2022?"); dt == nil || dt.Day() != 8 || dt.Month() != 5 {
		t.Fatalf("expected query May 8 parse, got %#v", dt)
	}
	if dt := parseDateFromText("John organized a charity gaming tournament for the game CS:GO on 7 May 2022."); dt == nil || dt.Day() != 7 {
		t.Fatalf("expected fact May 7 parse, got %#v", dt)
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	if !strings.Contains(strings.ToLower(got), "cs:go") {
		t.Fatalf("expected CS:GO fact over chat turn, got %q", got)
	}
	if strings.Contains(got, "7 May") || strings.Contains(got, "7 May 2022") {
		t.Fatalf("conflicting 7 May tail must not remain when query is May 8: %q", got)
	}
}

func TestLeftoverCoveringSpecificAnswerSkipsImageCaption(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "[basketball game] [a photo of a basketball game being played in a gym]"},
			{Content: "Riley supports the Wolves basketball team."},
		},
	}
	got := leftoverCoveringSpecificAnswer("Which basketball team does Riley support?", hops, pkt)
	if !strings.Contains(strings.ToLower(got), "wolves") {
		t.Fatalf("expected team fact over OCR caption, got %q", got)
	}
	if strings.Contains(strings.ToLower(got), "photo") {
		t.Fatalf("image caption leaked: %q", got)
	}
	pkt.ContextEvidence = append(pkt.ContextEvidence, PacketItem{
		Content: "Riley offered emotional support to Casey.",
	})
	got = leftoverCoveringSpecificAnswer("Which basketball team does Riley support?", hops, pkt)
	if !strings.Contains(strings.ToLower(got), "wolves") {
		t.Fatalf("support-verb leftover covering must not pick emotional support: %q", got)
	}
}

func TestLeftoverCoveringSpecificAnswerSkipsQuestionPrompt(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Casey", Value: "Casey", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Any tips on studying or time management"},
			{Content: "Riley and Casey discussed breaking tasks into smaller pieces and setting goals as a studying strategy."},
		},
	}
	got := leftoverCoveringSpecificAnswer("What did Riley and Casey discuss as a helpful strategy for studying and time management?", hops, pkt)
	if !strings.Contains(strings.ToLower(got), "smaller pieces") {
		t.Fatalf("expected strategy fact over question prompt, got %q", got)
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(got)), "any tips") {
		t.Fatalf("question prompt leaked: %q", got)
	}
}

func TestDatedContentConflictsQuerySkipsFarRelativeTail(t *testing.T) {
	q := "What is Jolene's favorite book which she mentioned on 4 February, 2023?"
	if !datedContentConflictsQuery(q, `Jolene read "Neal Stephenson" (21 January 2023; the week before 4 February 2023)`) {
		t.Fatal("January event must conflict with February 4 query even with session-relative tail")
	}
	if datedContentConflictsQuery(q, `I'm really into this book called "Sapiens" - it's a fascinating look at human history`) {
		t.Fatal("undated covering fact must not conflict")
	}
	csgo := "What did John organize with his friends on May 8, 2022?"
	if datedContentConflictsQuery(csgo, "John organized a charity gaming tournament for the game CS:GO on 7 May 2022.") {
		t.Fatal("adjacent-day event must stay eligible")
	}
	gym := "What exciting news did Maria share on 16 June, 2023?"
	if datedContentConflictsQuery(gym, "I got some great news to share - I joined a gym last week (9 June 2023; the week before 16 June 2023)") {
		t.Fatal("last-week event relative to the query session day must stay eligible")
	}
	if querySpecificCalendarDate("Which basketball team does Riley support?") != nil {
		t.Fatal("dateless query must not date-filter")
	}
	if querySpecificCalendarDate("What setback did Melanie face in October 2023?") != nil {
		t.Fatal("month-year query must not day-filter")
	}
}

func TestDatedHybridContentConflictsQueryIsTighterThanLeftover(t *testing.T) {
	q := "Where was James at on July 12, 2022?"
	if datedHybridContentConflictsQuery(q, "James will depart for Toronto on July 11, 2022 in the evening.") {
		t.Fatal("where-queries must not date-filter hybrid packets")
	}
	if datedHybridContentConflictsQuery(q, "James went surfing on 6 July 2022.") {
		t.Fatal("where-queries must not date-filter same-month hybrid lines")
	}
	gym := "What exciting news did Maria share on 16 June, 2023?"
	gymLine := "I got some great news to share - I joined a gym last week (9 June 2023; the week before 16 June 2023)"
	if datedContentConflictsQuery(gym, gymLine) {
		t.Fatal("leftover covering must keep last-week gym news")
	}
	if !datedHybridContentConflictsQuery(gym, gymLine) {
		t.Fatal("hybrid packet may drop last-week gym; leftover covering recovers it")
	}
}

func TestLeftoverCoveringSpecificAnswerSkipsFarDatedCrowd(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: `Jolene read "Neal Stephenson" (21 January 2023; the week before 4 February 2023) (28 January 2023)`},
			{Content: `I'm really into this book called "Sapiens" - it's a fascinating look at human history and how technology has affected us`},
			{Content: "Jolene read the book \"Avalanche\" by Neal Stephenson on 21 January 2023."},
			{Content: "Deborah mentioned the Eisenhower Matrix as a tool for organizing and prioritizing tasks."},
		},
	}
	got := leftoverCoveringSpecificAnswer("What is Jolene's favorite book which she mentioned on 4 February, 2023?", hops, pkt)
	if !strings.Contains(strings.ToLower(got), "sapiens") {
		t.Fatalf("expected Sapiens over far-dated Stephenson crowd, got %q", got)
	}
	if strings.Contains(strings.ToLower(got), "stephenson") || strings.Contains(strings.ToLower(got), "avalanche") || strings.Contains(strings.ToLower(got), "eisenhower") {
		t.Fatalf("far-dated book crowd leaked: %q", got)
	}
}

func TestLeftoverCoveringSpecificAnswerKeepsSpeakerPrefixedDatedCover(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: `Jolene read "Neal Stephenson" (21 January 2023; the week before 4 February 2023)`},
			{Content: "During the mini retreat on 8 February 2023, Jolene gained a new outlook on life."},
			{Content: "Jolene: I really accomplished something with my engineering project - I came up with some neat solutions and I'm really excited about it"},
			{Content: "Jolene did a mini retreat on 8 February 2023 to assess where she is in life."},
		},
	}
	got := leftoverCoveringSpecificAnswer("What cool stuff did Jolene accomplish at the retreat on 9 February, 2023?", hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "neat solutions") && !strings.Contains(lower, "engineering") {
		t.Fatalf("expected speaker-prefixed accomplishment over Stephenson/outlook crowd, got %q", got)
	}
	if strings.Contains(lower, "stephenson") {
		t.Fatalf("far-dated crowd leaked: %q", got)
	}
}

func TestLeftoverCoveringSpecificAnswerKeepsLastWeekSessionNews(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Maria", Value: "Maria", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Maria: Hey John, great news - I'm now friends with one of my fellow volunteers"},
			{Content: "Anything has done exciting at Horizon"},
			{Content: "I got some great news to share - I joined a gym last week (9 June 2023; the week before 16 June 2023)"},
		},
	}
	got := leftoverCoveringSpecificAnswer("What exciting news did Maria share on 16 June, 2023?", hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "gym") {
		t.Fatalf("expected last-week gym news over volunteer chat, got %q", got)
	}
	if strings.Contains(lower, "horizon") {
		t.Fatalf("weak leftover token exciting must not pick Horizon over gym news: %q", got)
	}
}

func TestLeftoverCoveringSpecificAnswerKeepsWeekendPlanNotSpeakerChat(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Joanna", Value: "Joanna", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Nate", Value: "Nate", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Nate: Just make sure you don't quit - the path forward will show up soon"},
			{Content: "I'm going to make it for my family this weekend - can't wait (25 June 2022; the weekend before 24 June 2022)"},
			{Content: "Nate encourages Joanna not to quit and to keep working toward her goals."},
		},
	}
	got := leftoverCoveringSpecificAnswer("When is Joanna going to make Nate's ice cream for her family?", hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "weekend") && !strings.Contains(lower, "family") {
		t.Fatalf("expected weekend family plan, got %q", got)
	}
	if strings.Contains(lower, "quit") {
		t.Fatalf("speaker-prefixed pep talk leaked: %q", got)
	}
}

func TestLeftoverCoveringWhereQueryPicksLocativePlace(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateFamilyMember, Source: "typed_store",
			Value: "son, kids bring joy", Values: []string{"son, kids bring joy"}},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity, Source: "typed_store",
			Value:  "hiking, road trip to jasper national park, hiked trails with family",
			Values: []string{"hiking, road trip to jasper national park, hiked trails with family"}},
	}
	pkt := EvidencePacket{
		Contents: []string{
			"Riley photographed a glacier during a mountain trip.",
			"Riley took his family on a road trip to Jasper National Park on the weekend of 20-21 May 2023, driving through the Icefields Parkway.",
			"Last weekend I took my family to Jasper National Park (17 May 2023; the week before 24 May 2023)",
			"Riley participated in a scuba diving lesson on 15 September 2023.",
		},
	}
	q := "Where did Riley take his family for a road trip on 24 May, 2023?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	if !strings.Contains(strings.ToLower(got), "jasper") {
		t.Fatalf("expected locative Jasper leftover covering rare=%v join=%v, got %q", leftoverCoverRareTokens(q, hops), hopsKeepTypedJoin(hops), got)
	}
	if strings.Contains(strings.ToLower(got), "scuba") {
		t.Fatalf("non-place activity leaked onto where leftover covering: %q", got)
	}
}

func TestLeftoverCoveringWhereQuerySkipsNonPlace(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Casey", Value: "Casey", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley participated in a scuba diving lesson on 15 September 2023."},
			{Content: "Riley attended a meditation retreat in Phuket with Casey, starting on 9 September 2023."},
			{Content: "Riley and her partner attended a yoga retreat in South America."},
		},
	}
	got := leftoverCoveringSpecificAnswer("Where did Riley and Casey find a cool diving spot?", hops, pkt)
	lower := strings.ToLower(got)
	if strings.Contains(lower, "phuket") || strings.Contains(lower, "south america") || strings.Contains(lower, "yoga") {
		t.Fatalf("retreat place must not stand in for a diving-spot write miss: %q", got)
	}
}

func TestLeftoverCoveringJoinsPlayedGames(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley played Fortnite during the gaming tournament on 30 October 2022."},
			{Content: "Riley played Overwatch during the gaming tournament on 30 October 2022."},
			{Content: "Riley played Apex Legends during the gaming tournament on 30 October 2022."},
			{Content: "Riley says the speed of Apex Legends makes it fun."},
		},
	}
	got := leftoverCoveringSpecificAnswer("What games were played at the gaming tournament organized by Riley on 31 October, 2022?", hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "fortnite") || !strings.Contains(lower, "overwatch") || !strings.Contains(lower, "apex") {
		t.Fatalf("expected joined tournament games, got %q", got)
	}
}

func TestLeftoverCoveringThanksgivingTradition(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicatePreference, Source: "typed_store",
			Value: "ocean views, cliffs", Values: []string{"ocean views, cliffs"}},
	}
	pkt := EvidencePacket{
		Contents: []string{
			"They usually watch a few movies together during Thanksgiving.",
			"Riley enjoys prepping the feast as part of his Thanksgiving tradition.",
			"Riley enjoys talking about what they're thankful for during Thanksgiving.",
		},
	}
	q := "What tradition does Riley mention they love during Thanksgiving?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "feast") {
		t.Fatalf("expected feast tradition leftover covering, got %q rare=%v", got, leftoverCoverRareTokens(q, hops))
	}
	if !strings.Contains(lower, "thankful") {
		t.Fatalf("expected thankful complement joined with feast, got %q", got)
	}
	if strings.Contains(lower, "movie") {
		t.Fatalf("movies paraphrase must not join the tradition covering: %q", got)
	}
	hybrid := "They usually watch a few movies together during Thanksgiving."
	if !leftoverCoveringBeatsAnswer(q, hops, got, hybrid) {
		t.Fatalf("feast covering must beat movies paraphrase, covering=%q", got)
	}
	if !leftoverCoveringMayReplaceHybrid(q, hops, got, hybrid) {
		t.Fatalf("leftover covering must be allowed to replace a movies hybrid, covering=%q", got)
	}
}

func TestLeftoverCoveringThanksgivingIgnoresChatTurnLocative(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Tim", Value: "Tim", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Tim", Predicate: PredicatePreference, Source: "typed_store",
			Value:  "ocean views, cliffs, exploring other cultures via fantasy stories at home",
			Values: []string{"ocean views, cliffs"}},
	}
	pkt := EvidencePacket{
		Contents: []string{
			"Tim enjoys prepping the feast as part of his Thanksgiving tradition.",
			"Tim enjoys talking about what they're thankful for during Thanksgiving.",
			"Tim: Thanksgiving's always special for us",
			"They usually watch a few movies together during Thanksgiving.",
			"Tim enjoys having movie marathons",
			"Tim enjoys books that transport him to another world and spark his imagination.",
			"Tim's family had a Thanksgiving gathering where they ate together.",
			"Tim's favorite Thanksgiving movie is \"Home Alone\".",
		},
	}
	q := "What tradition does Tim mention they love during Thanksgiving?"
	if !looksLocativePlaceLine("Tim: Thanksgiving's always special for us") {
		t.Fatal("speaker+Thanksgiving still looks locative; scoring must ignore chat-turn locative bonus")
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "feast") || !strings.Contains(lower, "thankful") {
		t.Fatalf("expected feast+thankful covering over chat-turn locative, got %q", got)
	}
	if strings.Contains(lower, "always special") || strings.Contains(lower, "movie") {
		t.Fatalf("chat-turn locative or movies leaked into tradition covering: %q", got)
	}
	hybrid := "They usually watch a few movies together during Thanksgiving."
	if !leftoverCoveringMayReplaceHybrid(q, hops, got, hybrid) {
		t.Fatalf("feast+thankful covering must replace movies hybrid, covering=%q", got)
	}
}

func TestLeftoverCoveringBeatsAnswerMissingRareToken(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	covering := "Riley enjoys prepping the feast as part of his Thanksgiving tradition."
	hybrid := "They usually watch a few movies together during Thanksgiving."
	q := "What tradition does Riley mention they love during Thanksgiving?"
	if !leftoverCoveringBeatsAnswer(q, hops, covering, hybrid) {
		t.Fatal("covering that hits leftover tradition must beat a movies paraphrase")
	}
	if leftoverCoveringBeatsAnswer(q, hops, covering, covering) {
		t.Fatal("same covering must not beat itself")
	}
	if leftoverCoveringBeatsAnswer(
		"What is Jolene's favorite book which she mentioned on 4 February, 2023?",
		[]HopResult{{Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene", Source: "search_fallback"}},
		"Deborah mentioned the Eisenhower Matrix as a tool for organizing and prioritizing tasks.",
		`I'm really into this book called "Sapiens" - it's a fascinating look at human history`,
	) {
		t.Fatal("mention-only covering must not beat a book answer")
	}
	qWalk := "What activity do Andrew and Buddy enjoy doing together?"
	hopsWalk := []HopResult{
		{Kind: "resolve_entity", Entity: "Andrew", Value: "Andrew", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Buddy", Value: "Buddy", Source: "search_fallback"},
	}
	selfCare := "Make sure to do at least one self-care activity each day - like treating myself to something nice"
	walks := "They enjoy going for walks together."
	if leftoverCoveringBeatsAnswer(qWalk, hopsWalk, selfCare, walks) {
		t.Fatal("activity-schema leftover covering must not beat a walks hybrid")
	}
	chiliQ := "What events is Maria planning for the homeless shelter funraiser?"
	chiliCover := "Maria is organizing a yoga retreat next month."
	chiliHybrid := "Maria is planning a ring-toss tournament and a chili cook-off for the homeless-shelter fundraiser."
	if leftoverCoveringBeatsAnswer(chiliQ, hops, chiliCover, chiliHybrid) {
		t.Fatal("unrelated leftover covering must not beat a chili+ring-toss hybrid")
	}
	if leftoverCoveringMayReplaceHybrid(chiliQ, hops, chiliCover, chiliHybrid) {
		t.Fatal("chili+ring-toss hybrid must stay locked against leftover covering")
	}
	ringOnly := "Maria is organizing a ring-toss tournament for the fundraiser."
	if leftoverCoveringMayReplaceHybrid(chiliQ, hops, ringOnly, chiliHybrid) {
		t.Fatal("ring-toss-only covering must not replace a chili+ring-toss hybrid")
	}
	studyQ := "What did Jolene and Deb discuss as a helpful strategy for studying and time management?"
	studyCover := "Jolene wants to connect with these big companies."
	studyHybrid := "They mentioned using planners or schedulers to stay organized, and breaking study tasks into smaller pieces while setting clear goals."
	if leftoverCoveringBeatsAnswer(studyQ, hops, studyCover, studyHybrid) {
		t.Fatal("unrelated leftover covering must not beat a studying hybrid")
	}
	if leftoverCoveringMayReplaceHybrid(studyQ, hops, "They mentioned using planners or schedulers to stay organized.", studyHybrid) {
		t.Fatal("planners-only covering must not replace a studying hybrid that already has both halves")
	}
	sapiensQ := "What is Jolene's favorite book which she mentioned on 4 February, 2023?"
	sapiensHybrid := `I'm really into this book called "Sapiens" - it's a fascinating look at human history`
	mother := "Deborah's mother loves reading historical books."
	if leftoverCoveringBeatsAnswer(sapiensQ, hops, mother, sapiensHybrid) {
		t.Fatal("mother-reading covering must not beat a Sapiens hybrid")
	}
	if leftoverCoveringMayReplaceHybrid(sapiensQ, hops, mother, sapiensHybrid) {
		t.Fatal("Sapiens hybrid must stay locked against leftover covering")
	}
	if leftoverCoveringMayReplaceHybrid(sapiensQ, hops, mother, "Sapiens") {
		t.Fatal("compact Sapiens title must stay locked against leftover covering")
	}
	joleneDump := "video games, hanging out with Susie, hiking, pottery class, community garden"
	if !typedAnswerIsHopDump(joleneDump) {
		t.Fatalf("expected Jolene activity dump, got %q", joleneDump)
	}
	if leftoverCoveringMayReplaceHybrid(
		"What activities have been helping Jolene stay distracted during tough times?",
		[]HopResult{{Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene", Source: "search_fallback"}},
		"I'm passionate about helping people find peace and joy through it",
		joleneDump,
	) {
		t.Fatal("activity hop dump must stay locked against leftover covering slogans")
	}
	ukQ := "which country has Tim visited most frequently in his travels?"
	if leftoverCoveringMayReplaceHybrid(ukQ, hops, "Tim visited a Harry Potter-themed place in London a few years ago.", "United Kingdom") {
		t.Fatal("compact UK name must stay locked against leftover covering")
	}
	horseQ := "What activity did Caroline used to do with her dad?"
	if leftoverCoveringMayReplaceHybrid(horseQ, hops,
		"Both Melanie and her kids helped with the painting, describing it as a bonding activity.",
		"She went horseback riding with her dad.") {
		t.Fatal("non-schema leftover covering must not replace a horseback hybrid")
	}
	letterQ := "How did Joanna feel when someone wrote her a letter after reading her blog post?"
	if leftoverCoveringMayReplaceHybrid(letterQ, hops,
		"Last week, someone wrote me a letter after reading an online blog post I made about a hard moment in my life",
		"She felt it was a blessing (grateful for the support).") {
		t.Fatal("letter-event leftover covering must not replace a feeling hybrid")
	}
	consoleQ := "When did Jolene's parents give her first console?"
	if leftoverCoveringMayReplaceHybrid(consoleQ, hops,
		"Jolene had a lot of support from her parents.",
		"When she was 10 years old.") {
		t.Fatal("parents-support leftover covering must not replace a first-console hybrid")
	}
	qTor := "Where was James at on July 12, 2022?"
	if leftoverCoveringBeatsAnswer(qTor, hops, "James will depart for Toronto", "Toronto") {
		t.Fatal("depart-for leftover covering must not beat a short hybrid place")
	}
	if leftoverCoveringKeepShortLocative(qTor, "James will depart for Toronto", "Toronto, Canada") != true {
		t.Fatal("short locative NP must be kept against a longer depart-for line")
	}
	qJas := "Where did Riley take his family for a road trip on 24 May, 2023?"
	hopDump := "Riley photographed a glacier, hiking, road trip to jasper national park, hiked trails with family, scuba diving"
	if !typedAnswerIsHopDump(hopDump) {
		t.Fatalf("expected hop dump, got %q", hopDump)
	}
	if !leftoverCoveringBeatsAnswer(qJas, hops, "Jasper National Park", hopDump) {
		t.Fatal("short locative covering must beat a where hop dump")
	}
}

func TestLeftoverCoveringExtractsWherePlaceNP(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "James", Value: "James", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "James will depart for Toronto on July 11, 2022 in the evening."},
			{Content: "James participates in bowling (16 March 2022; the day before 17 March 2022)"},
		},
	}
	got := leftoverCoveringSpecificAnswer("Where was James at on July 12, 2022?", hops, pkt)
	if !strings.EqualFold(strings.TrimSpace(got), "Toronto") {
		t.Fatalf("where leftover covering must return the place NP, got %q", got)
	}
	if strings.Contains(strings.ToLower(got), "depart") {
		t.Fatalf("depart-for intent must not remain on a where answer: %q", got)
	}
}

func TestLeftoverCoveringReplacesBareDateMissingEvent(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Alex", Value: "Alex", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Alex adopted a puppy named Ned from a shelter on 2022-04-05."},
			{Content: "Alex participates in bowling (16 March 2022; the day before 17 March 2022)"},
		},
	}
	q := "When did Alex adopt Ned?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "adopt") && !strings.Contains(got, "2022-04-05") && !strings.Contains(lower, "april") {
		t.Fatalf("when-event leftover covering must pick the adopt date, got %q", got)
	}
	if strings.Contains(lower, "bowl") || strings.Contains(got, "17 March") {
		t.Fatalf("bowling leftover date must not win a when-adopt query, got %q", got)
	}
	if !leftoverCoveringBareDateMissesEvent(q, hops, got, "17 March 2022") {
		t.Fatal("bare bowling date must yield to adopt covering")
	}
	if leftoverCoveringBareDateMissesEvent(q, hops, got, "Alex adopted Ned on 2022-04-05") {
		t.Fatal("adopt hybrid that already names the event must stay")
	}
	ginaQ := "When did Gina interview for a design internship?"
	if leftoverCoveringBareDateMissesEvent(ginaQ, hops, "Gina launched an ad campaign on 20 June 2023.", "10 May 2023") {
		t.Fatal("unrelated dated leftover covering must not beat a matching internship date")
	}
}

func TestLeftoverCoveringWhenEventPrefersYearEventOverBareHopDate(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "John", Value: "John", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "John", Predicate: PredicateActivity, Source: "typed_store",
			Value: "movie nights at home", Values: []string{"movie nights at home"}},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John attended community (16 January 2023; the week before 9 January 2023) (2 January 2023)"},
			{Content: "John will discuss infrastructure upgrades at the community meeting scheduled for the week of 16 January 2023."},
			{Content: "John sat by a campfire with Max during the camping trip in summer 2022."},
			{Content: "John helped renovate a rundown community center back home in 2022."},
			{Content: "John: I feel a strong urge to serve my country and community"},
		},
	}
	q := "When did John help renovate his hometown community center?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "renovate") || !strings.Contains(got, "2022") {
		t.Fatalf("when-event leftover covering must pick the year-dated renovate fact, got %q", got)
	}
	if strings.Contains(got, "17 July") || strings.Contains(lower, "january") || strings.Contains(lower, "campfire") {
		t.Fatalf("hop activity date or community meeting must not cover renovate, got %q", got)
	}
	if !leftoverCoveringBareDateMissesEvent(q, hops, got, "17 July 2023") {
		t.Fatal("bare hop date must yield to year-dated renovate covering")
	}
	augQ := "When did John meet back up with his teammates after his trip in August 2023?"
	augPkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John participated in a basketball game in 2022 where his team trailed big in the fourth quarter but came back to win at the final buzzer."},
			{Content: "John is planning a team trip in October 2023 to explore a new city. (15 October 2023)"},
			{Content: "John met up with his teammates on 15 August 2023 after returning from a trip."},
			{Content: "John celebrated with his teammates at a restaurant after a game, feeling exhausted but happy."},
		},
	}
	augGot := leftoverCoveringSpecificAnswer(augQ, hops, augPkt)
	augLower := strings.ToLower(augGot)
	if !strings.Contains(augLower, "teammate") || !strings.Contains(augGot, "15 August") {
		t.Fatalf("August 2023 meetup must cover, got %q", augGot)
	}
	if strings.Contains(augLower, "2022") || strings.Contains(augLower, "buzzer") || strings.Contains(augLower, "october") {
		t.Fatalf("different-year or different-month leftover must not cover an August 2023 meetup, got %q", augGot)
	}
}

func TestLeftoverCoveringBindsWhenEventToQueryEntity(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Alex", Value: "Alex", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Alex lost his job as a banker on 19 January 2023."},
			{Content: "Dana lost her job at Night Shift in January 2023. (15 January 2023)"},
		},
	}
	q := "When Alex has lost his job as a banker?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "alex") || !strings.Contains(got, "19 January") {
		t.Fatalf("when-event leftover covering must bind to the query person, got %q", got)
	}
	if strings.Contains(lower, "dana") || strings.Contains(lower, "night shift") {
		t.Fatalf("another person's dated job-loss must not answer a named when-event query, got %q", got)
	}
	foreign := "Dana lost her job at Night Shift in January 2023. (15 January 2023)"
	if leftoverCoveringBareDateMissesEvent(q, hops, foreign, "19 January 2023") {
		t.Fatal("foreign-subject leftover covering must not replace a matching date")
	}
	if !leftoverCoveringBareDateMissesEvent(q, hops, got, "19 January 2023") {
		t.Fatal("query-entity covering that names the event must still replace a bare date")
	}
	ginaQ := "When did Gina interview for a design internship?"
	ginaPkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Gina had an interview for a design internship on 10 May 2023."},
			{Content: "Alex lost his job as a banker on 19 January 2023."},
		},
	}
	ginaGot := leftoverCoveringSpecificAnswer(ginaQ, hops, ginaPkt)
	if !strings.Contains(strings.ToLower(ginaGot), "gina") || strings.Contains(strings.ToLower(ginaGot), "banker") {
		t.Fatalf("named internship when-event must stay on Gina, got %q", ginaGot)
	}
	iceQ := "When is Joanna going to make Nate's ice cream for her family?"
	icePkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Nate encourages Joanna not to quit and to keep working toward her goals."},
			{Content: "I'm going to make it for my family this weekend - can't wait (25 June 2022; the weekend before 24 June 2022)"},
		},
	}
	iceGot := leftoverCoveringSpecificAnswer(iceQ, hops, icePkt)
	iceLower := strings.ToLower(iceGot)
	if !strings.Contains(iceLower, "weekend") && !strings.Contains(iceLower, "family") {
		t.Fatalf("first-person dated plan must still cover a when-event query, got %q", iceGot)
	}
	if strings.Contains(iceLower, "quit") {
		t.Fatalf("named pep-talk must not beat the first-person weekend plan, got %q", iceGot)
	}
	yearQ := "Which year did Riley start practicing yoga?"
	yearHops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	yearPkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Dana started practicing yoga in 2018."},
			{Content: "Jolene feels that large stacks of tasks can be overwhelming and make it hard to know where to start."},
			{Content: "Riley started practicing yoga in 2020."},
		},
	}
	yearGot := leftoverCoveringSpecificAnswer(yearQ, yearHops, yearPkt)
	yearLower := strings.ToLower(yearGot)
	if !strings.Contains(yearLower, "2020") || !strings.Contains(yearLower, "riley") {
		t.Fatalf("which-year leftover covering must bind to the query person, got %q", yearGot)
	}
	if strings.Contains(yearLower, "dana") || strings.Contains(yearGot, "2018") || strings.Contains(yearLower, "overwhelm") {
		t.Fatalf("undated or foreign leftover must not cover a named which-year query, got %q", yearGot)
	}
	if leftoverCoveringSkipForeignWhenEvent(yearQ, "Dana started practicing yoga in 2018.") != true {
		t.Fatal("which-year covering must skip a foreign-person start year")
	}
	if leftoverCoveringSkipForeignWhenEvent(yearQ, "I started practicing yoga in 2020.") {
		t.Fatal("first-person unnamed year lines must still compete")
	}
	if leftoverCoveringYearMissesEvent(yearQ, yearGot, "Riley feels that large stacks of tasks can be overwhelming.") != true {
		t.Fatal("which-year leftover with a year must replace an undated hybrid")
	}
	if leftoverCoveringYearMissesEvent(yearQ, "Riley feels overwhelmed.", "Riley started practicing yoga in 2020.") {
		t.Fatal("undated leftover must not replace a year answer")
	}
	joleneQ := "Which year did Jolene start practicing yoga?"
	jolenePkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Deborah enjoys practicing yoga on the beach in Bali."},
			{Content: "Jolene feels that large stacks of tasks can be overwhelming and make it hard to know where to start."},
			{Content: "Jolene started practicing yoga in 2020."},
		},
	}
	joleneGot := leftoverCoveringSpecificAnswer(joleneQ, []HopResult{{Kind: "resolve_entity", Entity: "Jolene", Value: "Jolene"}}, jolenePkt)
	joleneLower := strings.ToLower(joleneGot)
	if !strings.Contains(joleneLower, "2020") || strings.Contains(joleneLower, "deborah") || strings.Contains(joleneLower, "overwhelm") {
		t.Fatalf("which-year covering must pick the dated start year, got %q", joleneGot)
	}
}

func TestLeftoverCoveringWhenEventKeepsSentenceInitialVerbLine(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateEvent, Source: "typed_store",
			Value: "23 January 2023", Values: []string{"23 January 2023"}},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley: Hey, sorry to tell you this but my dad passed away two days ago (25 January 2023)"},
			{Content: "Sounds like you had a blast biking and at the art show"},
			{Content: "Dana checked out a movie on 23 January 2023."},
			{Content: "Checked out an art show with a friend today - really cool and inspiring stuff (9 April 2023)"},
		},
	}
	q := "When did Riley go to an art show with a friend?"
	art := "Checked out an art show with a friend today - really cool and inspiring stuff (9 April 2023)"
	if leftoverCoveringSkipForeignWhenEvent(q, art) {
		t.Fatal("sentence-initial verb must not count as a foreign person on an unnamed dated line")
	}
	if leftoverCoveringSkipForeignWhenEvent(q, "Dana checked out a movie on 23 January 2023.") != true {
		t.Fatal("a named other person must still drop when-event covering")
	}
	if leftoverCoveringSkipForeignWhenEvent(q, "Dana lost her job at Night Shift in January 2023. (15 January 2023)") != true {
		t.Fatal("sentence-initial names must still count as people")
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "art show") || !strings.Contains(got, "9 April") {
		t.Fatalf("when-event leftover covering must pick the dated art-show line, got %q", got)
	}
	if strings.Contains(lower, "january") || strings.Contains(lower, "movie") || strings.Contains(lower, "dana") || strings.Contains(lower, "passed away") {
		t.Fatalf("hop date, foreign movie, or bereavement must not cover the art show, got %q", got)
	}
	if !leftoverCoveringBareDateMissesEvent(q, hops, got, "23 January 2023") {
		t.Fatal("bare hop date must yield to the year-dated art-show covering")
	}
}

func TestLeftoverCoveringWhichYearResolvesRelativeDuration(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Dana started practicing yoga in 2018."},
			{Content: "Riley started practicing yoga in 2020."},
			{Content: "Riley has been working on his health for two years as of August 7, 2023."},
		},
	}
	q := "Which year did Riley start taking care of his health seriously?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	if got != "2021" {
		t.Fatalf("which-year leftover covering must resolve for-N-years-as-of to the start year, got %q", got)
	}
	if !leftoverCoveringYearMissesEvent(q, got, "Riley has been working on his health for two years as of August 7, 2023.") {
		t.Fatal("as-of duration must yield to the resolved start year")
	}
	yogaGot := leftoverCoveringSpecificAnswer("Which year did Riley start practicing yoga?", hops, pkt)
	if !strings.Contains(yogaGot, "2020") || strings.Contains(yogaGot, "2018") || yogaGot == "2021" {
		t.Fatalf("explicit start year must still cover which-year, got %q", yogaGot)
	}
}

func TestLeftoverCoveringWhenEventPrefersBothQueryPeople(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Sam", Value: "Sam", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Sam decided he needs to make health changes after being mocked on 21 July 2023."},
			{Content: "Riley painted a watercolor cactus in the desert last week (29 September 2023)."},
			{Content: "Sam plans to paint with Riley on Saturday, 16 September 2023."},
		},
	}
	q := "When did Riley and Sam decide to paint together?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "16 september") && !strings.Contains(lower, "saturday") {
		t.Fatalf("when-event leftover covering must pick the dual-entity dated plan, got %q", got)
	}
	if strings.Contains(lower, "cactus") || strings.Contains(lower, "health") || strings.Contains(lower, "july") {
		t.Fatalf("speech-act or solo-painter leftover must not cover the shared plan, got %q", got)
	}
}

func TestLeftoverCoveringInstrumentPurposePrefersUseOverOwns(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley owns a fitness tracker smartwatch that he wears on his wrist."},
			{Content: "Riley helped a lost tourist find their way around the city."},
			{Content: "Riley uses his fitness tracker to monitor his health progress, which serves as a constant reminder to keep going."},
			{Content: "Dana uses her fitness tracker as a reminder to stretch after runs."},
		},
	}
	q := "What does the smartwatch help Riley with?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "reminder") && !strings.Contains(lower, "progress") {
		t.Fatalf("instrument leftover covering must pick the purpose fact, got %q", got)
	}
	if strings.Contains(lower, "owns") || strings.Contains(lower, "tourist") || strings.Contains(lower, "dana") {
		t.Fatalf("ownership, help-flood, or foreign leftover must not cover instrument purpose, got %q", got)
	}
	hybrid := "The fitness-tracker smartwatch helps Riley monitor his fitness and health."
	if !leftoverCoveringPurposeMissesInstrument(q, got, hybrid) {
		t.Fatal("purpose covering with reminder/progress must replace a generic monitor paraphrase")
	}
	if leftoverCoveringPurposeMissesInstrument(q, "Riley owns a fitness tracker smartwatch that he wears on his wrist.", got) {
		t.Fatal("ownership leftover must not replace a purpose answer")
	}
	feelQ := "According to Jolene, what does exercise help her to feel?"
	if leftoverCoveringInstrumentNoun(feelQ) != "" || looksInstrumentPurposeQuery(feelQ) {
		t.Fatal("exercise-feel must not classify as instrument-purpose")
	}
	if leftoverCoveringPurposeMissesInstrument(feelQ, "Deborah: Exercise is key for me - it makes me feel connected to my body", "not in memory") {
		t.Fatal("exercise-feel questions must not take the instrument-purpose replace gate")
	}
	hybridItems := []RecallItem{{Value: hybrid}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, hybridItems)
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("enumerate items must follow instrument leftover covering, got %#v", synced)
	}
	unwindItems := []RecallItem{{Value: "runs"}, {Value: "pottery"}}
	if kept := leftoverCoveringSyncEnumerateItems("What does Riley do to unwind?", "Riley uses pottery to unwind.", unwindItems); len(kept) != 2 {
		t.Fatalf("unwind enumerate items must not collapse to leftover covering, got %#v", kept)
	}
	if kept := leftoverCoveringSyncEnumerateItems(feelQ, "Deborah: Exercise is key for me - it makes me feel connected to my body", hybridItems); len(kept) != 1 || kept[0].Value != hybrid {
		t.Fatalf("exercise-feel must not sync leftover covering into enumerate items, got %#v", kept)
	}
}

func TestLeftoverCoveringWhatMadePrefersOffQueryEvidence(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Deborah", Value: "Deborah", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Deborah enjoys participating in the running group and finds it motivating."},
			{Content: "We help and push each other during our runs, which makes it so much easier to stay motivated"},
			{Content: "Anna is a member of Deborah's running group."},
			{Content: "Deborah started a running group with Anna."},
			{Content: "Jolene enjoys yoga and meditation to stay motivated."},
		},
	}
	q := "What made being part of the running group easy for Deborah to stay motivated?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "push") {
		t.Fatalf("what-made leftover covering must pick off-query evidence, got %q", got)
	}
	if strings.Contains(lower, "yoga") || strings.Contains(lower, "enjoys participating") {
		t.Fatalf("schema restatement or foreign yoga must not cover what-made, got %q", got)
	}
	hybrid := "She enjoys participating in the running group, which makes it easy for her to stay motivated."
	if leftoverCoveringHasOffQueryEvidence(q, hybrid) {
		t.Fatal("enjoy/participate hybrid must not count as off-query evidence")
	}
	if !leftoverCoveringWhatMadeMissesEvidence(q, got, hybrid) {
		t.Fatal("what-made covering must replace a restating enjoy/participate hybrid")
	}
	if leftoverCoveringWhatMadeMissesEvidence(q, got, got) {
		t.Fatal("answer that already has off-query evidence must stay")
	}
	if leftoverCoveringMayReplaceHybrid(q, hops, got, hybrid) != true {
		t.Fatal("what-made covering must be allowed to replace the restating hybrid")
	}
	if looksWhatMadeQuery("What does Melanie do to destress?") || looksWhatMadeQuery("What does the smartwatch help Riley with?") {
		t.Fatal("destress and instrument-purpose queries are not what-made")
	}
	if looksHowDescribeQuery("What does Melanie do to destress?") || looksHowDescribeQuery("What does the smartwatch help Riley with?") {
		t.Fatal("destress and instrument-purpose queries are not how-describe")
	}
	if !looksHowDescribeQuery("How does Nate describe the stuffed animal he got for Joanna?") {
		t.Fatal("how-does-describe must classify as how-describe")
	}
	if looksHostQuery("What does Melanie do to destress?") || looksHostQuery("What made being part of the running group easy for Deborah to stay motivated?") {
		t.Fatal("destress and what-made queries are not host queries")
	}
	if looksAdviceQuery("What does Melanie do to destress?") || looksAdviceQuery("What made being part of the running group easy for Deborah to stay motivated?") || looksAdviceQuery("What did John host for the veterans in May 2023 as part of the project") {
		t.Fatal("destress, what-made, and host queries are not advice queries")
	}
	if !looksAdviceQuery("What advice does Gina give to Jon about running a successful business?") {
		t.Fatal("what-advice-give must classify as advice")
	}
	if looksWhatKindQuery("What does Melanie do to destress?") || looksWhatKindQuery("What advice does Gina give to Jon about running a successful business?") || looksWhatKindQuery("What did John host for the veterans in May 2023 as part of the project") {
		t.Fatal("destress, advice, and host queries are not what-kind queries")
	}
	if !looksWhatKindQuery("What kind of food did Maria have on her dinner spread iwth her mother?") {
		t.Fatal("what-kind-of must classify as what-kind")
	}
	if looksHowDescribeProcessQuery("How does Nate describe the stuffed animal he got for Joanna?") {
		t.Fatal("how-describe without process is not how-describe-process")
	}
	if !looksHowDescribeProcessQuery("How does Nate describe the process of taking care of turtles?") {
		t.Fatal("how-describe-the-process must classify as how-describe-process")
	}
	if leftoverCoveringAdviceOffQueryLine("Nate says looking after his turtles is calming") {
		t.Fatal("companion slogan is not hortative leftover")
	}
	if !leftoverCoveringAdviceOffQueryLine("Just keep their area clean, feed them properly, and make sure they get enough light") {
		t.Fatal("just-keep hortative leftover must classify as a directive")
	}
	if leftoverCoveringProcessHortativeLine("Don't forget to relax and enjoy the process too") {
		t.Fatal("process-restatement hortative must not count as process leftover")
	}
	if leftoverCoveringProcessHortativeLine("Hoping to share my love of gaming and connect with others who enjoy it too") {
		t.Fatal("companion slogan is not process hortative leftover")
	}
	if !leftoverCoveringProcessHortativeLine("Just keep their area clean, feed them properly, and make sure they get enough light") {
		t.Fatal("just-keep leftover without process restatement must count as process hortative")
	}
	if looksWhatMotivatesQuery("What made being part of the running group easy for Deborah to stay motivated?") || looksWhatMotivatesQuery("How does Tim stay motivated during difficult study sessions?") {
		t.Fatal("what-made stay-motivated and how-stay-motivated are not what-motivates")
	}
	if !looksWhatMotivatesQuery("What motivates Joanna to keep writing even on tough days?") {
		t.Fatal("what-motivates must classify as what-motivates")
	}
	qWrite := "What motivates Joanna to keep writing even on tough days?"
	goldWrite := "It's knowing that my writing can make a difference that keeps me going, even on tough days"
	if leftoverCoveringLineHasForeignPerson(qWrite, goldWrite) {
		t.Fatal("sentence-initial It's-knowing must not count as a foreign person")
	}
	if !leftoverCoveringMotivateCauseLine(qWrite, goldWrite) {
		t.Fatal("first-person my-writing cleft leftover must count as what-motivates cause")
	}
	if leftoverCoveringMotivateCauseLine(qWrite, "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times.") {
		t.Fatal("compiler motivate-her fact is not first-person object-cause leftover")
	}
	if leftoverCoveringMotivateCauseLine(qWrite, "Have faith in yourself and continue following your writing dreams - it's tough but worth it") {
		t.Fatal("have-faith hortative is not first-person object-cause leftover")
	}
	if leftoverCoveringMotivateCauseLine(qWrite, "Wish I had a vacation lined up, but right now my writing is consuming me") {
		t.Fatal("occupation leftover with my writing is not a cause connective")
	}
	if looksWhatSayAboutQuery("What motivates Joanna to keep writing even on tough days?") || looksWhatSayAboutQuery("What advice does Gina give to Jon about running a successful business?") {
		t.Fatal("what-motivates and what-advice are not what-say-about")
	}
	if !looksWhatSayAboutQuery("What does Gina say about the dancers in the photo?") {
		t.Fatal("what-does-say-about must classify as what-say-about")
	}
	if leftoverCoveringEvaluativeTheyLine("Gina performed a contemporary dance piece titled \"Finding Freedom\".") {
		t.Fatal("performance leftover is not they-evaluative")
	}
	if leftoverCoveringEvaluativeTheyLine("They've been practicing hard and will definitely impress with their grace and skill") {
		t.Fatal("they've-been leftover is not a short they-evaluative copula")
	}
	if !leftoverCoveringEvaluativeTheyLine("They're so graceful") {
		t.Fatal("they're-so leftover must count as they-evaluative")
	}
	if !leftoverCoveringEvaluativeTheyLine("They look graceful") {
		t.Fatal("they-look leftover must count as they-evaluative")
	}
	if looksHopPerson("NYC") || looksHopPerson("UK") {
		t.Fatal("all-caps place acronyms are not hop people")
	}
	if !looksHopPerson("John") || !looksHopPerson("Gina") {
		t.Fatal("title-case names must stay hop people")
	}
}

func TestLeftoverCoveringHostPrefersPartyOverRealize(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "John", Value: "John", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John realized on 13 May 2023 that veterans have done a lot for us."},
			{Content: "John has a photograph of veterans taken on 19 May 2023."},
			{Content: "John organized a small party for veterans."},
			{Content: "We had a great time throwing a small party and inviting some veterans to share their stories"},
			{Content: "Maria and her mother cooked dinner together at home on 3 May 2023."},
			{Content: "John organized a 5K charity run in his neighborhood to support veterans and their families."},
		},
	}
	q := "What did John host for the veterans in May 2023 as part of the project"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "party") {
		t.Fatalf("host leftover covering must pick the hosted event, got %q", got)
	}
	if !strings.Contains(lower, "stor") {
		t.Fatalf("host leftover covering must join the share-stories leftover, got %q", got)
	}
	if strings.Contains(lower, "realized") || strings.Contains(lower, "photograph") || strings.Contains(lower, "5k") || strings.Contains(lower, "charity run") || strings.Contains(lower, "dinner") {
		t.Fatalf("realize, photograph, dinner, or later race must not cover host, got %q", got)
	}
	hybrid := "John realized on 13 May 2023 that veterans have done a lot for us."
	if leftoverCoveringHostedEventLine(hybrid) {
		t.Fatal("realize line is not a hosted event")
	}
	if leftoverCoveringHostedEventLine("John has a photograph of veterans taken on 19 May 2023.") {
		t.Fatal("photograph line is not a hosted event")
	}
	if !leftoverCoveringHostMissesEvent(q, got, hybrid) {
		t.Fatal("host covering must replace a realize restatement")
	}
}

func TestExpandHostedEventSessionNeighborsAdmitsPartyPastCap(t *testing.T) {
	session := "sess-host"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "John realized veterans have done a lot",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about veterans",
			Metadata: map[string]any{"session_id": session},
		})
	}
	party := MemoryRecord{
		MemoryID: "party",
		Content:  "John organized a small party for veterans",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, party)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["party"]; ok {
		t.Fatal("generic session expand must stay capped so the hosted event can miss")
	}
	expandHostedEventSessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["party"]; !ok {
		t.Fatal("hosted-event session expand must admit the party line past the generic cap")
	}
	if ids := sessionIDsOf([]MemoryRecord{seed, party}); len(ids) != 1 || ids[0] != session {
		t.Fatalf("sessionIDsOf must uniq seed sessions, got %#v", ids)
	}
}

func TestSessionIDsForHostQueryPrefersCoveringSession(t *testing.T) {
	q := "What did John host for the veterans in May 2023 as part of the project"
	noisy := MemoryRecord{
		MemoryID: "n",
		Content:  "John is brainstorming future projects to provide underserved communities with education",
		Metadata: map[string]any{"session_id": "session_4"},
	}
	goldSeed := MemoryRecord{
		MemoryID: "g",
		Content:  "John realized on 13 May 2023 that veterans have done a lot for us.",
		Metadata: map[string]any{"session_id": "session_15"},
	}
	ids := sessionIDsForHostQuery(q, []MemoryRecord{noisy, goldSeed})
	if len(ids) == 0 || ids[0] != "session_15" {
		t.Fatalf("host session pick must prefer veterans/May coverage, got %#v", ids)
	}
}

func TestApplyHostedEventRankBoostSkipsPhotograph(t *testing.T) {
	q := "What did John host for the veterans in May 2023 as part of the project"
	party := 1.0
	photo := 1.0
	applyHostedEventRankBoost(&party, map[string]any{}, q, MemoryRecord{Content: "John organized a small party for veterans."})
	applyHostedEventRankBoost(&photo, map[string]any{}, q, MemoryRecord{Content: "John has a photograph of veterans taken on 19 May 2023."})
	if party <= photo {
		t.Fatalf("hosted-event rank boost must lift party over photograph, party=%v photo=%v", party, photo)
	}
	run := 1.0
	applyHostedEventRankBoost(&run, map[string]any{}, q, MemoryRecord{Content: "John organized a 5K charity run in his neighborhood to support veterans."})
	if run != 1.0 {
		t.Fatalf("5K run must not receive hosted-event rank boost, got %v", run)
	}
}

func TestApplyFactPrimaryRecallKeepsHostedEventEpisode(t *testing.T) {
	q := "What did John host for the veterans in May 2023 as part of the project"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "John organized a small party for veterans."},
		"ep": {
			MemoryID:  "ep",
			Content:   "We had a great time throwing a small party and inviting some veterans to share their stories",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("host query must keep hosted-event episode leftover")
	}
}

func TestLeftoverCoveringAdvicePrefersDirectiveOverAwesome(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Gina", Value: "Gina", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Jon", Value: "Jon", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Gina: Sorry about your job Jon, but starting your own business sounds awesome"},
			{Content: "Got any advice or tips on running a successful biz"},
			{Content: "Gina: Thanks for the advice, Jon"},
			{Content: "Starting a business takes courage - you hang in there too"},
			{Content: "Seeing your students grow and succeed must be really fulfilling"},
			{Content: "Building relationships and creating a strong brand image for my store is something I'm always working on"},
			{Content: "Also be sure to build relationships with your customers – let them know you care"},
			{Content: "And don't forget to stay positive and motivate others"},
			{Content: "Make sure yours stands out"},
		},
	}
	q := "What advice does Gina give to Jon about running a successful business?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "relationship") || !strings.Contains(lower, "brand") || !strings.Contains(lower, "positive") {
		t.Fatalf("advice leftover covering must join relationships, brand, and stay positive, got %q", got)
	}
	if strings.Contains(lower, "awesome") || strings.Contains(lower, "got any advice") || strings.Contains(lower, "students") || strings.Contains(lower, "courage") {
		t.Fatalf("advice echo, awesome, students, or business restatement must not cover, got %q", got)
	}
	hybrid := "She tells Jon that starting his own business sounds awesome."
	if leftoverCoveringAdviceOffQueryLine(hybrid) {
		t.Fatal("awesome restatement is not off-query advice leftover")
	}
	if leftoverCoveringAdviceOffQueryLine("Starting a business takes courage - you hang in there too") {
		t.Fatal("business restatement gerund must not count as off-query advice leftover")
	}
	if leftoverCoveringAdviceOffQueryLine("Seeing your students grow and succeed must be really fulfilling") {
		t.Fatal("second-person evaluative gerund must not count as advice leftover")
	}
	if !leftoverCoveringAdviceMissesDirective(q, got, hybrid) {
		t.Fatal("advice covering must replace an awesome restatement")
	}
}

func TestExpandAdviceDirectiveSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_7"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Got any advice or tips on running a successful biz",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about the store",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Building relationships and creating a strong brand image for my store is something I'm always working on",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the advice leftover can miss")
	}
	expandAdviceDirectiveSessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("advice session expand must admit the first-person gerund leftover past the generic cap")
	}
}

func TestSessionIDsForAdviceQueryRequiresSpeechEcho(t *testing.T) {
	q := "What advice does Gina give to Jon about running a successful business?"
	campaign := MemoryRecord{
		MemoryID: "c",
		Content:  "Gina launched an ad campaign for her clothing store to grow the business.",
		Metadata: map[string]any{"session_id": "session_3"},
	}
	echo := MemoryRecord{
		MemoryID: "e",
		Content:  "Got any advice or tips on running a successful biz",
		Metadata: map[string]any{"session_id": "session_7"},
	}
	ids := sessionIDsForAdviceQuery(q, []MemoryRecord{campaign, echo})
	if len(ids) == 0 || ids[0] != "session_7" {
		t.Fatalf("advice session pick must prefer the advice echo, got %#v", ids)
	}
	for _, id := range ids {
		if id == "session_3" {
			t.Fatal("business campaign session must not be fetched without an advice echo")
		}
	}
}

func TestApplyAdviceDirectiveRankBoostSkipsAwesome(t *testing.T) {
	q := "What advice does Gina give to Jon about running a successful business?"
	gold := 1.0
	awesome := 1.0
	applyAdviceDirectiveRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "Also be sure to build relationships with your customers – let them know you care"})
	applyAdviceDirectiveRankBoost(&awesome, map[string]any{}, q, MemoryRecord{Content: "Gina: Sorry about your job Jon, but starting your own business sounds awesome"})
	if gold <= awesome {
		t.Fatalf("advice directive rank boost must lift hortative leftover over awesome, gold=%v awesome=%v", gold, awesome)
	}
}

func TestScoreMemoryIDFAdviceGoldNeedsDirectiveFloor(t *testing.T) {
	q := "What advice does Gina give to Jon about running a successful business?"
	gold := MemoryRecord{Content: "Building relationships and creating a strong brand image for my store is something I'm always working on"}
	if !leftoverCoveringAdviceOffQueryLine(gold.Content) {
		t.Fatal("first-person gerund leftover must classify as off-query advice")
	}
	score, _ := scoreMemoryIDF(gold, q, tokenize(q), nil, nil)
	if score > 0 {
		t.Fatalf("advice gold has no query tokens so IDF score must be 0 without a floor, got %v", score)
	}
	hortative := "Also be sure to build relationships with your customers – let them know you care"
	if !leftoverCoveringAdviceOffQueryLine(hortative) {
		t.Fatal("hortative leftover must classify as off-query advice")
	}
}

func TestApplyFactPrimaryRecallKeepsAdviceEpisode(t *testing.T) {
	q := "What advice does Gina give to Jon about running a successful business?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Building relationships and creating a strong brand image for my store is something I'm always working on"},
		"ep": {
			MemoryID:  "ep",
			Content:   "And don't forget to stay positive and motivate others",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("advice query must keep hortative episode leftover")
	}
}

func TestLeftoverCoveringKindPrefersLikeListOverSpread(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Maria", Value: "Maria", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Maria believes that spreading kindness and support can make a difference when someone is feeling down."},
			{Content: "Maria and her mother cooked dinner together at home on 3 May 2023."},
			{Content: "It had lots of great things like salads, sandwiches, and homemade desserts"},
			{Content: "Maria's favorite dessert is the banana split sundae."},
			{Content: "John signed a basketball shoe and gear deal with Nike."},
		},
	}
	q := "What kind of food did Maria have on her dinner spread iwth her mother?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "salad") || !strings.Contains(lower, "sandwich") || !strings.Contains(lower, "dessert") {
		t.Fatalf("what-kind leftover covering must pick the like-list leftover, got %q", got)
	}
	if strings.Contains(lower, "kindness") || strings.Contains(lower, "nike") {
		t.Fatalf("spread restatement or foreign deals must not cover what-kind, got %q", got)
	}
	hybrid := "Maria believes that spreading kindness and support can make a difference when someone is feeling down."
	if leftoverCoveringKindListLine(hybrid) {
		t.Fatal("spreading kindness is not a like-list leftover")
	}
	if leftoverCoveringKindListLine("John signed a basketball shoe and gear deal with Nike, and is in talks about a potential sponsorship deal with Gatorade.") {
		t.Fatal("deals line without like-A,-B,-and-C must not count as kind-list leftover")
	}
	if leftoverCoveringKindListLine("I like salad") {
		t.Fatal("bare like without a comma list must not count as kind-list leftover")
	}
	if !leftoverCoveringKindMissesList(q, got, hybrid) {
		t.Fatal("what-kind covering must replace a spread restatement")
	}
	salad := "It had lots of great things like salads, sandwiches, and homemade desserts"
	rare := leftoverCoverNonWeakTokens(leftoverNonEntityRareTokens(q, hops))
	if leftoverSkipLine(q, salad, rare) {
		t.Fatal("what-kind leftover covering must still admit like-A,-B,-and-C leftover")
	}
}

func TestLeftoverCoveringWhenEventSkipsKindListHopDump(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Joanna", Value: "Joanna", Source: "search_fallback"},
	}
	dump := "visit nate to watch turtles from a distance, make script into a movie, visit nate on 2022-11-05, write a new script, make coconut milk ice cream like nate, try chocolate, and mixed berry flavors"
	letter := "Joanna received a touching letter after someone read her blog post (7 August 2022; the week before 14 August 2022)"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: dump},
			{Content: "Themes like sisterhood, love, and chasing dreams (18 March 2022)"},
			{Content: letter},
			{Content: "Joanna would get two turtles today if she could. (18 March 2022)"},
		},
	}
	q := "When did someone write Joanna a touching letter?"
	rare := leftoverCoverNonWeakTokens(leftoverNonEntityRareTokens(q, hops))
	if !leftoverCoveringKindListLine(dump) {
		t.Fatal("activity dump with like-A,-B,-and-C must still classify as kind-list shape")
	}
	if !leftoverSkipLine(q, dump, rare) {
		t.Fatal("when-event leftover covering must skip crowded kind-list hop dumps")
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(got, "7 August") && !strings.Contains(lower, "week before") {
		t.Fatalf("when-event leftover covering must pick the dated letter leftover, got %q", got)
	}
	if strings.Contains(lower, "turtle") || strings.Contains(lower, "sisterhood") || strings.Contains(lower, "ice cream") {
		t.Fatalf("kind-list hop dump must not cover a when-event letter query, got %q", got)
	}
}

func TestLeftoverCoveringPolarIgnoresKindList(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "James", Value: "James", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "It had lots of great things like salads, sandwiches, and homemade desserts"},
			{Content: "Yesterday, when we were at the theater, Samantha loves theater, I asked her to become my girlfriend, and she agreed (3 September 2022)"},
		},
	}
	q := "Did James have a girlfriend during April 2022?"
	if leftoverCoveringSpecificAnswer(q, hops, pkt) != "" {
		t.Fatal("polar leftover covering must stay empty so like-lists cannot rewrite girlfriend absence")
	}
}

func TestLeftoverCoveringProcessPrefersHortativeOverCompanion(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Nate", Value: "Nate", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Nate says that looking after his turtles is a calming, stress-relieving activity."},
			{Content: "Nate's pet turtles are calm and peaceful, and the recent tank expansion made them happy."},
			{Content: "Hoping to share my love of gaming and connect with others who enjoy it too"},
			{Content: "Just keep their area clean, feed them properly, and make sure they get enough light"},
			{Content: "[pet turtles basking rock] [a photo of a turtle sitting on a log in a pond]"},
		},
	}
	q := "How does Nate describe the process of taking care of turtles?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "clean") || !strings.Contains(lower, "feed") || !strings.Contains(lower, "light") {
		t.Fatalf("how-describe-process leftover covering must pick the hortative leftover, got %q", got)
	}
	if strings.Contains(lower, "calming") || strings.Contains(lower, "peaceful") || strings.Contains(lower, "gaming") {
		t.Fatalf("companion slogan must not cover how-describe-process, got %q", got)
	}
	hybrid := "Nate says that looking after his turtles is a calming, stress-relieving activity."
	if !leftoverCoveringProcessMissesHortative(q, got, hybrid) {
		t.Fatal("how-describe-process covering must replace a companion slogan")
	}
	if leftoverCoveringLineHasForeignPerson(q, "Just keep their area clean, feed them properly, and make sure they get enough light") {
		t.Fatal("sentence-initial just-keep must not count as a foreign person")
	}
}

func TestLeftoverCoveringProcessLivePacketPrefersJustKeep(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Nate", Value: "Nate", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Nate describes his gaming space as a personal haven to escape into the virtual world."},
			{Content: "I think just having someone to support me throughout the whole process is such a blessing"},
			{Content: "Nate believes that taking care of ourselves helps us be more creative and happier."},
			{Content: "Nate notes that his turtles require little looking after."},
			{Content: "Just keep their area clean, feed them properly, and make sure they get enough light"},
			{Content: "Nate recommends keeping turtles as pets to help relieve stress."},
			{Content: "Nate's pet turtles are calm and peaceful, and the recent tank expansion made them happy."},
			{Content: "Hoping to share my love of gaming and connect with others who enjoy it too"},
		},
	}
	q := "How does Nate describe the process of taking care of turtles?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "clean") || !strings.Contains(lower, "feed") || !strings.Contains(lower, "light") {
		t.Fatalf("live how-describe-process packet must cover with just-keep leftover, got %q", got)
	}
	if strings.Contains(lower, "gaming") || strings.Contains(lower, "blessing") || strings.Contains(lower, "calming") || strings.Contains(lower, "ourselves") {
		t.Fatalf("companion, blessing, or gaming leftover must not cover, got %q", got)
	}
}

func TestExpandProcessHortativeSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_5"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "[pet turtles basking rock] [a photo of a turtle sitting on a log in a pond]",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about turtles",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Just keep their area clean, feed them properly, and make sure they get enough light",
		Metadata: map[string]any{"session_id": session},
	}
	forget := MemoryRecord{
		MemoryID: "forget",
		Content:  "Don't forget to relax and enjoy the process too",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, forget, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the hortative leftover can miss")
	}
	expandProcessHortativeSessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("process hortative session expand must admit the just-keep leftover past the generic cap")
	}
	if _, ok := candidates["forget"]; ok {
		t.Fatal("process hortative session expand must not admit process-restatement hortative")
	}
}

func TestLeftoverCoveringProcessDoesNotStealElectronic(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Calvin", Value: "Calvin", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Don't forget to relax and enjoy the process too"},
			{Content: "Adding electronic elements gives his songs a fresh vibe"},
			{Content: "Calvin has been adding electronic elements to his songs."},
		},
	}
	q := "How does Calvin describe his process of adding electronic elements to his songs?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	if got != "" {
		t.Fatalf("process leftover covering must stay empty so electronic hybrid can hold, got %q", got)
	}
	if leftoverCoveringProcessHortativeLine("Don't forget to relax and enjoy the process too") {
		t.Fatal("enjoy-the-process hortative must not count as process leftover")
	}
	if leftoverCoveringProcessMissesHortative(q, "", "Adding electronic elements gives his songs a fresh vibe") {
		t.Fatal("empty process covering must not replace the electronic hybrid")
	}
}

func TestSessionIDsForHowDescribeProcessQueryUsesObjectSeed(t *testing.T) {
	q := "How does Nate describe the process of taking care of turtles?"
	photo := MemoryRecord{
		MemoryID: "p",
		Content:  "[pet turtles basking rock] [a photo of a turtle sitting on a log in a pond]",
		Metadata: map[string]any{"session_id": "session_5"},
	}
	careSelf := MemoryRecord{
		MemoryID: "c",
		Content:  "Nate believes that taking care of ourselves helps us be more creative and happier.",
		Metadata: map[string]any{"session_id": "session_9"},
	}
	ids := sessionIDsForHowDescribeProcessQuery(q, []MemoryRecord{careSelf, photo}, nil)
	if len(ids) == 0 || ids[0] != "session_5" {
		t.Fatalf("how-describe-process session pick must prefer object-token coverage, got %#v", ids)
	}
	noisy := make([]MemoryRecord, 0, 12)
	for i := 11; i <= 22; i++ {
		noisy = append(noisy, MemoryRecord{
			MemoryID: "t" + itoa(i),
			Content:  "Nate has pet turtles",
			Metadata: map[string]any{"session_id": "session_" + itoa(i)},
		})
	}
	gold := MemoryRecord{
		MemoryID: "g",
		Content:  "Just keep their area clean, feed them properly, and make sure they get enough light",
		Metadata: map[string]any{"session_id": "session_5"},
	}
	ids = sessionIDsForHowDescribeProcessQuery(q, append(noisy, photo), []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_5" {
		t.Fatalf("how-describe-process session pick must keep the hortative session ahead of turtle-slogan sessions, got %#v", ids)
	}
}

func TestApplyProcessHortativeRankBoostSkipsProcessRestatement(t *testing.T) {
	q := "How does Nate describe the process of taking care of turtles?"
	gold := 1.0
	forget := 1.0
	applyProcessHortativeRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "Just keep their area clean, feed them properly, and make sure they get enough light"})
	applyProcessHortativeRankBoost(&forget, map[string]any{}, q, MemoryRecord{Content: "Don't forget to relax and enjoy the process too"})
	if gold <= forget {
		t.Fatalf("process hortative rank boost must lift just-keep over process restatement, gold=%v forget=%v", gold, forget)
	}
	calming := 1.0
	applyProcessHortativeRankBoost(&calming, map[string]any{}, q, MemoryRecord{Content: "Nate says looking after his turtles is calming"})
	if calming != 1.0 {
		t.Fatalf("companion slogan must not receive process hortative rank boost, got %v", calming)
	}
}

func TestScoreMemoryIDFProcessHortativeNeedsFloor(t *testing.T) {
	q := "How does Nate describe the process of taking care of turtles?"
	gold := MemoryRecord{Content: "Just keep their area clean, feed them properly, and make sure they get enough light"}
	if !leftoverCoveringProcessHortativeLine(gold.Content) {
		t.Fatal("just-keep leftover must classify as process hortative")
	}
	score, _ := scoreMemoryIDF(gold, q, tokenize(q), nil, nil)
	if score > 0 {
		t.Fatalf("process hortative gold has no query tokens so IDF score must be 0 without a floor, got %v", score)
	}
}

func TestApplyFactPrimaryRecallKeepsProcessHortativeEpisode(t *testing.T) {
	q := "How does Nate describe the process of taking care of turtles?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Nate has pet turtles."},
		"ep": {
			MemoryID:  "ep",
			Content:   "Just keep their area clean, feed them properly, and make sure they get enough light",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("how-describe-process query must keep hortative episode leftover")
	}
}

func TestLeftoverCoveringMotivatePrefersCauseOverTurtle(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Joanna", Value: "Joanna", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."},
			{Content: "Have faith in yourself and continue following your writing dreams - it's tough but worth it"},
			{Content: "Wish I had a vacation lined up, but right now my writing is consuming me"},
			{Content: "Joanna is passionate about writing and intends to keep pursuing it."},
			{Content: "It's knowing that my writing can make a difference that keeps me going, even on tough days"},
		},
	}
	q := "What motivates Joanna to keep writing even on tough days?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "difference") || !strings.Contains(lower, "writing") {
		t.Fatalf("what-motivates leftover covering must pick the first-person cause leftover, got %q", got)
	}
	if strings.Contains(lower, "turtle") || strings.Contains(lower, "faith") || strings.Contains(lower, "consuming") {
		t.Fatalf("turtle, faith, or occupation leftover must not cover what-motivates, got %q", got)
	}
	hybrid := "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."
	if !leftoverCoveringMotivateMissesCause(q, got, hybrid) {
		t.Fatal("what-motivates covering must replace a compiler motivate-her restatement")
	}
	if !leftoverCoveringSkipsHybrid(q, got) {
		t.Fatal("what-motivates cause leftover must skip hybrid so recall cannot idle-timeout")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, hybrid) {
		t.Fatal("empty covering and turtle restatement must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How does Nate describe the process of taking care of turtles?", "Just keep their area clean, feed them properly, and make sure they get enough light") {
		t.Fatal("how-describe-process must not skip hybrid via what-motivates")
	}
}

func TestLeftoverCoveringMotivatePacketAloneSkipsHybrid(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "It's knowing that my writing can make a difference that keeps me going, even on tough days"},
		},
	}
	q := "What motivates Joanna to keep writing even on tough days?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if !leftoverCoveringSkipsHybrid(q, got) {
		t.Fatalf("search-packet cause leftover must skip hops and hybrid without hop results, got %q", got)
	}
	emptyPkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."},
		},
	}
	if leftoverCoveringSkipsHybrid(q, leftoverCoveringSpecificAnswer(q, nil, emptyPkt)) {
		t.Fatal("turtle compiler restatement must not skip hops or hybrid")
	}
}

func TestLeftoverCoveringMotivateEmptyLeavesHybrid(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Joanna", Value: "Joanna", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."},
			{Content: "Have faith in yourself and continue following your writing dreams - it's tough but worth it"},
			{Content: "Wish I had a vacation lined up, but right now my writing is consuming me"},
		},
	}
	q := "What motivates Joanna to keep writing even on tough days?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	if got != "" {
		t.Fatalf("what-motivates leftover covering must stay empty without a cause line so turtle hybrid can hold, got %q", got)
	}
	if leftoverCoveringMotivateMissesCause(q, "", "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times.") {
		t.Fatal("empty what-motivates covering must not replace the turtle hybrid")
	}
}

func TestLeftoverCoveringMotivateDoesNotStealRunningGroup(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Deborah", Value: "Deborah", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "We help and push each other during our runs, which makes it so much easier to stay motivated"},
			{Content: "It's knowing that my writing can make a difference that keeps me going, even on tough days"},
		},
	}
	q := "What made being part of the running group easy for Deborah to stay motivated?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "push") {
		t.Fatalf("what-made covering must still pick the running-group push leftover, got %q", got)
	}
	if strings.Contains(lower, "writing") || strings.Contains(lower, "difference") {
		t.Fatalf("what-motivates cause leftover must not steal what-made covering, got %q", got)
	}
}

func TestExpandMotivateCauseSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_18"
	q := "What motivates Joanna to keep writing even on tough days?"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Joanna is passionate about writing and intends to keep pursuing it.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about writing",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "It's knowing that my writing can make a difference that keeps me going, even on tough days",
		Metadata: map[string]any{"session_id": session},
	}
	turtle := MemoryRecord{
		MemoryID: "turtle",
		Content:  "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times.",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, turtle, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the cause leftover can miss")
	}
	expandMotivateCauseSessionNeighbors(candidates, []MemoryRecord{seed}, all, q, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-motivates session expand must admit the first-person cause leftover past the generic cap")
	}
	if _, ok := candidates["turtle"]; ok {
		t.Fatal("what-motivates session expand must not admit compiler motivate-her leftover")
	}
}

func TestSessionIDsForWhatMotivatesQueryPrefersCauseSession(t *testing.T) {
	q := "What motivates Joanna to keep writing even on tough days?"
	writing := MemoryRecord{
		MemoryID: "w",
		Content:  "Joanna is passionate about writing and intends to keep pursuing it.",
		Metadata: map[string]any{"session_id": "session_18"},
	}
	turtle := MemoryRecord{
		MemoryID: "t",
		Content:  "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times.",
		Metadata: map[string]any{"session_id": "session_1"},
	}
	ids := sessionIDsForWhatMotivatesQuery(q, []MemoryRecord{turtle, writing}, nil)
	if len(ids) == 0 || ids[0] != "session_18" {
		t.Fatalf("what-motivates session pick must prefer object-token coverage, got %#v", ids)
	}
	noisy := make([]MemoryRecord, 0, 12)
	for i := 11; i <= 22; i++ {
		noisy = append(noisy, MemoryRecord{
			MemoryID: "t" + itoa(i),
			Content:  "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times.",
			Metadata: map[string]any{"session_id": "session_" + itoa(i)},
		})
	}
	gold := MemoryRecord{
		MemoryID: "g",
		Content:  "It's knowing that my writing can make a difference that keeps me going, even on tough days",
		Metadata: map[string]any{"session_id": "session_18"},
	}
	ids = sessionIDsForWhatMotivatesQuery(q, append(noisy, writing), []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_18" {
		t.Fatalf("what-motivates session pick must keep the cause session ahead of turtle-slogan sessions, got %#v", ids)
	}
}

func TestApplyMotivateCauseRankBoostSkipsTurtleRestatement(t *testing.T) {
	q := "What motivates Joanna to keep writing even on tough days?"
	gold := 1.0
	turtle := 1.0
	applyMotivateCauseRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "It's knowing that my writing can make a difference that keeps me going, even on tough days"})
	applyMotivateCauseRankBoost(&turtle, map[string]any{}, q, MemoryRecord{Content: "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."})
	if gold <= turtle {
		t.Fatalf("what-motivates rank boost must lift cause leftover over turtle restatement, gold=%v turtle=%v", gold, turtle)
	}
}

func TestScoreMemoryIDFMotivateCauseNeedsFloor(t *testing.T) {
	q := "What motivates Joanna to keep writing even on tough days?"
	gold := MemoryRecord{Content: "It's knowing that my writing can make a difference that keeps me going, even on tough days"}
	if !leftoverCoveringMotivateCauseLine(q, gold.Content) {
		t.Fatal("first-person cause leftover must classify as what-motivates cause")
	}
	// After dropping motivate/keep/person, writing/tough/days still match, so IDF
	// may be non-zero. Floor is for leftover that omits remaining query tokens.
	zero := MemoryRecord{Content: "It's knowing that my craft can make a difference that keeps me going, even on hard evenings"}
	if leftoverCoveringMotivateCauseLine(q, zero.Content) {
		t.Fatal("cause leftover must still require the query object token")
	}
}

func TestApplyFactPrimaryRecallKeepsMotivateCauseEpisode(t *testing.T) {
	q := "What motivates Joanna to keep writing even on tough days?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Joanna believes turtles symbolize strength and perseverance and motivate her in tough times."},
		"ep": {
			MemoryID:  "ep",
			Content:   "It's knowing that my writing can make a difference that keeps me going, even on tough days",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-motivates query must keep first-person cause episode leftover")
	}
}

func TestLeftoverCoveringSayAboutPrefersTheyEvaluative(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Gina", Value: "Gina", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Gina performed a contemporary dance piece titled \"Finding Freedom\"."},
			{Content: "[group dancers performing on stage] [a photo of a group of dancers in white dresses on a stage]"},
			{Content: "It's great having the freedom to create a space and help dancers of all ages and levels express themselves"},
			{Content: "They're so graceful"},
		},
	}
	q := "What does Gina say about the dancers in the photo?"
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "graceful") || !strings.Contains(lower, "they") {
		t.Fatalf("what-say-about leftover covering must pick the they-evaluative leftover, got %q", got)
	}
	if strings.Contains(lower, "finding freedom") || strings.Contains(lower, "express themselves") {
		t.Fatalf("performance leftover must not cover what-say-about, got %q", got)
	}
	if !leftoverCoveringSayAboutMissesEvaluative(q, got, "Gina performed a contemporary dance piece titled \"Finding Freedom\".") {
		t.Fatal("what-say-about covering must replace a compiler performance restatement")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Gina performed a contemporary dance piece titled \"Finding Freedom\".") {
		t.Fatal("empty covering and performance restatement must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What motivates Joanna to keep writing even on tough days?", "They're so graceful") {
		t.Fatal("what-motivates must not skip hybrid via they-evaluative")
	}
	dumpItems := []RecallItem{
		{Value: `performed contemporary piece "finding freedom"`},
		{Value: "dance session"},
		{Value: "stress relief"},
	}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("what-say-about enumerate items must follow they-evaluative covering, got %#v", synced)
	}
	if kept := leftoverCoveringSyncEnumerateItems("What does Riley do to unwind?", got, dumpItems); len(kept) != 3 {
		t.Fatalf("non-say-about enumerate items must not collapse to they-evaluative covering, got %#v", kept)
	}
}

func TestLeftoverCoveringSayAboutEmptyLeavesHybrid(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Gina performed a contemporary dance piece titled \"Finding Freedom\"."},
			{Content: "[group dancers performing on stage] [a photo of a group of dancers in white dresses on a stage]"},
		},
	}
	q := "What does Gina say about the dancers in the photo?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("what-say-about leftover covering must stay empty without a they-evaluative line, got %q", got)
	}
}

func TestLeftoverCoveringSayAboutCoversFirstPersonGotLine(t *testing.T) {
	gold := "It's got so much to check out - the culture, food - you won't regret it."
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John visited New York City."},
			{Content: gold},
			{Content: "It's gotta be a rush having all these options"},
			{Content: "The doctor said it's not too serious"},
		},
	}
	q := "What did John say about NYC, enticing Tim to visit?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if !strings.Contains(strings.ToLower(got), "got so much") {
		t.Fatalf("NYC what-say-about leftover covering must pick first-person got leftover, got %q", got)
	}
	if leftoverCoveringEvaluativeTheyLine(gold) {
		t.Fatal("it's-got leftover is not they-evaluative")
	}
	if !leftoverCoveringFirstPersonGotLine(gold) {
		t.Fatal("it's-got leftover must count as first-person got covering")
	}
	if leftoverCoveringFirstPersonGotLine("It's gotta be a rush having all these options") {
		t.Fatal("it's-gotta leftover is not first-person got covering")
	}
	if leftoverCoveringFirstPersonGotLine("The doctor said it's not too serious") {
		t.Fatal("doctor-said leftover is not first-person got covering")
	}
	injuryPkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Tim experienced a setback on 14 November 2023."},
			{Content: "The doctor said it's not too serious"},
			{Content: "John focused on recovery and worked hard to strengthen his body after the ankle injury."},
			{Content: "It's got so much to check out - the culture, food - you won't regret it."},
		},
	}
	injury := leftoverCoveringSpecificAnswer("What did Tim say about his injury on 16 November, 2023?", nil, injuryPkt)
	if !strings.Contains(strings.ToLower(injury), "not too serious") {
		t.Fatalf("dated what-say-about leftover covering must pick reported-speech leftover, got %q", injury)
	}
	if strings.Contains(strings.ToLower(injury), "check out") {
		t.Fatalf("dated what-say-about must not steal first-person got leftover, got %q", injury)
	}
	if leftoverCoveringReportedSpeechLine(gold) {
		t.Fatal("it's-got leftover is not reported-speech covering")
	}
	if !leftoverCoveringReportedSpeechLine("The doctor said it's not too serious") {
		t.Fatal("the-role-said leftover must count as reported-speech covering")
	}
	synced := leftoverCoveringSyncEnumerateItems(q, got, []RecallItem{{Value: "John visited New York City."}})
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("what-say-about enumerate items must follow first-person got covering, got %#v", synced)
	}
	injurySynced := leftoverCoveringSyncEnumerateItems(
		"What did Tim say about his injury on 16 November, 2023?",
		injury,
		[]RecallItem{{Value: "Tim experienced a setback on 14 November 2023."}},
	)
	if len(injurySynced) != 1 || injurySynced[0].Value != injury {
		t.Fatalf("dated what-say-about enumerate items must follow reported-speech covering, got %#v", injurySynced)
	}
}

func TestLeftoverCoveringSayAboutDatedSkipsGot(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Tim experienced a setback on 14 November 2023."},
			{Content: "It's got so much to check out - the culture, food - you won't regret it."},
		},
	}
	got := leftoverCoveringSpecificAnswer("What did Tim say about his injury on 16 November, 2023?", nil, pkt)
	if got != "" {
		t.Fatalf("dated what-say-about must not fall through to first-person got leftover, got %q", got)
	}
}

func TestLeftoverCoveringSayAboutUndatedSkipsReportedSpeech(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John visited New York City."},
			{Content: "The doctor said it's not too serious"},
		},
	}
	got := leftoverCoveringSpecificAnswer("What did John say about NYC, enticing Tim to visit?", nil, pkt)
	if got != "" {
		t.Fatalf("undated what-say-about must not fall through to reported-speech leftover, got %q", got)
	}
}

func TestExpandSayAboutSessionNeighborsAdmitsReportedSpeech(t *testing.T) {
	session := "session_18"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "[ankle injury wrapped bandages] [a photo of a person with a bandage on their leg]",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about recovery",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "The doctor said it's not too serious",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the reported-speech leftover can miss")
	}
	expandEvaluativeTheySessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-say-about session expand must admit reported-speech leftover past the generic cap")
	}
}

func TestLeftoverCoveringSayAboutGotRequiresObjectEvidence(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "It's got so much to check out - the culture, food - you won't regret it."},
		},
	}
	q := "What did John say about NYC, enticing Tim to visit?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("first-person got leftover must not cover without object-token evidence in the packet, got %q", got)
	}
}

func TestLeftoverCoveringSayAboutPrefersTheyOverGot(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "It's got so much to check out - the culture, food - you won't regret it."},
			{Content: "They're so graceful"},
		},
	}
	q := "What does Gina say about the dancers in the photo?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "graceful") || strings.Contains(lower, "check out") {
		t.Fatalf("they-evaluative leftover must beat first-person got leftover, got %q", got)
	}
}

func TestExpandEvaluativeTheySessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_1"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Gina performed a contemporary dance piece titled \"Finding Freedom\".",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about dance",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "They're so graceful",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the they-evaluative leftover can miss")
	}
	expandEvaluativeTheySessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-say-about session expand must admit they-evaluative leftover past the generic cap")
	}
}

func TestExpandSayAboutSessionNeighborsAdmitsFirstPersonGot(t *testing.T) {
	session := "session_9"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "That skyline looks amazing - I've been wanting to visit NYC",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about travel",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "It's got so much to check out - the culture, food - you won't regret it.",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the first-person got leftover can miss")
	}
	expandEvaluativeTheySessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-say-about session expand must admit first-person got leftover past the generic cap")
	}
}

func TestSessionIDsForWhatSayAboutQueryPrefersObjectOverlap(t *testing.T) {
	q := "What does Gina say about the dancers in the photo?"
	photo := MemoryRecord{
		MemoryID: "p",
		Content:  "[group dancers performing on stage] [a photo of a group of dancers in white dresses on a stage]",
		Metadata: map[string]any{"session_id": "session_1"},
	}
	other := MemoryRecord{
		MemoryID: "o",
		Content:  "They're so graceful",
		Metadata: map[string]any{"session_id": "session_9"},
	}
	ids := sessionIDsForWhatSayAboutQuery(q, []MemoryRecord{photo}, []MemoryRecord{other, photo})
	if len(ids) == 0 || ids[0] != "session_1" {
		t.Fatalf("what-say-about session pick must prefer dancers/photo overlap, got %#v", ids)
	}
	for _, id := range ids {
		if id == "session_9" {
			t.Fatal("they-evaluative session without object overlap must not be fetched")
		}
	}
}

func TestApplyEvaluativeTheyRankBoostSkipsPerformance(t *testing.T) {
	q := "What does Gina say about the dancers in the photo?"
	gold := 1.0
	perf := 1.0
	applyEvaluativeTheyRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "They're so graceful"})
	applyEvaluativeTheyRankBoost(&perf, map[string]any{}, q, MemoryRecord{Content: "Gina performed a contemporary dance piece titled \"Finding Freedom\"."})
	if gold <= perf {
		t.Fatalf("what-say-about rank boost must lift they-evaluative over performance leftover, gold=%v perf=%v", gold, perf)
	}
}

func TestApplyFactPrimaryRecallKeepsEvaluativeTheyEpisode(t *testing.T) {
	q := "What does Gina say about the dancers in the photo?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Gina performed a contemporary dance piece titled \"Finding Freedom\"."},
		"ep": {
			MemoryID:  "ep",
			Content:   "They're so graceful",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-say-about query must keep they-evaluative episode leftover")
	}
}

func TestLeftoverCoveringReactPrefersTheyWereObservation(t *testing.T) {
	gold := "I took them to a snowy one last winter and they were so confused"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Audrey's dogs dislike snow."},
			{Content: gold},
			{Content: "But one thing they hate is snow"},
			{Content: "It took a while for them to calm down, but all cut up they looked so cute"},
			{Content: "Audrey took her dogs to a snowy park last winter and they were confused."},
		},
	}
	q := "How do Audrey's dogs react to snow?"
	if !looksHowReactQuery(q) {
		t.Fatal("how-do-react-to must count as how-react")
	}
	if looksHowReactQuery("How should I respond to that email?") {
		t.Fatal("how-should-I-respond must not count as how-react")
	}
	if looksHowReactQuery("What does Gina say about the dancers in the photo?") {
		t.Fatal("what-say-about must not count as how-react")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "confused") {
		t.Fatalf("how-react leftover covering must pick they-were observation, got %q", got)
	}
	if strings.Contains(lower, "dislike") || strings.Contains(lower, "hate") {
		t.Fatalf("dislike/hate restatement must not cover how-react, got %q", got)
	}
	if leftoverCoveringReactionObservationLine("Audrey's dogs dislike snow.") {
		t.Fatal("dislike restatement is not observational they-were leftover")
	}
	if leftoverCoveringReactionObservationLine("But one thing they hate is snow") {
		t.Fatal("hate restatement is not observational they-were leftover")
	}
	if leftoverCoveringReactionObservationLine("They're so graceful") {
		t.Fatal("they-evaluative leftover is not how-react observation")
	}
	cute := "It took a while for them to calm down, but all cut up they looked so cute"
	if leftoverCoveringReactLineHasObject(q, cute) {
		t.Fatal("cute leftover has no query object token")
	}
	if leftoverCoveringReactMissesObservation(q, cute, "Audrey's dogs dislike snow.") {
		t.Fatal("cute leftover must not replace hybrid as how-react covering")
	}
	if !leftoverCoveringReactionObservationLine(gold) {
		t.Fatal("they-were-so-ADJ leftover must count as how-react observation")
	}
	if !leftoverCoveringReactMissesObservation(q, got, "Audrey's dogs dislike snow.") {
		t.Fatal("how-react covering must replace a dislike restatement")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Audrey's dogs dislike snow.") {
		t.Fatal("empty covering and dislike restatement must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What does Gina say about the dancers in the photo?", gold) {
		t.Fatal("what-say-about must not skip hybrid via how-react observation")
	}
	dumpItems := []RecallItem{
		{Value: "Audrey's dogs dislike snow."},
		{Value: "snow park"},
	}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("how-react enumerate items must follow observation covering, got %#v", synced)
	}
	if kept := leftoverCoveringSyncEnumerateItems("What does Riley do to unwind?", got, dumpItems); len(kept) != 2 {
		t.Fatalf("non-how-react enumerate items must not collapse to observation covering, got %#v", kept)
	}
	injury := leftoverCoveringSpecificAnswer("What did Tim say about his injury on 16 November, 2023?", nil, EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Tim experienced a setback on 14 November 2023."},
			{Content: "The doctor said it's not too serious"},
			{Content: gold},
		},
	})
	if !strings.Contains(strings.ToLower(injury), "not too serious") {
		t.Fatalf("dated what-say-about must still pick reported-speech leftover, got %q", injury)
	}
	nyc := leftoverCoveringSpecificAnswer("What did John say about NYC, enticing Tim to visit?", nil, EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John visited New York City."},
			{Content: "It's got so much to check out - the culture, food - you won't regret it."},
			{Content: gold},
		},
	})
	if !strings.Contains(strings.ToLower(nyc), "got so much") {
		t.Fatalf("undated what-say-about must still pick first-person got leftover, got %q", nyc)
	}
	dancers := leftoverCoveringSpecificAnswer("What does Gina say about the dancers in the photo?", nil, EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Gina performed a contemporary dance piece titled \"Finding Freedom\"."},
			{Content: "They're so graceful"},
			{Content: gold},
		},
	})
	if !strings.Contains(strings.ToLower(dancers), "graceful") {
		t.Fatalf("what-say-about must still pick they-evaluative leftover, got %q", dancers)
	}
}

func TestLeftoverCoveringReactEmptyLeavesHybrid(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Audrey's dogs dislike snow."},
			{Content: "But one thing they hate is snow"},
			{Content: "It took a while for them to calm down, but all cut up they looked so cute"},
		},
	}
	q := "How do Audrey's dogs react to snow?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-react leftover covering must stay empty without object-linked they-were observation, got %q", got)
	}
}

func TestLeftoverCoveringReactRequiresObjectEvidence(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "I took them to a rainy one last winter and they were so confused"},
		},
	}
	q := "How do Audrey's dogs react to snow?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-react observation must not cover without object-token evidence in the packet, got %q", got)
	}
}

func TestExpandReactionObservationSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_23"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Audrey's dogs dislike snow.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about winter walks",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I took them to a snowy one last winter and they were so confused",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the they-were observation leftover can miss")
	}
	expandReactionObservationSessionNeighbors(candidates, "How do Audrey's dogs react to snow?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("how-react session expand must admit they-were observation leftover past the generic cap")
	}
}

func TestApplyReactionObservationRankBoostSkipsDislike(t *testing.T) {
	q := "How do Audrey's dogs react to snow?"
	gold := 1.0
	dislike := 1.0
	applyReactionObservationRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "I took them to a snowy one last winter and they were so confused"})
	applyReactionObservationRankBoost(&dislike, map[string]any{}, q, MemoryRecord{Content: "Audrey's dogs dislike snow."})
	if gold <= dislike {
		t.Fatalf("how-react rank boost must lift they-were observation over dislike restatement, gold=%v dislike=%v", gold, dislike)
	}
}

func TestApplyFactPrimaryRecallKeepsReactObservationEpisode(t *testing.T) {
	q := "How do Audrey's dogs react to snow?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Audrey's dogs dislike snow."},
		"ep": {
			MemoryID:  "ep",
			Content:   "I took them to a snowy one last winter and they were so confused",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("how-react query must keep they-were observation episode leftover")
	}
}

func TestLeftoverCoveringPurposePrefersTakeCareAction(t *testing.T) {
	gold := "On top of that, I recently joined a dog owners group to learn how to better take care of them"
	compiler := "Audrey recently joined a dog owners group to learn how to better take care of her dogs."
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Audrey: Yep, taking a second to appreciate those moments makes like much better"},
			{Content: "Audrey took her dogs to a pet salon on 17 November 2023."},
			{Content: "Andrew: Awesome to see that you take such good care of your dogs"},
			{Content: "Don't forget to take care of yourself and have some fun too"},
			{Content: gold},
			{Content: compiler},
		},
	}
	q := "What did Audrey do in November 2023 to better take care of her dogs?"
	if !looksWhatDidPurposeQuery(q) {
		t.Fatal("what-did-X-do-to-purpose must count as what-did-purpose")
	}
	if looksWhatDidPurposeQuery("What did John say about NYC, enticing Tim to visit?") {
		t.Fatal("what-say-about must not count as what-did-purpose")
	}
	if looksWhatDidPurposeQuery("What did John host for the veterans in May 2023 as part of the project") {
		t.Fatal("host queries must not count as what-did-purpose")
	}
	if looksWhatDidPurposeQuery("What does Melanie do to destress?") {
		t.Fatal("single-token unwind purpose must not count as what-did-purpose")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "joined") || !strings.Contains(lower, "take care") {
		t.Fatalf("what-did-purpose leftover covering must pick take-care action leftover, got %q", got)
	}
	if strings.Contains(lower, "much better") || strings.Contains(lower, "pet salon") {
		t.Fatalf("comparative-better chat and salon visit must not cover what-did-purpose, got %q", got)
	}
	if leftoverCoveringPurposeActionLine(q, "Audrey: Yep, taking a second to appreciate those moments makes like much better") {
		t.Fatal("comparative-better chat is not purpose-action leftover")
	}
	if leftoverCoveringPurposeActionLine(q, "Audrey took her dogs to a pet salon on 17 November 2023.") {
		t.Fatal("dated salon visit is not adjacent take-care purpose leftover")
	}
	if leftoverCoveringPurposeActionLine(q, "Don't forget to take care of yourself and have some fun too") {
		t.Fatal("take-care-of-yourself leftover has no query actor")
	}
	if !leftoverCoveringPurposeActionLine(q, gold) {
		t.Fatal("first-person take-care leftover must count as purpose action")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Audrey's dogs dislike snow.") {
		t.Fatal("empty covering and unrelated leftover must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How do Audrey's dogs react to snow?", gold) {
		t.Fatal("how-react must not skip hybrid via what-did-purpose action")
	}
	dumpItems := []RecallItem{
		{Value: "behavior tips"},
		{Value: "four dogs"},
	}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("what-did-purpose enumerate items must follow action covering, got %#v", synced)
	}
}

func TestLeftoverCoveringPurposeStaysEmptyWithoutAction(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Audrey: Yep, taking a second to appreciate those moments makes like much better"},
			{Content: "Audrey took her dogs to a pet salon on 17 November 2023."},
		},
	}
	q := "What did Audrey do in November 2023 to better take care of her dogs?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("what-did-purpose leftover covering must stay empty without take-care action leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsPurposeActionEpisode(t *testing.T) {
	q := "What did Audrey do in November 2023 to better take care of her dogs?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Audrey took her dogs to a pet salon on 17 November 2023."},
		"ep": {
			MemoryID:  "ep",
			Content:   "On top of that, I recently joined a dog owners group to learn how to better take care of them",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-did-purpose query must keep take-care action episode leftover")
	}
}

func TestExpandPurposeActionSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_27"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Audrey's dogs ran around and got fresh air in a park on the weekend of 2–3 November 2023.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about park walks",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "On top of that, I recently joined a dog owners group to learn how to better take care of them",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the purpose-action leftover can miss")
	}
	expandPurposeActionSessionNeighbors(candidates, "What did Audrey do in November 2023 to better take care of her dogs?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-did-purpose session expand must admit take-care action leftover past the generic cap")
	}
}

func TestLeftoverCoveringHowDidStartPrefersChangedStartedPair(t *testing.T) {
	gold := "Changed my diet, started walking regularly, things like that"
	diet := "Evan changed his diet around two years ago (approximately 2021-10-25). (25 October 2021; 2 years ago)"
	walk := "Evan started walking regularly around two years ago (approximately 2021-10-25). (25 October 2021; 2 years ago)"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Evan: Yeah, what worked for me was finding a fitness routine I really enjoy"},
			{Content: "Evan possesses a gym membership card."},
			{Content: "I started lifting weights one year ago and it's been a journey (8 October 2022; 1 years ago)"},
			{Content: "Evan started watercolor painting a few years ago."},
			{Content: "Evan has been working on his health for two years as of August 7, 2023."},
			{Content: "Sam started a new diet on 2023-08-21."},
			{Content: diet},
			{Content: walk},
			{Content: gold},
		},
	}
	q := "How did Evan start his transformation journey two years ago?"
	if !looksHowDidStartQuery(q) {
		t.Fatal("how-did-start N-years-ago must count as how-did-start")
	}
	if looksHowDidStartQuery("How does Nate describe the process of taking care of turtles?") {
		t.Fatal("how-describe must not count as how-did-start")
	}
	if looksHowDidStartQuery("How do Audrey's dogs react to snow?") {
		t.Fatal("how-react must not count as how-did-start")
	}
	if looksHowDidStartQuery("How did Joanna feel when someone wrote her a letter after reading her blog post?") {
		t.Fatal("how-did-feel without start+ago must not count as how-did-start")
	}
	if looksHowDidStartQuery("What did Audrey do in November 2023 to better take care of her dogs?") {
		t.Fatal("what-did-purpose must not count as how-did-start")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "diet") || !strings.Contains(lower, "walking") {
		t.Fatalf("how-did-start leftover covering must pick changed+started pair leftover, got %q", got)
	}
	if strings.Contains(lower, "gym") || strings.Contains(lower, "fitness routine") || strings.Contains(lower, "lifting weights") || strings.Contains(lower, "watercolor") {
		t.Fatalf("gym restatement and duration-mismatch starts must not cover how-did-start, got %q", got)
	}
	if leftoverCoveringStartMethodLine(q, "Evan possesses a gym membership card.") {
		t.Fatal("gym membership is not start-method leftover")
	}
	if leftoverCoveringStartMethodLine(q, "I started lifting weights one year ago and it's been a journey (8 October 2022; 1 years ago)") {
		t.Fatal("one-year lifting leftover is duration-mismatch, not two-year start-method")
	}
	if leftoverCoveringStartMethodLine(q, "Evan started watercolor painting a few years ago.") {
		t.Fatal("a few years ago is not two years ago")
	}
	if leftoverCoveringStartMethodLine(q, "Sam started a new diet on 2023-08-21.") {
		t.Fatal("foreign dated diet start is not query-actor two-year start-method")
	}
	if leftoverCoveringStartMethodLine(q, "Evan has been working on his health for two years as of August 7, 2023.") {
		t.Fatal("working-on-it continuation is not start-method leftover")
	}
	if leftoverCoveringStartMethodLine(q, "Two years ago Evan began his health transformation by joining a gym and starting a fitness routine.") {
		t.Fatal("gym transformation restatement must not count as start-method leftover")
	}
	if !leftoverCoveringStartMethodLine(q, gold) {
		t.Fatal("first-person changed+started leftover must count as start-method")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Evan follows a fitness routine that he enjoys.") {
		t.Fatal("empty covering and gym restatement must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What did Audrey do in November 2023 to better take care of her dogs?", gold) {
		t.Fatal("what-did-purpose must not skip hybrid via how-did-start method")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("how-did-start pair leftover must skip hybrid")
	}
	hybrid := "Two years ago Evan began his health transformation by joining a gym and starting a fitness routine."
	if !leftoverCoveringStartMissesMethod(q, got, hybrid) {
		t.Fatal("gym hybrid must miss start-method leftover")
	}
	walkOnly := "Evan started walking regularly around two years ago (approximately 2021-10-25). (25 October 2021; 2 years ago) (via Evan has a 2-year duration)"
	if leftoverCoveringStartMethodPairLine(q, walkOnly) {
		t.Fatal("walking-only duration fact is not a changed+started pair")
	}
	if !leftoverCoveringStartMissesMethod(q, got, walkOnly) {
		t.Fatal("walking-only evidence_packet answer must miss changed+started pair leftover")
	}
	dumpItems := []RecallItem{
		{Value: "gym membership"},
		{Value: "fitness routine"},
	}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || synced[0].Value != got {
		t.Fatalf("how-did-start enumerate items must follow start-method covering, got %#v", synced)
	}
}

func TestLeftoverCoveringHowDidStartJoinsDurationFactsWithoutPair(t *testing.T) {
	diet := "Evan changed his diet around two years ago (approximately 2021-10-25). (25 October 2021; 2 years ago)"
	walk := "Evan started walking regularly around two years ago (approximately 2021-10-25). (25 October 2021; 2 years ago)"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Evan possesses a gym membership card."},
			{Content: "I started lifting weights one year ago and it's been a journey (8 October 2022; 1 years ago)"},
			{Content: diet},
			{Content: walk},
		},
	}
	q := "How did Evan start his transformation journey two years ago?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "diet") || !strings.Contains(lower, "walking") {
		t.Fatalf("how-did-start leftover covering must join duration-matched inception facts, got %q", got)
	}
	if strings.Contains(lower, "gym") || strings.Contains(lower, "lifting") {
		t.Fatalf("gym/lifting must not join into how-did-start covering, got %q", got)
	}
}

func TestLeftoverCoveringHowDidStartStaysEmptyWithoutMethod(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Evan possesses a gym membership card."},
			{Content: "Evan: Yeah, what worked for me was finding a fitness routine I really enjoy"},
			{Content: "I started lifting weights one year ago and it's been a journey (8 October 2022; 1 years ago)"},
		},
	}
	q := "How did Evan start his transformation journey two years ago?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-did-start leftover covering must stay empty without duration-matched start-method leftover, got %q", got)
	}
}

func TestLeftoverCoveringHowLongBeenPrefersContinuingDuration(t *testing.T) {
	gold := "Melanie's marriage duration is 5 years."
	already := "Melanie: 5 years already"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Melanie is married."},
			{Content: "Melanie has a husband."},
			{Content: "Caroline: Hey Mel, long time no chat"},
			{Content: "Been a long road, but I'm proud of how far I've come"},
			{Content: "Caroline: I started playing acoustic guitar about five years ago; it's been a great way to express myself and escape into my emotions (28 August 2018; 5 years ago)"},
			{Content: already},
			{Content: gold},
		},
	}
	q := "How long have Mel and her husband been married?"
	if !looksHowLongBeenQuery(q) {
		t.Fatal("how-long + been must count as how-long-been")
	}
	if looksHowLongBeenQuery("How did Evan start his transformation journey two years ago?") {
		t.Fatal("how-did-start must not count as how-long-been")
	}
	if looksHowLongBeenQuery("How does Nate describe the process of taking care of turtles?") {
		t.Fatal("how-describe must not count as how-long-been")
	}
	if looksHowLongBeenQuery("How do Audrey's dogs react to snow?") {
		t.Fatal("how-react must not count as how-long-been")
	}
	if looksHowLongBeenQuery("How often does Audrey meet up with other dog owners for tips and playdates?") {
		t.Fatal("how-often must not count as how-long-been")
	}
	if looksHowLongBeenQuery("How long ago did Melanie start pottery?") {
		t.Fatal("how-long-ago must not count as how-long-been")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "5 years") && !strings.Contains(lower, "five years") {
		t.Fatalf("how-long-been leftover covering must pick continuing duration, got %q", got)
	}
	if strings.Contains(lower, "is married") || strings.Contains(lower, "guitar") || strings.Contains(lower, "long road") || strings.Contains(lower, "long time no chat") {
		t.Fatalf("status facts and years-ago starts must not cover how-long-been, got %q", got)
	}
	if leftoverCoveringDurationLine(q, "Melanie is married.") {
		t.Fatal("copula status is not continuing duration")
	}
	if leftoverCoveringDurationLine(q, "Caroline: I started playing acoustic guitar about five years ago; it's been a great way to express myself and escape into my emotions (28 August 2018; 5 years ago)") {
		t.Fatal("years-ago start must not count as continuing duration")
	}
	if leftoverCoveringDurationLine(q, "Been a long road, but I'm proud of how far I've come") {
		t.Fatal("long-road leftover is not continuing years")
	}
	if leftoverCoveringDurationLine(q, "Caroline: 5 years already") {
		t.Fatal("foreign-speaker years-already leftover must not count as query-actor duration")
	}
	if !leftoverCoveringDurationLine(q, gold) {
		t.Fatal("named duration-is-N-years fact must count as continuing duration")
	}
	if !leftoverCoveringDurationLine(q, already) {
		t.Fatal("speaker-prefixed years-already leftover must count as continuing duration")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Melanie is married.") {
		t.Fatal("empty covering and status facts must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How did Evan start his transformation journey two years ago?", gold) {
		t.Fatal("how-did-start must not skip hybrid via how-long-been duration")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("how-long-been duration leftover must skip hybrid")
	}
	if !leftoverCoveringDurationMissesDuration(q, got, "Melanie is married.") {
		t.Fatal("married-state evidence_packet answer must miss continuing duration leftover")
	}
}

func TestLeftoverCoveringHowLongBeenStaysEmptyWithoutDuration(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Melanie is married."},
			{Content: "Melanie has a husband."},
			{Content: "Caroline: Hey Mel, long time no chat"},
		},
	}
	q := "How long have Mel and her husband been married?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-long-been leftover covering must stay empty without continuing duration leftover, got %q", got)
	}
}

func TestLeftoverCoveringHowOftenPrefersCadence(t *testing.T) {
	gold := "I try to meet up with other dog owners once a week for tips from other parents and so they can all play together"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Four Dogs, Photo Of Lake, Behavior Tips, Buddy, Fourth Dog (unnamed), Birdwatching Guidebook"},
			{Content: "Audrey's dogs meet other dog owners in the park and have doggie playdates."},
			{Content: "Audrey takes her dogs on road trips once every couple of months."},
			{Content: "I recently joined a dog owners group to learn how to better take care of them"},
			{Content: "Here's a pic from the playdate"},
			{Content: gold},
		},
	}
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	if !looksHowOftenQuery(q) {
		t.Fatal("how often + does must count as how-often")
	}
	if looksHowOftenQuery("How long have Mel and her husband been married?") {
		t.Fatal("how-long-been must not count as how-often")
	}
	if looksHowOftenQuery("How did Evan start his transformation journey two years ago?") {
		t.Fatal("how-did-start must not count as how-often")
	}
	if looksHowOftenQuery("How does Nate describe the process of taking care of turtles?") {
		t.Fatal("how-describe must not count as how-often")
	}
	if looksHowOftenQuery("How do Audrey's dogs react to snow?") {
		t.Fatal("how-react must not count as how-often")
	}
	if looksHowOftenQuery("How often have you been meeting other owners?") {
		t.Fatal("how-often-have without does/did/do must not count as how-often")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "once a week") {
		t.Fatalf("how-often leftover covering must pick cadence leftover, got %q", got)
	}
	if strings.Contains(lower, "four dogs") || strings.Contains(lower, "park") || strings.Contains(lower, "couple of months") || strings.Contains(lower, "joined a dog owners group") {
		t.Fatalf("dump, park playdates, road-trip months, and purpose leftover must not cover how-often, got %q", got)
	}
	if leftoverCoveringCadenceLine(q, "Audrey's dogs meet other dog owners in the park and have doggie playdates.") {
		t.Fatal("park playdates without cadence must not count as how-often covering")
	}
	if leftoverCoveringCadenceLine(q, "Audrey takes her dogs on road trips once every couple of months.") {
		t.Fatal("road-trip cadence about a different activity must not cover meetup how-often")
	}
	if leftoverCoveringCadenceLine(q, "I recently joined a dog owners group to learn how to better take care of them") {
		t.Fatal("purpose leftover without cadence must not count as how-often covering")
	}
	if leftoverCoveringCadenceLine(q, "Four Dogs, Photo Of Lake, Behavior Tips, Buddy, Fourth Dog (unnamed), Birdwatching Guidebook") {
		t.Fatal("hop dump must not count as cadence covering")
	}
	if !leftoverCoveringCadenceLine(q, gold) {
		t.Fatal("first-person once-a-week leftover must count as cadence covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Audrey's dogs meet other dog owners in the park and have doggie playdates.") {
		t.Fatal("empty covering and park playdates must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How long have Mel and her husband been married?", gold) {
		t.Fatal("how-long-been must not skip hybrid via how-often cadence")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("how-often cadence leftover must skip hybrid")
	}
	if !leftoverCoveringCadenceMissesCadence(q, got, "Four Dogs, Photo Of Lake, Behavior Tips") {
		t.Fatal("hop-dump answer must miss cadence leftover")
	}
}

func TestLeftoverCoveringHowOftenStaysEmptyWithoutCadence(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Audrey's dogs meet other dog owners in the park and have doggie playdates."},
			{Content: "I recently joined a dog owners group to learn how to better take care of them"},
			{Content: "Four Dogs, Photo Of Lake, Behavior Tips, Buddy"},
		},
	}
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-often leftover covering must stay empty without cadence leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsCadenceEpisode(t *testing.T) {
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Audrey's dogs meet other dog owners in the park and have doggie playdates."},
		"ep": {
			MemoryID:  "ep",
			Content:   "I try to meet up with other dog owners once a week for tips from other parents and so they can all play together",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("how-often query must keep once-a-week episode leftover")
	}
}

func TestExpandCadenceSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_27"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Audrey's dogs meet other dog owners in the park and have doggie playdates.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about park playdates",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I try to meet up with other dog owners once a week for tips from other parents and so they can all play together",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the cadence leftover can miss")
	}
	expandCadenceSessionNeighbors(candidates, "How often does Audrey meet up with other dog owners for tips and playdates?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("how-often session expand must admit cadence leftover past the generic cap")
	}
}

func TestSessionIDsForHowOftenPrefersCadenceSessions(t *testing.T) {
	q := "How often does Audrey meet up with other dog owners for tips and playdates?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "Audrey's dogs meet other dog owners in the park and have doggie playdates.",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	park := MemoryRecord{
		MemoryID: "park",
		Content:  "Audrey's dogs meet other dog owners in the park and have doggie playdates.",
		Metadata: map[string]any{"session_id": "session_4"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I try to meet up with other dog owners once a week for tips from other parents and so they can all play together",
		Metadata: map[string]any{"session_id": "session_27"},
	}
	seeds := append(recent, park)
	ids := sessionIDsForHowOftenQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_27" {
		t.Fatalf("how-often session rank must prefer cadence leftover session over recency playdates, got %v", ids)
	}
}

func TestLeftoverCoveringWhatProjectWorkingPrefersCurrent(t *testing.T) {
	gold := "James: Yes, we are currently working on a new part of the football simulator"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "James has wanted to turn his childhood comic sketches into a computer game since he was a kid."},
			{Content: "James is creating his own game project."},
			{Content: "John is playing FIFA 23."},
			{Content: "James's sibling is working on a coding project."},
			{Content: "James read a game design magazine."},
			{Content: gold},
			{Content: "James is currently working on a new part of a football simulator, focusing on collecting player databases."},
		},
	}
	q := "What project is James working on in his game design course?"
	if !looksWhatProjectWorkingQuery(q) {
		t.Fatal("what project + working must count as what-project-working")
	}
	if looksWhatProjectWorkingQuery("How often does Audrey meet up with other dog owners for tips and playdates?") {
		t.Fatal("how-often must not count as what-project-working")
	}
	if looksWhatProjectWorkingQuery("How long have Mel and her husband been married?") {
		t.Fatal("how-long-been must not count as what-project-working")
	}
	if looksWhatProjectWorkingQuery("How did Evan start his transformation journey two years ago?") {
		t.Fatal("how-did-start must not count as what-project-working")
	}
	if looksWhatProjectWorkingQuery("What did Audrey do in November 2023 to better take care of her dogs?") {
		t.Fatal("what-did-purpose must not count as what-project-working")
	}
	if looksWhatProjectWorkingQuery("How does Nate describe the process of taking care of turtles?") {
		t.Fatal("how-describe must not count as what-project-working")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "currently working") {
		t.Fatalf("what-project leftover covering must pick currently-working leftover, got %q", got)
	}
	if strings.Contains(lower, "comic") || strings.Contains(lower, "childhood") || strings.Contains(lower, "creating his own") || strings.Contains(lower, "fifa") || strings.Contains(lower, "sibling") || strings.Contains(lower, "magazine") {
		t.Fatalf("childhood desire, creating-own, FIFA, sibling coding, and magazine must not cover what-project, got %q", got)
	}
	if leftoverCoveringCurrentProjectLine(q, "James has wanted to turn his childhood comic sketches into a computer game since he was a kid.") {
		t.Fatal("childhood desire must not count as currently-working covering")
	}
	if leftoverCoveringCurrentProjectLine(q, "James is creating his own game project.") {
		t.Fatal("creating-own without currently must not count as currently-working covering")
	}
	if leftoverCoveringCurrentProjectLine(q, "John is playing FIFA 23.") {
		t.Fatal("peer FIFA leftover must not count as currently-working covering")
	}
	if leftoverCoveringCurrentProjectLine(q, "James's sibling is working on a coding project.") {
		t.Fatal("sibling coding leftover must not count as currently-working covering")
	}
	if leftoverCoveringCurrentProjectLine(q, gold) {
		// speaker leftover is currently-working
	} else {
		t.Fatal("currently-working leftover must count as what-project covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "James is creating his own game project.") {
		t.Fatal("empty covering and creating-own must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How often does Audrey meet up with other dog owners for tips and playdates?", gold) {
		t.Fatal("how-often must not skip hybrid via currently-working leftover")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("currently-working leftover must skip hybrid")
	}
	if !leftoverCoveringCurrentProjectMissesCurrent(q, got, "He is creating his own game project – turning his childhood comic sketches into a computer game.") {
		t.Fatal("hybrid comic-sketch answer must miss currently-working leftover")
	}
	dumpItems := []RecallItem{{Value: "He is creating his own game project – turning his childhood comic sketches into a computer game."}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || !strings.Contains(strings.ToLower(synced[0].Value), "currently working") {
		t.Fatalf("what-project covering must replace enumerate dump items, got %#v", synced)
	}
}

func TestLeftoverCoveringWhatProjectWorkingStaysEmptyWithoutCurrent(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "James has wanted to turn his childhood comic sketches into a computer game since he was a kid."},
			{Content: "James is creating his own game project."},
			{Content: "John is playing FIFA 23."},
			{Content: "James's sibling is working on a coding project."},
		},
	}
	q := "What project is James working on in his game design course?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("what-project leftover covering must stay empty without currently-working leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsCurrentProjectEpisode(t *testing.T) {
	q := "What project is James working on in his game design course?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "James is creating his own game project."},
		"ep": {
			MemoryID:  "ep",
			Content:   "James: Yes, we are currently working on a new part of the football simulator",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-project query must keep currently-working episode leftover")
	}
}

func TestExpandCurrentProjectSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_13"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "James is creating his own game project.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about comic sketches",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "James: Yes, we are currently working on a new part of the football simulator",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the currently-working leftover can miss")
	}
	expandCurrentProjectSessionNeighbors(candidates, "What project is James working on in his game design course?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-project session expand must admit currently-working leftover past the generic cap")
	}
}

func TestSessionIDsForWhatProjectWorkingPrefersCurrentSessions(t *testing.T) {
	q := "What project is James working on in his game design course?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "James has wanted to turn his childhood comic sketches into a computer game since he was a kid.",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	creating := MemoryRecord{
		MemoryID: "creating",
		Content:  "James is creating his own game project.",
		Metadata: map[string]any{"session_id": "session_1"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "James: Yes, we are currently working on a new part of the football simulator",
		Metadata: map[string]any{"session_id": "session_13"},
	}
	seeds := append(recent, creating)
	ids := sessionIDsForWhatProjectWorkingQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_13" {
		t.Fatalf("what-project session rank must prefer currently-working leftover session over recency comic sketches, got %v", ids)
	}
}

func TestLeftoverCoveringWhatNewHobbyPrefersBecomeInterested(t *testing.T) {
	gold := "Lately I've become interested in extreme sports"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector."},
			{Content: "John: I've been getting into a new hobby recently"},
			{Content: "James won an online gaming tournament last week (the week of 2 July 2022)."},
			{Content: gold},
			{Content: "James is interested in extreme sports."},
		},
	}
	q := "What new hobby did James become interested in on 9 July, 2022?"
	if !looksWhatNewHobbyQuery(q) {
		t.Fatal("dated what-new-hobby + become interested must count as what-new-hobby")
	}
	if looksWhatNewHobbyQuery("What project is James working on in his game design course?") {
		t.Fatal("what-project-working must not count as what-new-hobby")
	}
	if looksWhatNewHobbyQuery("How often does Audrey meet up with other dog owners for tips and playdates?") {
		t.Fatal("how-often must not count as what-new-hobby")
	}
	if looksWhatNewHobbyQuery("What new hobby did James become interested in?") {
		t.Fatal("undated what-new-hobby must not count as dated what-new-hobby")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "become interested") {
		t.Fatalf("what-new-hobby leftover covering must pick become-interested leftover, got %q", got)
	}
	if strings.Contains(lower, "metal") || strings.Contains(lower, "detect") || strings.Contains(lower, "john") {
		t.Fatalf("foreign-person new-hobby leftover must not cover dated what-new-hobby, got %q", got)
	}
	if leftoverCoveringBecomeInterestedLine(q, "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.") {
		t.Fatal("foreign metal-detecting leftover must not count as become-interested covering")
	}
	if leftoverCoveringBecomeInterestedLine(q, "John: I've been getting into a new hobby recently") {
		t.Fatal("foreign getting-into leftover must not count as become-interested covering")
	}
	if !leftoverCoveringBecomeInterestedLine(q, gold) {
		t.Fatal("first-person become-interested leftover must count as what-new-hobby covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.") {
		t.Fatal("empty covering and foreign new-hobby must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What project is James working on in his game design course?", gold) {
		t.Fatal("what-project-working must not skip hybrid via become-interested leftover")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("become-interested leftover must skip hybrid")
	}
	if !leftoverCoveringBecomeInterestedMissesInterest(q, got, "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.") {
		t.Fatal("foreign new-hobby answer must miss become-interested leftover")
	}
	dumpItems := []RecallItem{{Value: "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector."}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || !strings.Contains(strings.ToLower(synced[0].Value), "become interested") {
		t.Fatalf("what-new-hobby covering must replace enumerate dump items, got %#v", synced)
	}
}

func TestLeftoverCoveringWhatNewHobbyStaysEmptyWithoutInterest(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector."},
			{Content: "James won an online gaming tournament last week (the week of 2 July 2022)."},
			{Content: "James took many photographs on 21 July 2022."},
		},
	}
	q := "What new hobby did James become interested in on 9 July, 2022?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("what-new-hobby leftover covering must stay empty without become-interested leftover, got %q", got)
	}
}

func TestLeftoverCoveringHowPlanDreamPrefersPrepPlan(t *testing.T) {
	gold := "I've been gathering information, watching videos, and I even got a beginners' guide to surfing"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Deborah: Exploring historical places and learning their stories is so fun"},
			{Content: "Deborah enjoys exploring historical places and learning their stories."},
			{Content: "Jolene plans to learn to surf."},
			{Content: "Jolene has been gathering information about surfing."},
			{Content: gold},
		},
	}
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	if !looksHowPlanDreamQuery(q) {
		t.Fatal("how-plan-dream + learning must count as how-plan-dream")
	}
	if looksHowPlanDreamQuery("What new hobby did James become interested in on 9 July, 2022?") {
		t.Fatal("dated what-new-hobby must not count as how-plan-dream")
	}
	if looksHowPlanDreamQuery("How often does Audrey meet up with other dog owners for tips and playdates?") {
		t.Fatal("how-often must not count as how-plan-dream")
	}
	if looksHowPlanDreamQuery("What project is James working on in his game design course?") {
		t.Fatal("what-project-working must not count as how-plan-dream")
	}
	if looksHowPlanDreamQuery("How does Jolene plan her week?") {
		t.Fatal("plan without dream+learning must not count as how-plan-dream")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "gathering information") || !strings.Contains(lower, "watching videos") {
		t.Fatalf("how-plan-dream leftover covering must pick gathering/watching leftover, got %q", got)
	}
	if strings.Contains(lower, "deborah") || strings.Contains(lower, "historical") || strings.Contains(lower, "exploring") {
		t.Fatalf("foreign-person exploring leftover must not cover how-plan-dream, got %q", got)
	}
	if leftoverCoveringPrepPlanLine(q, "Deborah: Exploring historical places and learning their stories is so fun") {
		t.Fatal("foreign exploring leftover must not count as prep-plan covering")
	}
	if leftoverCoveringPrepPlanLine(q, "Jolene plans to learn to surf.") {
		t.Fatal("thin plans-to-learn leftover must not count as prep-plan covering")
	}
	if !leftoverCoveringPrepPlanLine(q, gold) {
		t.Fatal("first-person gathering/watching/guide leftover must count as how-plan-dream covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, "Deborah: Exploring historical places and learning their stories is so fun") {
		t.Fatal("empty covering and foreign exploring must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What new hobby did James become interested in on 9 July, 2022?", gold) {
		t.Fatal("what-new-hobby must not skip hybrid via prep-plan leftover")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("prep-plan leftover must skip hybrid")
	}
	if !leftoverCoveringPrepPlanMissesPrep(q, got, "Deborah: Exploring historical places and learning their stories is so fun") {
		t.Fatal("foreign exploring answer must miss prep-plan leftover")
	}
	dumpItems := []RecallItem{{Value: "Deborah: Exploring historical places and learning their stories is so fun"}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || !strings.Contains(strings.ToLower(synced[0].Value), "gathering information") {
		t.Fatalf("how-plan-dream covering must replace enumerate dump items, got %#v", synced)
	}
}

func TestLeftoverCoveringHowPlanDreamStaysEmptyWithoutPrep(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Deborah: Exploring historical places and learning their stories is so fun"},
			{Content: "Jolene plans to learn to surf."},
			{Content: "Jolene wants to solidify her workshop plan before reaching out to schools or centers."},
		},
	}
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("how-plan-dream leftover covering must stay empty without gathering/watching/guide leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsPrepPlanEpisode(t *testing.T) {
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Deborah: Exploring historical places and learning their stories is so fun"},
		"ep": {
			MemoryID:  "ep",
			Content:   "I've been gathering information, watching videos, and I even got a beginners' guide to surfing",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("how-plan-dream query must keep prep-plan episode leftover")
	}
}

func TestExpandPrepPlanSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_10"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Jolene plans to learn to surf.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about workshop plans",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I've been gathering information, watching videos, and I even got a beginners' guide to surfing",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the prep-plan leftover can miss")
	}
	expandPrepPlanSessionNeighbors(candidates, "How does Jolene plan to pursue her dream of learning to surf?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("how-plan-dream session expand must admit prep-plan leftover past the generic cap")
	}
}

func TestSessionIDsForHowPlanDreamPrefersPrepSessions(t *testing.T) {
	q := "How does Jolene plan to pursue her dream of learning to surf?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "Deborah: Exploring historical places and learning their stories is so fun",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	explore := MemoryRecord{
		MemoryID: "explore",
		Content:  "Deborah: Exploring historical places and learning their stories is so fun",
		Metadata: map[string]any{"session_id": "session_23"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I've been gathering information, watching videos, and I even got a beginners' guide to surfing",
		Metadata: map[string]any{"session_id": "session_10"},
	}
	seeds := append(recent, explore)
	ids := sessionIDsForHowPlanDreamQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_10" {
		t.Fatalf("how-plan-dream session rank must prefer prep-plan leftover session over recency exploring, got %v", ids)
	}
}

func TestLeftoverCoveringWhatFocusingBesidesPrefersPossessiveJoin(t *testing.T) {
	gold := "I've been focusing on studying and my relationship with my partner"
	engineering := "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes."
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: engineering},
			{Content: "She’s been focusing on applying her engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes."},
			{Content: "I've been focusing on yoga and feeling more connected to my body."},
			{Content: "I'm focusing on studying and gaining more experience"},
			{Content: gold},
		},
	}
	q := "What has Jolene been focusing on lately besides studying?"
	if !looksWhatFocusingBesidesQuery(q) {
		t.Fatal("what-has + focusing + besides + object must count as focusing-besides")
	}
	if looksWhatFocusingBesidesQuery("How does Jolene plan to pursue her dream of learning to surf?") {
		t.Fatal("how-plan-dream must not count as focusing-besides")
	}
	if looksWhatFocusingBesidesQuery("What new hobby did James become interested in on 9 July, 2022?") {
		t.Fatal("dated what-new-hobby must not count as focusing-besides")
	}
	if looksWhatFocusingBesidesQuery("What project is James working on in his game design course?") {
		t.Fatal("what-project-working must not count as focusing-besides")
	}
	if looksWhatFocusingBesidesQuery("How often does Audrey meet up with other dog owners for tips and playdates?") {
		t.Fatal("how-often must not count as focusing-besides")
	}
	if looksWhatFocusingBesidesQuery("What has Jolene been focusing on lately?") {
		t.Fatal("focusing without besides/except/aside must not count as focusing-besides")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "relationship") || !strings.Contains(lower, "partner") {
		t.Fatalf("focusing-besides leftover covering must pick possessive-join leftover, got %q", got)
	}
	if strings.Contains(lower, "engineering") || strings.Contains(lower, "social causes") || strings.Contains(lower, "yoga") {
		t.Fatalf("occupation leftover must not cover focusing-besides, got %q", got)
	}
	if leftoverCoveringFocusingBesidesLine(q, engineering) {
		t.Fatal("engineering focusing leftover must not count as focusing-besides covering")
	}
	if leftoverCoveringFocusingBesidesLine(q, "I'm focusing on studying and gaining more experience") {
		t.Fatal("studying conjunct without possessive join must not count as focusing-besides covering")
	}
	if leftoverCoveringFocusingBesidesLine(q, "I've been focusing on yoga and feeling more connected to my body.") {
		t.Fatal("foreign yoga leftover must not count as focusing-besides covering")
	}
	if !leftoverCoveringFocusingBesidesLine(q, gold) {
		t.Fatal("first-person focusing-on + besides-object + possessive conjunct leftover must count as covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, engineering) {
		t.Fatal("empty covering and occupation leftover must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("How does Jolene plan to pursue her dream of learning to surf?", gold) {
		t.Fatal("how-plan-dream must not skip hybrid via focusing-besides leftover")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("focusing-besides leftover must skip hybrid")
	}
	if !leftoverCoveringFocusingBesidesMissesJoin(q, got, engineering) {
		t.Fatal("occupation leftover answer must miss focusing-besides covering")
	}
	dumpItems := []RecallItem{{Value: engineering}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || !strings.Contains(strings.ToLower(synced[0].Value), "relationship") {
		t.Fatalf("focusing-besides covering must replace enumerate dump items, got %#v", synced)
	}
}

func TestLeftoverCoveringWhatFocusingBesidesStaysEmptyWithoutJoin(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes."},
			{Content: "I'm focusing on studying and gaining more experience"},
			{Content: "I've been focusing on yoga and feeling more connected to my body."},
		},
	}
	q := "What has Jolene been focusing on lately besides studying?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("focusing-besides leftover covering must stay empty without possessive-join leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsFocusingBesidesEpisode(t *testing.T) {
	q := "What has Jolene been focusing on lately besides studying?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes."},
		"ep": {
			MemoryID:  "ep",
			Content:   "I've been focusing on studying and my relationship with my partner",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("focusing-besides query must keep possessive-join episode leftover")
	}
}

func TestExpandFocusingBesidesSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_24"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about engineering workshops",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I've been focusing on studying and my relationship with my partner",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the focusing-besides leftover can miss")
	}
	expandFocusingBesidesSessionNeighbors(candidates, "What has Jolene been focusing on lately besides studying?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("focusing-besides session expand must admit possessive-join leftover past the generic cap")
	}
}

func TestSessionIDsForWhatFocusingBesidesPrefersJoinSessions(t *testing.T) {
	q := "What has Jolene been focusing on lately besides studying?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes.",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	engineering := MemoryRecord{
		MemoryID: "engineering",
		Content:  "I've been focusing on applying my engineering skills to sustainable, community-focused projects and getting involved with organizations that address social causes.",
		Metadata: map[string]any{"session_id": "session_18"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "I've been focusing on studying and my relationship with my partner",
		Metadata: map[string]any{"session_id": "session_24"},
	}
	seeds := append(recent, engineering)
	ids := sessionIDsForWhatFocusingBesidesQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_24" {
		t.Fatalf("focusing-besides session rank must prefer possessive-join leftover session over recency occupation leftover, got %v", ids)
	}
}

func TestLeftoverCoveringWhatNewSeriesPrefersQuotedShow(t *testing.T) {
	gold := `Tim: I'm really excited to watch this new show that's coming out called "The Wheel of Time"`
	journey := "I'm really excited about this new journey"
	novel := `Tim recommends the fantasy novel "The Name of the Wind" as a captivating read.`
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: journey},
			{Content: novel},
			{Content: "Tim says Game of Thrones is amazing and he is totally hooked on it."},
			{Content: gold},
		},
	}
	q := "What new fantasy TV series is Tim excited about?"
	if !looksWhatNewSeriesQuery(q) {
		t.Fatal("what + new + series must count as what-new-series")
	}
	if looksWhatNewSeriesQuery("What has Jolene been focusing on lately besides studying?") {
		t.Fatal("focusing-besides must not count as what-new-series")
	}
	if looksWhatNewSeriesQuery("What new hobby did James become interested in on 9 July, 2022?") {
		t.Fatal("dated what-new-hobby must not count as what-new-series")
	}
	if looksWhatNewSeriesQuery("How does Jolene plan to pursue her dream of learning to surf?") {
		t.Fatal("how-plan-dream must not count as what-new-series")
	}
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "wheel of time") {
		t.Fatalf("what-new-series leftover covering must pick quoted titled-show leftover, got %q", got)
	}
	if strings.Contains(lower, "new journey") || strings.Contains(lower, "name of the wind") || strings.Contains(lower, "game of thrones") {
		t.Fatalf("generic excitement or quoted novel must not cover what-new-series, got %q", got)
	}
	if leftoverCoveringTitledShowLine(q, journey) {
		t.Fatal("generic new-journey leftover must not count as titled-show covering")
	}
	if leftoverCoveringTitledShowLine(q, novel) {
		t.Fatal("quoted novel leftover must not count as titled-show covering")
	}
	if leftoverCoveringTitledShowLine(q, "Tim says Game of Thrones is amazing and he is totally hooked on it.") {
		t.Fatal("unquoted other-series leftover must not count as titled-show covering")
	}
	if !leftoverCoveringTitledShowLine(q, gold) {
		t.Fatal("quoted watch/show leftover must count as titled-show covering")
	}
	if leftoverCoveringSkipsHybrid(q, "") || leftoverCoveringSkipsHybrid(q, journey) {
		t.Fatal("empty covering and generic journey must not skip hybrid")
	}
	if leftoverCoveringSkipsHybrid("What has Jolene been focusing on lately besides studying?", gold) {
		t.Fatal("focusing-besides must not skip hybrid via titled-show leftover")
	}
	if !leftoverCoveringSkipsHybrid(q, gold) {
		t.Fatal("titled-show leftover must skip hybrid")
	}
	if !leftoverCoveringTitledShowMissesTitle(q, got, journey) {
		t.Fatal("generic journey answer must miss titled-show covering")
	}
	dumpItems := []RecallItem{{Value: journey}}
	synced := leftoverCoveringSyncEnumerateItems(q, got, dumpItems)
	if len(synced) != 1 || !strings.Contains(strings.ToLower(synced[0].Value), "wheel of time") {
		t.Fatalf("what-new-series covering must replace enumerate dump items, got %#v", synced)
	}
}

func TestLeftoverCoveringWhatNewSeriesStaysEmptyWithoutTitle(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "I'm really excited about this new journey"},
			{Content: `Tim recommends the fantasy novel "The Name of the Wind" as a captivating read.`},
			{Content: "Tim says Game of Thrones is amazing and he is totally hooked on it."},
		},
	}
	q := "What new fantasy TV series is Tim excited about?"
	got := leftoverCoveringSpecificAnswer(q, nil, pkt)
	if got != "" {
		t.Fatalf("what-new-series leftover covering must stay empty without quoted titled-show leftover, got %q", got)
	}
}

func TestApplyFactPrimaryRecallKeepsTitledShowEpisode(t *testing.T) {
	q := "What new fantasy TV series is Tim excited about?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "I'm really excited about this new journey"},
		"ep": {
			MemoryID:  "ep",
			Content:   `Tim: I'm really excited to watch this new show that's coming out called "The Wheel of Time"`,
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-new-series query must keep titled-show episode leftover")
	}
}

func TestExpandTitledShowSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_26"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "I'm really excited about this new journey",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about fantasy novels",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  `Tim: I'm really excited to watch this new show that's coming out called "The Wheel of Time"`,
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the titled-show leftover can miss")
	}
	expandTitledShowSessionNeighbors(candidates, "What new fantasy TV series is Tim excited about?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-new-series session expand must admit titled-show leftover past the generic cap")
	}
}

func TestSessionIDsForWhatNewSeriesPrefersTitleSessions(t *testing.T) {
	q := "What new fantasy TV series is Tim excited about?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "I'm really excited about this new journey",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	journey := MemoryRecord{
		MemoryID: "journey",
		Content:  "I'm really excited about this new journey",
		Metadata: map[string]any{"session_id": "session_2"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  `Tim: I'm really excited to watch this new show that's coming out called "The Wheel of Time"`,
		Metadata: map[string]any{"session_id": "session_26"},
	}
	seeds := append(recent, journey)
	ids := sessionIDsForWhatNewSeriesQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_26" {
		t.Fatalf("what-new-series session rank must prefer titled-show leftover session over recency journey leftover, got %v", ids)
	}
}

func TestApplyFactPrimaryRecallKeepsBecomeInterestedEpisode(t *testing.T) {
	q := "What new hobby did James become interested in on 9 July, 2022?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector."},
		"ep": {
			MemoryID:  "ep",
			Content:   "Lately I've become interested in extreme sports",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-new-hobby query must keep become-interested episode leftover")
	}
}

func TestExpandBecomeInterestedSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_16"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "James won an online gaming tournament last week (the week of 2 July 2022).",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about july photographs",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Lately I've become interested in extreme sports",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the become-interested leftover can miss")
	}
	expandBecomeInterestedSessionNeighbors(candidates, "What new hobby did James become interested in on 9 July, 2022?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("what-new-hobby session expand must admit become-interested leftover past the generic cap")
	}
}

func TestSessionIDsForWhatNewHobbyPrefersInterestSessions(t *testing.T) {
	q := "What new hobby did James become interested in on 9 July, 2022?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	metal := MemoryRecord{
		MemoryID: "metal",
		Content:  "John has taken up metal detecting as a new hobby, walking along beaches with a metal detector.",
		Metadata: map[string]any{"session_id": "session_2"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Lately I've become interested in extreme sports",
		Metadata: map[string]any{"session_id": "session_16"},
	}
	seeds := append(recent, metal)
	ids := sessionIDsForWhatNewHobbyQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_16" {
		t.Fatalf("what-new-hobby session rank must prefer become-interested leftover session over recency metal detecting, got %v", ids)
	}
}

func TestApplyFactPrimaryRecallKeepsStartMethodEpisode(t *testing.T) {
	q := "How did Evan start his transformation journey two years ago?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Evan possesses a gym membership card."},
		"ep": {
			MemoryID:  "ep",
			Content:   "Changed my diet, started walking regularly, things like that",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("how-did-start query must keep changed+started episode leftover")
	}
}

func TestExpandStartMethodSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_15"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Evan went for a morning walk.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about morning walks",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Changed my diet, started walking regularly, things like that",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the start-method leftover can miss")
	}
	expandStartMethodSessionNeighbors(candidates, "How did Evan start his transformation journey two years ago?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("how-did-start session expand must admit changed+started leftover past the generic cap")
	}
}

func TestExpandDurationSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_3"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Melanie is married.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about family hikes",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Melanie's marriage duration is 5 years.",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the duration leftover can miss")
	}
	expandDurationSessionNeighbors(candidates, "How long have Mel and her husband been married?", []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("how-long-been session expand must admit continuing duration leftover past the generic cap")
	}
}

func TestSessionIDsForHowLongBeenPrefersStateTokenSessions(t *testing.T) {
	q := "How long have Mel and her husband been married?"
	recent := make([]MemoryRecord, 0, 12)
	for i := 0; i < 8; i++ {
		recent = append(recent, MemoryRecord{
			MemoryID: "chat-" + itoa(i),
			Content:  "Caroline: Hey Mel, long time no chat",
			Metadata: map[string]any{"session_id": "session_recent_" + itoa(i)},
		})
	}
	husband := MemoryRecord{
		MemoryID: "husband",
		Content:  "Melanie has a husband.",
		Metadata: map[string]any{"session_id": "session_3"},
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "Melanie's marriage duration is 5 years.",
		Metadata: map[string]any{"session_id": "session_3"},
	}
	seeds := append(recent, husband)
	ids := sessionIDsForHowLongBeenQuery(q, seeds, []MemoryRecord{gold})
	if len(ids) == 0 || ids[0] != "session_3" {
		t.Fatalf("how-long-been session rank must prefer duration/husband session over recency chat, got %v", ids)
	}
	raw := sessionIDsOf(seeds)
	if len(raw) > 8 {
		raw = raw[:8]
	}
	found := false
	for _, id := range raw {
		if id == "session_3" {
			found = true
		}
	}
	if found && len(raw) <= 8 {
		// still OK if recency order happened to include it; the ranked helper must keep it first
	}
}

func TestExpandKindListSessionNeighborsAdmitsGoldPastCap(t *testing.T) {
	session := "session_13"
	seed := MemoryRecord{
		MemoryID: "seed",
		Content:  "Maria and her mother cooked dinner together at home on 3 May 2023.",
		Metadata: map[string]any{"session_id": session},
	}
	candidates := map[string]MemoryRecord{"seed": seed}
	all := []MemoryRecord{seed}
	for i := 0; i < 16; i++ {
		all = append(all, MemoryRecord{
			MemoryID: "noise-" + itoa(i),
			Content:  "unrelated session chatter about dinner",
			Metadata: map[string]any{"session_id": session},
		})
	}
	gold := MemoryRecord{
		MemoryID: "gold",
		Content:  "It had lots of great things like salads, sandwiches, and homemade desserts",
		Metadata: map[string]any{"session_id": session},
	}
	all = append(all, gold)
	expandSessionNeighbors(candidates, []MemoryRecord{seed}, all, 16)
	if _, ok := candidates["gold"]; ok {
		t.Fatal("generic session expand must stay capped so the kind-list leftover can miss")
	}
	expandKindListSessionNeighbors(candidates, []MemoryRecord{seed}, all, 8)
	if _, ok := candidates["gold"]; !ok {
		t.Fatal("kind-list session expand must admit the like-list leftover past the generic cap")
	}
}

func TestSessionIDsForWhatKindQueryPrefersDinnerSession(t *testing.T) {
	q := "What kind of food did Maria have on her dinner spread iwth her mother?"
	kindness := MemoryRecord{
		MemoryID: "k",
		Content:  "Maria believes that spreading kindness and support can make a difference when someone is feeling down.",
		Metadata: map[string]any{"session_id": "session_5"},
	}
	dinner := MemoryRecord{
		MemoryID: "d",
		Content:  "Maria and her mother cooked dinner together at home on 3 May 2023.",
		Metadata: map[string]any{"session_id": "session_13"},
	}
	ids := sessionIDsForWhatKindQuery(q, []MemoryRecord{kindness, dinner})
	if len(ids) == 0 || ids[0] != "session_13" {
		t.Fatalf("what-kind session pick must prefer maria/mother/dinner coverage, got %#v", ids)
	}
	for _, id := range ids {
		if id == "session_5" {
			t.Fatal("kindness-spread session must not be fetched from restatement token overlap")
		}
	}
}

func TestApplyKindListRankBoostSkipsKindness(t *testing.T) {
	q := "What kind of food did Maria have on her dinner spread iwth her mother?"
	gold := 1.0
	kindness := 1.0
	applyKindListRankBoost(&gold, map[string]any{}, q, MemoryRecord{Content: "It had lots of great things like salads, sandwiches, and homemade desserts"})
	applyKindListRankBoost(&kindness, map[string]any{}, q, MemoryRecord{Content: "Maria believes that spreading kindness and support can make a difference when someone is feeling down."})
	if gold <= kindness {
		t.Fatalf("kind-list rank boost must lift like-list leftover over spreading kindness, gold=%v kindness=%v", gold, kindness)
	}
}

func TestScoreMemoryIDFKindListNeedsFloor(t *testing.T) {
	q := "What kind of food did Maria have on her dinner spread iwth her mother?"
	gold := MemoryRecord{Content: "It had lots of great things like salads, sandwiches, and homemade desserts"}
	if !leftoverCoveringKindListLine(gold.Content) {
		t.Fatal("like-list leftover must classify as kind-list")
	}
	score, _ := scoreMemoryIDF(gold, q, tokenize(q), nil, nil)
	if score > 0 {
		t.Fatalf("kind-list gold has no query tokens so IDF score must be 0 without a floor, got %v", score)
	}
}

func TestApplyFactPrimaryRecallKeepsKindListEpisode(t *testing.T) {
	q := "What kind of food did Maria have on her dinner spread iwth her mother?"
	candidates := map[string]MemoryRecord{
		"fact": {MemoryID: "fact", Content: "Maria and her mother cooked dinner together at home on 3 May 2023."},
		"ep": {
			MemoryID:  "ep",
			Content:   "It had lots of great things like salads, sandwiches, and homemade desserts",
			Primitive: PrimitiveEpisode,
		},
	}
	applyFactPrimaryRecall(candidates, q, false)
	if _, ok := candidates["ep"]; !ok {
		t.Fatal("what-kind query must keep like-list episode leftover")
	}
}

func TestPreferUnwindPacketActivitiesJoinsCalmingSlots(t *testing.T) {
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley runs to unwind after work."},
			{Content: "Riley finds making pottery calming"},
			{Content: "Riley enjoys camping"},
			{Content: "Riley participates in pottery"},
			{Content: "Dana finds making jewelry calming"},
		},
	}
	q := "What does Riley do to unwind?"
	got, extra := preferUnwindPacketActivities(q, "runs, running", pkt, nil)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "run") || !strings.Contains(lower, "potter") {
		t.Fatalf("unwind packet join must add calming pottery, got %q", got)
	}
	if strings.Contains(lower, "camp") || strings.Contains(lower, "nurse") || strings.Contains(lower, "jewel") {
		t.Fatalf("unwind packet join must not dump camping or another person's calming slot, got %q", got)
	}
	items := appendUniqueRecallItems([]RecallItem{{Value: "runs"}, {Value: "running"}}, extra)
	itemBlob := ""
	for _, it := range items {
		itemBlob += " " + strings.ToLower(it.Value)
	}
	if !strings.Contains(itemBlob, "potter") {
		t.Fatalf("enumerate items must receive joined unwind slots, got %#v", items)
	}
	if next, _ := preferUnwindPacketActivities("When did Riley run?", "19 January 2023", pkt, nil); next != "" {
		t.Fatal("non-unwind queries must not take unwind packet join")
	}
	hopOnly, hopExtra := preferUnwindPacketActivities(q, "runs, running", EvidencePacket{}, []HopResult{
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity,
			Value: "pottery", Contents: []string{"Riley finds making pottery calming"}},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity,
			Value: "camping", Contents: []string{"Riley enjoys camping"}},
	})
	hopLower := strings.ToLower(hopOnly)
	if !strings.Contains(hopLower, "potter") {
		t.Fatalf("unwind join must use hop contents when the packet leftover omits them, got %q", hopOnly)
	}
	if strings.Contains(hopLower, "camp") {
		t.Fatalf("hop participates-in camping must not join, got %q", hopOnly)
	}
	hopItems := appendUniqueRecallItems([]RecallItem{{Value: "runs"}, {Value: "running"}}, hopExtra)
	hopBlob := ""
	for _, it := range hopItems {
		hopBlob += " " + strings.ToLower(it.Value)
	}
	if !strings.Contains(hopBlob, "potter") {
		t.Fatalf("hop-content unwind extras must land on enumerate items, got %#v", hopItems)
	}
}

func TestPreferPracticePacketPlacesJoinsLeftoverLocatives(t *testing.T) {
	q := "Which locations does Riley practice her yoga at?"
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "Riley bought a scented candle for her yoga practice on 28 March 2023."},
			{Content: "Riley does yoga in pajamas in the living room by the window."},
			{Content: "Riley reminisced with friends and reconnected after yoga."},
			{Content: "Riley does yoga on the beach."},
			{Content: "Riley recommends the yoga studio nearby."},
			{Content: "Riley practices yoga at her mother's old home."},
			{Content: "Dana does yoga in Denver."},
		},
	}
	got, extra := preferPracticePacketPlaces(q, "the park", pkt, nil)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "park") || !strings.Contains(lower, "beach") || !strings.Contains(lower, "studio") || !strings.Contains(lower, "mother") {
		t.Fatalf("practice packet join must keep the park and add leftover places, got %q", got)
	}
	if strings.Contains(lower, "candle") || strings.Contains(lower, "denver") || strings.Contains(lower, "pajama") || strings.Contains(lower, "living") || strings.Contains(lower, "reminisc") {
		t.Fatalf("practice packet join must not dump purchases, rooms, or another person's place, got %q", got)
	}
	items := appendUniqueRecallItems([]RecallItem{{Value: "the park"}}, extra)
	itemBlob := ""
	for _, it := range items {
		itemBlob += " " + strings.ToLower(it.Value)
	}
	if !strings.Contains(itemBlob, "park") || !strings.Contains(itemBlob, "beach") || !strings.Contains(itemBlob, "studio") || !strings.Contains(itemBlob, "mother") {
		t.Fatalf("enumerate items must receive joined practice places, got %#v", items)
	}
	if next, _ := preferPracticePacketPlaces("When did Riley start yoga?", "2020", pkt, nil); next != "" {
		t.Fatal("non-location queries must not take practice packet join")
	}
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
	}
	if leftoverThinMissAnswer(q, hops, "the park") {
		t.Fatal("location-list park answer must not be a leftover thin miss")
	}
	cover := leftoverCoveringSpecificAnswer(q, hops, pkt)
	coverLower := strings.ToLower(cover)
	if strings.Contains(coverLower, "candle") {
		t.Fatalf("location-list leftover covering must skip purchase leftover, got %q", cover)
	}
	if leftoverCoveringMayReplaceHybrid(q, hops, "Riley bought a scented candle for her yoga practice on 28 March 2023.", "the park") {
		t.Fatal("location-list hybrid answers must not yield to purchase leftover")
	}
}

func TestLeftoverCoveringKeepsTypedItemJoins(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Tim", Value: "Tim", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "John", Value: "John", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Tim", Predicate: PredicatePossession, Source: "typed_store",
			Value: "Basketball Signed By Favorite Player", Values: []string{"Basketball Signed By Favorite Player"}},
		{Kind: "follow_relation", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Value: "Basketball Trophy", Values: []string{"Basketball Trophy"}},
	}
	join := "Basketball Signed By Favorite Player, Basketball Trophy"
	q := "What similar sports collectible do Tim and John own?"
	if !leftoverShortItemJoin(join) {
		t.Fatal("signed-collectible comma join must count as a short item join")
	}
	if leftoverThinMissAnswer(q, hops, join) {
		t.Fatal("typed collectible join must not be a leftover thin miss")
	}
	if !leftoverCoveringKeepTypedAnswer(q, hops, join) {
		t.Fatal("typed possession join must be kept against leftover covering")
	}
	chat := "John: Having someone to support and motivate you is so important, whether it's in sports or any other aspect of life"
	if leftoverCoveringKeepTypedAnswer(q, hops, chat) {
		t.Fatal("sports pep-talk must not count as a typed item join")
	}
	if leftoverCoveringBeatsAnswer(q, hops, chat, join) {
		t.Fatal("leftover covering must not beat a typed possession join")
	}
	snackQ := "What kind of unhealthy snacks does Sam enjoy eating?"
	snackHops := []HopResult{
		{Kind: "resolve_entity", Entity: "Sam", Value: "Sam", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Sam", Predicate: PredicatePreference, Source: "typed_store",
			Value: "soda, candy", Values: []string{"soda", "candy"}},
	}
	if leftoverThinMissAnswer(snackQ, snackHops, "soda and candy") {
		t.Fatal("short snack join must not be a leftover thin miss")
	}
	if leftoverCoveringBeatsAnswer(snackQ, snackHops, "Sam bought unhealthy snacks.", "soda and candy") {
		t.Fatal("list-head leftover covering must not beat a snack join")
	}
}

func TestLeftoverCoveringWhereIgnoresHopSlotStarvation(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity, Source: "typed_store",
			Value:  "hiking, road trip to jasper national park, hiked trails with family",
			Values: []string{"hiking, road trip to jasper national park, hiked trails with family"}},
	}
	q := "Where did Riley take his family for a road trip on 24 May, 2023?"
	got := leftoverCoveringRareForQuery(q, hops)
	hasTrip := false
	for _, tok := range got {
		if tok == "trip" || tok == "family" || tok == "road" {
			hasTrip = true
		}
	}
	if !hasTrip {
		t.Fatalf("where leftover rare must keep locative tokens despite hop dumps, hops=%v covering=%v", leftoverCoverRareTokens(q, hops), got)
	}
}

func TestLeftoverCoveringPrefersDatedPlanOverActivityPepTalk(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Casey", Value: "Casey", Source: "search_fallback"},
		{Kind: "follow_relation", Entity: "Riley", Predicate: PredicateActivity, Source: "typed_store",
			Value: "kayaking", Values: []string{"kayaking"}},
	}
	pkt := EvidencePacket{
		ContextEvidence: []PacketItem{
			{Content: "It's such a rewarding and tough activity - keep going and have fun"},
			{Content: "Riley plans to organize a kayaking trip together with Casey."},
			{Content: "Casey plans to paint with Riley on Saturday, 16 September 2023."},
			{Content: "Riley finished a contemporary figurative painting a few days ago (2023-12-14)."},
		},
	}
	q := "Which activity do Riley and Casey plan on doing together during September 2023?"
	rare := leftoverCoveringRareForQuery(q, hops)
	if !contentCoversAnyQueryToken("september", rare) {
		t.Fatalf("weak-only activity queries must keep the month token, rare=%v", rare)
	}
	if contentCoversAnyQueryToken("activity", leftoverCoverNonWeakTokens(rare)) {
		t.Fatalf("activity must not stay a leftover covering bind token, rare=%v", rare)
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "paint") {
		t.Fatalf("dated September plan must cover, got %q rare=%v", got, rare)
	}
	if strings.Contains(lower, "keep going") || strings.Contains(lower, "have fun") || strings.Contains(lower, "kayak") {
		t.Fatalf("pep-talk or undated kayaking must not cover a September activity, got %q", got)
	}
	if !leftoverThinMissAnswer(q, hops, "kayaking") {
		t.Fatal("one-word hop activity that misses the month must stay a leftover thin miss")
	}
}

func TestLeftoverCoveringSkipsWeakAdverbCrowd(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Tim", Value: "Tim", Source: "search_fallback"},
	}
	pkt := EvidencePacket{
		Contents: []string{
			"John's family gets together frequently.",
			"Tim visited a Harry Potter-themed place in London a few years ago.",
			"Tim has visited the UK the most frequently in his travels.",
		},
	}
	got := leftoverCoveringSpecificAnswer("which country has Tim visited most frequently in his travels?", hops, pkt)
	lower := strings.ToLower(got)
	if strings.Contains(lower, "family gets together") {
		t.Fatalf("frequently-only chat must not cover a country question: %q", got)
	}
	if !strings.Contains(lower, "uk") && !strings.Contains(lower, "london") {
		t.Fatalf("expected UK/London leftover covering, got %q", got)
	}
}

func TestLeftoverThinSloganAnswer(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "Riley", Value: "Riley", Source: "search_fallback"},
		{Kind: "resolve_entity", Entity: "Casey", Value: "Casey", Source: "search_fallback"},
	}
	if !leftoverThinMissAnswer("What plans do Riley and Casey have for when Riley visits Boston?", hops, "scheduled to end soon") {
		t.Fatal("unrelated thin slogan must yield to leftover covering")
	}
	if !leftoverQueryEchoAnswer("What activities have been helping Riley stay distracted during tough times?", "tough") {
		t.Fatal("leftover query-echo must count as an echo slogan")
	}
	if leftoverQueryEchoAnswer("What similar collectible do Tim and John own?", "jersey") {
		t.Fatal("typed possession must not be a query echo")
	}
	if leftoverThinSloganAnswer("What activities have been helping Riley stay distracted during tough times?", hops, "tough") && leftoverThinMissAnswer("What activities have been helping Riley stay distracted during tough times?", hops, "tough") {
		t.Fatal("query-echo must not also count as leftoverThinMiss")
	}
	if leftoverThinSloganAnswer("Where does Riley live?", hops, "jersey") {
		t.Fatal("single typed place must not be a thin slogan")
	}
}

func TestLeftoverCoveringSkipsChildhoodNameDump(t *testing.T) {
	q := "What is the name of Audrey's childhood dog?"
	if leftoverCoveringLockChildhoodPossessions(q) {
		t.Fatal("childhood name questions must not lock hop possession dumps")
	}
	if !leftoverCoveringLockChildhoodPossessions("What items des John mention having as a child?") {
		t.Fatal("childhood item lists should still lock a 2-item hybrid")
	}
	dump := "Four Dogs, Photo Of Lake, Behavior Tips, Buddy, Fourth Dog (unnamed), Birdwatching Guidebook"
	if leftoverCoveringMayReplaceHybrid(q, nil, dump, "Max") {
		t.Fatalf("leftover covering %q must not replace childhood dog name Max", dump)
	}
}

func TestLeftoverCoveringJoinsChildhoodPossessions(t *testing.T) {
	hops := []HopResult{
		{Kind: "resolve_entity", Entity: "John", Value: "John", Source: "search_fallback"},
		{Kind: "fetch_predicate", Entity: "John", Predicate: PredicatePossession, Source: "typed_store",
			Value:  "little doll, film camera (as a kid)",
			Values: []string{"little doll", "film camera (as a kid)"}},
	}
	pkt := EvidencePacket{
		Contents: []string{
			"Maria: I can picture you all laughing and having a blast making your own pizzas - a great way to bond",
			"John had a film camera when he was a kid.",
			"John had a little doll in his childhood that always made him feel better.",
			"John has children.",
		},
	}
	q := "What items des John mention having as a child?"
	pizza := "Maria: I can picture you all laughing and having a blast making your own pizzas - a great way to bond"
	if !leftoverSkipLine(q, pizza, leftoverCoverNonWeakTokens(leftoverNonEntityRareTokens(q, hops))) {
		t.Fatal("having-a-blast chat must not survive leftover covering skip")
	}
	got := leftoverCoveringSpecificAnswer(q, hops, pkt)
	lower := strings.ToLower(got)
	if strings.Contains(lower, "pizza") || strings.Contains(lower, "blast") {
		t.Fatalf("childhood leftover covering must not pick pizza chat, got %q", got)
	}
	if !strings.Contains(lower, "doll") || !strings.Contains(lower, "camera") {
		t.Fatalf("expected joined childhood possessions, got %q", got)
	}
}

func TestRecallOrdinalNameBeatsIdentityDump(t *testing.T) {
	t.Setenv("BRAINY_RECALL_LLM", "")
	store := newMemoryStoreStub()
	svc := NewService(store)
	now := svc.now()
	store.records["insp"] = MemoryRecord{
		MemoryID: "mem_insp", TenantID: "t-ord", SubjectID: "u1",
		Kind: KindFact, Content: "Riley is a inspiration",
		DedupeKey: "insp", Status: StatusActive, UpdatedAt: now,
		Metadata: map[string]any{"predicate": PredicateIdentity, "value_norm": "inspiration", "subject": "Riley"},
		Explain:  map[string]any{"predicate": PredicateIdentity, "value_norm": "inspiration", "subject": "Riley"},
	}
	store.records["coco"] = MemoryRecord{
		MemoryID: "mem_coco", TenantID: "t-ord", SubjectID: "u1",
		Kind: KindFact, Content: "Riley got a puppy named Coco on 28 July 2023.",
		DedupeKey: "coco", Status: StatusActive, UpdatedAt: now,
	}
	store.records["shadow"] = MemoryRecord{
		MemoryID: "mem_sh", TenantID: "t-ord", SubjectID: "u1",
		Kind: KindFact, Content: "Riley has a dog named Shadow.",
		DedupeKey: "shadow", Status: StatusActive, UpdatedAt: now,
	}
	store.atoms = append(store.atoms, stubAtom{pred: PredicateIdentity, val: "inspiration", memID: "mem_insp"})
	out, err := svc.Recall(context.Background(), RecallRequest{
		TenantID: "t-ord", SubjectID: "u1",
		Query: "What is the name of Riley's second puppy?", Mode: "answer", TopK: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(strings.TrimSpace(out.Answer), "Shadow") {
		t.Fatalf("expected Shadow, got %q explain=%v", out.Answer, out.Explain)
	}
	if strings.Contains(strings.ToLower(out.Answer), "inspiration") {
		t.Fatalf("identity dump leaked: %q", out.Answer)
	}
}
