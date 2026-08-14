package memory

import "testing"

func TestPlanQueryTemporalAndEnumerate(t *testing.T) {
	plan := PlanQuery("what is Alex currently working on?", nil)
	if !plan.NeedsTemporal {
		t.Fatal("expected NeedsTemporal")
	}
	if plan.PrimaryIntent != IntentCurrentState && !hasIntent(plan.Intents, IntentCurrentState) {
		t.Fatalf("intents=%v", plan.Intents)
	}
	found := false
	for _, tool := range plan.Tools {
		if tool == "temporal_resolve" {
			found = true
		}
	}
	if !found {
		t.Fatalf("tools=%v", plan.Tools)
	}

	list := PlanQuery("list all hobbies Caroline enjoys", nil)
	if !list.NeedsEnumeration {
		t.Fatal("expected NeedsEnumeration")
	}
	if list.PreferredModeHint != "enumerate" {
		t.Fatalf("mode hint=%q", list.PreferredModeHint)
	}

	acts := PlanQuery("What activities does Jordan enjoy?", nil)
	foundRel := false
	for _, hop := range acts.Hops {
		if hop.Kind == "follow_relation" && hop.Predicate == PredicateActivity {
			foundRel = true
		}
	}
	if !foundRel {
		t.Fatalf("expected follow_relation activity hop, hops=%+v intents=%v", acts.Hops, acts.Intents)
	}
}

func TestBuildEvidencePacket(t *testing.T) {
	plan := PlanQuery("what is the current ticket status", nil)
	pkt := BuildEvidencePacket(plan, []SearchResult{
		{MemoryID: "m1", Content: "Ticket T-1 is open"},
	}, map[string]any{"temporal_answer": "status: open"})
	if len(pkt.MemoryIDs) != 1 || pkt.TemporalAnswer == "" {
		t.Fatalf("%+v", pkt)
	}
	if pkt.Coverage["satisfied"] != true {
		t.Fatalf("coverage=%v", pkt.Coverage)
	}
}
