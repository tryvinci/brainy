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

	like := PlanQuery("What animal do both Nate and Joanna like?", nil)
	foundPref := false
	ents := map[string]bool{}
	for _, hop := range like.Hops {
		if hop.Kind == "resolve_entity" {
			ents[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicatePreference {
			foundPref = true
		}
	}
	if ents["animal"] {
		t.Fatalf("topic noun must not be the hop entity, hops=%+v", like.Hops)
	}
	if !ents["nate"] && !ents["joanna"] {
		t.Fatalf("expected person hop entity, hops=%+v", like.Hops)
	}
	if !foundPref {
		t.Fatalf("expected preference hop for like-query, hops=%+v", like.Hops)
	}

	coord := PlanQuery("What similar collectible do Tim and John own?", nil)
	coordEnts := map[string]bool{}
	foundPoss := false
	for _, hop := range coord.Hops {
		if hop.Kind == "resolve_entity" {
			coordEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicatePossession {
			foundPoss = true
		}
	}
	if !coordEnts["tim"] || !coordEnts["john"] {
		t.Fatalf("coordinated names must both hop, hops=%+v", coord.Hops)
	}
	if coordEnts["collectible"] {
		t.Fatalf("topic noun must not be a hop entity, hops=%+v", coord.Hops)
	}
	if !foundPoss {
		t.Fatalf("expected possession hop for own-query, hops=%+v", coord.Hops)
	}

	bothAfter := PlanQuery("What do Nate and Joanna both like?", nil)
	bothEnts := map[string]bool{}
	for _, hop := range bothAfter.Hops {
		if hop.Kind == "resolve_entity" {
			bothEnts[strings.ToLower(hop.Entity)] = true
		}
	}
	if !bothEnts["nate"] || !bothEnts["joanna"] {
		t.Fatalf("both-after-names must hop both people, hops=%+v", bothAfter.Hops)
	}

	attrib := PlanQuery("What is Alex's occupation according to Dana?", nil)
	attribEnts := map[string]bool{}
	for _, hop := range attrib.Hops {
		if hop.Kind == "resolve_entity" {
			attribEnts[strings.ToLower(hop.Entity)] = true
		}
	}
	if !attribEnts["alex"] {
		t.Fatalf("expected alex hop, hops=%+v", attrib.Hops)
	}
	if attribEnts["dana"] {
		t.Fatalf("attribution source must not be a join entity, hops=%+v", attrib.Hops)
	}

	withJoin := PlanQuery("What activities does Jordan enjoy with Casey?", nil)
	withEnts := map[string]bool{}
	for _, hop := range withJoin.Hops {
		if hop.Kind == "resolve_entity" {
			withEnts[strings.ToLower(hop.Entity)] = true
		}
	}
	if !withEnts["jordan"] || !withEnts["casey"] {
		t.Fatalf("with-person must hop both, hops=%+v", withJoin.Hops)
	}

	count := PlanQuery("How many cars does Calvin own?", nil)
	countEnts := map[string]bool{}
	for _, hop := range count.Hops {
		if hop.Kind == "resolve_entity" {
			countEnts[strings.ToLower(hop.Entity)] = true
		}
	}
	if !countEnts["calvin"] {
		t.Fatalf("count subject must hop the person, hops=%+v", count.Hops)
	}

	kin := PlanQuery("What were Alex's mother's hobbies?", nil)
	foundFamily, foundActivity, relDep := false, false, false
	for _, hop := range kin.Hops {
		if hop.Kind == "follow_relation" && hop.Predicate == PredicateFamilyMember {
			foundFamily = true
			if len(hop.DependsOn) > 0 && hop.DependsOn[0] == "e1" {
				relDep = true
			}
		}
		if hop.Predicate == PredicateActivity && (hop.Kind == "fetch_predicate" || hop.Kind == "follow_relation") {
			foundActivity = true
			if len(hop.DependsOn) == 1 && hop.DependsOn[0] == "e_rel" {
				relDep = relDep && true
			} else {
				t.Fatalf("activity fetch must depend on kinship dest, hop=%+v", hop)
			}
		}
	}
	if !foundFamily || !foundActivity {
		t.Fatalf("expected kinship follow then activity, hops=%+v", kin.Hops)
	}
	if !relDep {
		t.Fatalf("kinship hops must chain e1 → e_rel → ans, hops=%+v", kin.Hops)
	}

	pronoun := PlanQuery("What do they like?", nil)
	for _, hop := range pronoun.Hops {
		if hop.Kind == "resolve_entity" && strings.EqualFold(hop.Entity, "they") {
			t.Fatalf("pronoun must not be a hop entity, hops=%+v", pronoun.Hops)
		}
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
	if len(pkt.ContextEvidence) != 1 || pkt.ContextEvidence[0].MemoryID != "m1" {
		t.Fatalf("typed context_evidence=%+v", pkt.ContextEvidence)
	}
	if pkt.ContextEvidence[0].Role != "context" || pkt.ContextEvidence[0].Span == "" {
		t.Fatalf("expected context span, got %+v", pkt.ContextEvidence[0])
	}
}
