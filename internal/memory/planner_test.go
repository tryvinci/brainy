package memory

import (
	"strings"
	"testing"
)

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

	origin := PlanQuery("Where did Jordan move from 4 years ago?", nil)
	foundOrigin := false
	for _, hop := range origin.Hops {
		if hop.Kind == "follow_relation" && hop.Predicate == PredicateOrigin && hop.Output == "ans" {
			foundOrigin = true
		}
	}
	if !foundOrigin {
		t.Fatalf("expected origin follow_relation as ans, hops=%+v", origin.Hops)
	}

	career := PlanQuery("What career path has Jordan decided to pursue?", nil)
	foundOcc, foundID := false, false
	for _, hop := range career.Hops {
		if hop.Kind == "follow_relation" && hop.Predicate == PredicateOccupation {
			foundOcc = true
		}
		if hop.Predicate == PredicateIdentity {
			foundID = true
		}
	}
	if !foundOcc || !foundID {
		t.Fatalf("expected occupation+identity hops, hops=%+v", career.Hops)
	}

	when := PlanQuery("When did Jordan go to the support group?", nil)
	if len(when.Hops) != 0 {
		t.Fatalf("when-questions must not dump event hops, hops=%+v", when.Hops)
	}

	kids := PlanQuery("What do Riley's kids like?", nil)
	ent := ""
	for _, hop := range kids.Hops {
		if hop.Kind == "resolve_entity" {
			ent = hop.Entity
		}
	}
	if !strings.EqualFold(ent, "riley") {
		t.Fatalf("expected entity riley, got %q hops=%+v", ent, kids.Hops)
	}

	research := PlanQuery("What did Jordan research?", nil)
	foundPlan := false
	for _, hop := range research.Hops {
		if hop.Predicate == PredicatePlan && strings.Contains(strings.ToLower(hop.Probe), "research") {
			foundPlan = true
		}
		if hop.Predicate == PredicateIdentity {
			t.Fatalf("research hops must not fetch identity, hops=%+v", research.Hops)
		}
	}
	if !foundPlan {
		t.Fatalf("expected plan hop with research probe, hops=%+v intents=%v", research.Hops, research.Intents)
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
