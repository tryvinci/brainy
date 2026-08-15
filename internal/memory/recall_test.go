package memory

import (
	"context"
	"strings"
	"testing"
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
	got, ok = slotValueFromMemoryContent("Riley: This book I read last year reminds me. [visible text: LIFE IS ELSEWHERE]")
	if !ok || !strings.Contains(strings.ToLower(got), "life is elsewhere") {
		t.Fatalf("visible-text title, got %q ok=%v", got, ok)
	}
	if strings.EqualFold(got, "Elsewhere") {
		t.Fatalf("visible-text title split on is: %q", got)
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
