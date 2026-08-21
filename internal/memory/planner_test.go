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
	if hopComposeAllowed("When did Jordan go to the support group?") {
		t.Fatal("when-questions must not dump hop values as answers")
	}
	whenEnt := ""
	for _, hop := range when.Hops {
		if hop.Kind == "resolve_entity" {
			whenEnt = hop.Entity
		}
	}
	if !strings.EqualFold(whenEnt, "jordan") {
		t.Fatalf("when-event must hop the person, hops=%+v", when.Hops)
	}

	injWhen := PlanQuery("When did Jordan get an ankle injury in 2023?", nil)
	foundInjHealth := false
	injEnts := map[string]bool{}
	for _, hop := range injWhen.Hops {
		if hop.Kind == "resolve_entity" {
			injEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateHealth {
			foundInjHealth = true
		}
	}
	if !injEnts["jordan"] {
		t.Fatalf("injury when-query must hop the person, hops=%+v", injWhen.Hops)
	}
	if !foundInjHealth {
		t.Fatalf("injury when-query must hop health, hops=%+v", injWhen.Hops)
	}

	xfer := PlanQuery("What kind of healthy food suggestions has Evan given to Sam?", nil)
	xferEnts := map[string]bool{}
	foundPrefXfer := false
	for _, hop := range xfer.Hops {
		if hop.Kind == "resolve_entity" {
			xferEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicatePreference {
			foundPrefXfer = true
		}
	}
	if !xferEnts["evan"] {
		t.Fatalf("transfer must hop the giver, hops=%+v", xfer.Hops)
	}
	if xferEnts["sam"] {
		t.Fatalf("transfer recipient must not be a join entity, hops=%+v", xfer.Hops)
	}
	if !foundPrefXfer {
		t.Fatalf("expected preference hop for given-to, hops=%+v", xfer.Hops)
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
	foundPossCount := false
	for _, hop := range count.Hops {
		if hop.Kind == "resolve_entity" {
			countEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicatePossession {
			foundPossCount = true
		}
	}
	if !countEnts["calvin"] {
		t.Fatalf("count subject must hop the person, hops=%+v", count.Hops)
	}
	if count.PrimaryIntent != IntentAggregation {
		t.Fatalf("expected aggregation intent, intents=%v", count.Intents)
	}
	if !foundPossCount {
		t.Fatalf("expected possession hop for how-many own, hops=%+v", count.Hops)
	}

	inj := PlanQuery("How many times has Jordan injured his ankle?", nil)
	foundHealth := false
	for _, hop := range inj.Hops {
		if hop.Predicate == PredicateHealth {
			foundHealth = true
		}
	}
	if !foundHealth {
		t.Fatalf("injury count must hop health, hops=%+v", inj.Hops)
	}

	polar := PlanQuery("Has Riley tried surfing?", nil)
	if !polar.NeedsMultiHop {
		t.Fatalf("polar query must hop, intents=%v", polar.Intents)
	}
	polarEnts := map[string]bool{}
	foundTriedActivity := false
	for _, hop := range polar.Hops {
		if hop.Kind == "resolve_entity" {
			polarEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateActivity {
			foundTriedActivity = true
		}
	}
	if !polarEnts["riley"] {
		t.Fatalf("expected riley hop, hops=%+v", polar.Hops)
	}
	if !foundTriedActivity {
		t.Fatalf("tried polar must hop activity, hops=%+v", polar.Hops)
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

	pets := PlanQuery("What are Riley's dogs' names?", nil)
	if !pets.NeedsEnumeration {
		t.Fatalf("expected enumeration for names list, intents=%v", pets.Intents)
	}
	petEnts := map[string]bool{}
	foundPossPets := false
	for _, hop := range pets.Hops {
		if hop.Kind == "resolve_entity" {
			petEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicatePossession {
			foundPossPets = true
		}
	}
	if !petEnts["riley"] {
		t.Fatalf("expected riley hop, hops=%+v", pets.Hops)
	}
	if petEnts["dogs"] || petEnts["names"] {
		t.Fatalf("slot nouns must not hop as entities, hops=%+v", pets.Hops)
	}
	if !foundPossPets {
		t.Fatalf("expected possession hop for dogs' names, hops=%+v", pets.Hops)
	}

	inst := PlanQuery("What instruments does Jordan play?", nil)
	instEnts := map[string]bool{}
	foundSkill := false
	for _, hop := range inst.Hops {
		if hop.Kind == "resolve_entity" {
			instEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateSkill {
			foundSkill = true
		}
	}
	if !instEnts["jordan"] {
		t.Fatalf("expected jordan hop, hops=%+v", inst.Hops)
	}
	if instEnts["instruments"] {
		t.Fatalf("instrument slot must not hop as entity, hops=%+v", inst.Hops)
	}
	if !foundSkill {
		t.Fatalf("expected skill hop for instruments, hops=%+v", inst.Hops)
	}

	tricks := PlanQuery("What kind of tricks do James's pets know?", nil)
	trickEnts := map[string]bool{}
	foundTrickSkill := false
	for _, hop := range tricks.Hops {
		if hop.Kind == "resolve_entity" {
			trickEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateSkill {
			foundTrickSkill = true
		}
	}
	if !trickEnts["james"] {
		t.Fatalf("expected james hop, hops=%+v", tricks.Hops)
	}
	if trickEnts["pets"] || trickEnts["tricks"] {
		t.Fatalf("slot nouns must not hop as entities, hops=%+v", tricks.Hops)
	}
	if !foundTrickSkill {
		t.Fatalf("expected skill hop for pets' tricks, hops=%+v", tricks.Hops)
	}

	unwind := PlanQuery("What does Riley do to unwind?", nil)
	if !unwind.NeedsEnumeration {
		t.Fatalf("expected enumeration for do-to unwind, intents=%v", unwind.Intents)
	}
	foundUnwindAct := false
	for _, hop := range unwind.Hops {
		if hop.Predicate == PredicateActivity {
			foundUnwindAct = true
		}
	}
	if !foundUnwindAct {
		t.Fatalf("expected activity hop for unwind, hops=%+v", unwind.Hops)
	}

	visit := PlanQuery("which country has Tim visited most frequently in his travels?", nil)
	foundVisitAct := false
	for _, hop := range visit.Hops {
		if hop.Predicate == PredicateActivity {
			foundVisitAct = true
		}
	}
	if !foundVisitAct {
		t.Fatalf("visited country must hop activity not origin-first, hops=%+v", visit.Hops)
	}
	if !visit.NeedsEnumeration {
		t.Fatalf("superlative must enumerate, intents=%v", visit.Intents)
	}

	who := PlanQuery("Who supports Calvin in tough times?", nil)
	foundFam := false
	whoEnts := map[string]bool{}
	for _, hop := range who.Hops {
		if hop.Kind == "resolve_entity" {
			whoEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateFamilyMember {
			foundFam = true
		}
	}
	if !whoEnts["calvin"] {
		t.Fatalf("expected calvin hop, hops=%+v", who.Hops)
	}
	if !foundFam {
		t.Fatalf("who-supports must hop family, hops=%+v", who.Hops)
	}

	childItems := PlanQuery("What items does John mention having as a child?", nil)
	foundChildPoss := false
	for _, hop := range childItems.Hops {
		if hop.Predicate == PredicatePossession {
			foundChildPoss = true
		}
		if hop.Predicate == PredicateFamilyMember {
			t.Fatalf("childhood items must not hop family, hops=%+v", childItems.Hops)
		}
	}
	if !foundChildPoss {
		t.Fatalf("expected possession hop for childhood items, hops=%+v", childItems.Hops)
	}

	whereKin := PlanQuery("Where did Jolene and her partner find a cool diving spot?", nil)
	foundWhereFam, foundWhereSrc, foundWhereDest := false, false, false
	for _, hop := range whereKin.Hops {
		if hop.Kind == "follow_relation" && hop.Predicate == PredicateFamilyMember {
			foundWhereFam = true
		}
		if hop.Predicate == PredicateActivity || hop.Predicate == PredicateEvent || hop.Predicate == PredicateResidence {
			if len(hop.DependsOn) == 1 && hop.DependsOn[0] == "e1" {
				foundWhereSrc = true
			}
			if len(hop.DependsOn) == 1 && hop.DependsOn[0] == "e_rel" {
				foundWhereDest = true
			}
		}
	}
	if !foundWhereFam || !foundWhereSrc || !foundWhereDest {
		t.Fatalf("where+kinship must fetch source and dest, hops=%+v", whereKin.Hops)
	}

	group := PlanQuery("What outdoor activities has John done with his colleagues?", nil)
	groupEnts := map[string]bool{}
	foundGroupAct := false
	for _, hop := range group.Hops {
		if hop.Kind == "resolve_entity" {
			groupEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateActivity {
			foundGroupAct = true
		}
	}
	if !groupEnts["john"] {
		t.Fatalf("group-with must hop the person, hops=%+v", group.Hops)
	}
	if groupEnts["colleagues"] {
		t.Fatalf("colleagues must not be a hop entity, hops=%+v", group.Hops)
	}
	if !foundGroupAct {
		t.Fatalf("group-with activities must hop activity, hops=%+v", group.Hops)
	}
	if toks := groupCompanionTokens("What outdoor activities has John done with his colleagues?"); len(toks) == 0 {
		t.Fatal("expected group companion tokens")
	}
	if toks := listHeadModifierTokens("What outdoor activities has John done with his colleagues?"); len(toks) != 1 || toks[0] != "outdoor" {
		t.Fatalf("expected outdoor modifier, toks=%v", toks)
	}
	if toks := listHeadModifierTokens("Which community activities have Riley and Casey participated in?"); len(toks) != 0 {
		t.Fatalf("community is a join cue, not a list-head adjective, toks=%v", toks)
	}
	if toks := listHeadModifierTokens("What unhealthy snacks does Casey avoid?"); len(toks) != 1 || toks[0] != "unhealthy" {
		t.Fatalf("expected unhealthy snack modifier, toks=%v", toks)
	}
	if toks := negatedModifierTokens("What kind of unhealthy snacks does Casey enjoy eating?"); len(toks) != 1 || toks[0] != "unhealthy" {
		t.Fatalf("expected un- modifier, toks=%v", toks)
	}
	if !looksListQuery(tokenize("What kind of unhealthy snacks does Casey enjoy eating?")) {
		t.Fatal("snack lists must enumerate")
	}
	if toks := listHeadModifierTokens("What activities does Riley enjoy?"); len(toks) != 0 {
		t.Fatalf("bare activities must not take a modifier, toks=%v", toks)
	}
	if !looksLocationListQuery("Which locations does Riley practice her yoga at?") {
		t.Fatal("practice location query must be a location list")
	}
	if looksLocationListQuery("Which community activities have Riley and Casey participated in?") {
		t.Fatal("community activities must not be a location list")
	}
	if !looksLocationListQuery("Where does Riley practice yoga?") {
		t.Fatal("where+practice must be a location list")
	}
	if hopComposeAllowed("Which locations does Riley practice her yoga at?") {
		t.Fatal("location lists must not dump hop values")
	}
	if hopComposeAllowed("Has Riley tried surfing?") {
		t.Fatal("polar queries must not dump hop values")
	}
	if toks := practiceObjectTokens("Which locations does Riley practice her yoga at?"); len(toks) != 1 || toks[0] != "yoga" {
		t.Fatalf("expected yoga practice object, toks=%v", toks)
	}
	if toks := listHeadModifierTokens("What are Riley's pets' names?"); len(toks) != 0 {
		t.Fatalf("pets names must not take a list-head modifier, toks=%v", toks)
	}
	if toks := nameCueTokens("What are Riley's pets' names?"); len(toks) != 1 || toks[0] != "named" {
		t.Fatalf("expected named cue for names lists, toks=%v", toks)
	}
	if toks := childhoodClauseTokens("What items did Riley have as a child?"); len(toks) == 0 {
		t.Fatal("expected childhood clause tokens")
	}

	conseq := PlanQuery("What did Audrey get with having so many dogs?", nil)
	foundConseqHealth, foundConseqPoss := false, false
	for _, hop := range conseq.Hops {
		if hop.Predicate == PredicateHealth {
			foundConseqHealth = true
		}
		if hop.Predicate == PredicatePossession {
			foundConseqPoss = true
		}
	}
	if !foundConseqHealth {
		t.Fatalf("get-with-having must hop health, hops=%+v", conseq.Hops)
	}
	if foundConseqPoss {
		t.Fatalf("get-with-having must not hop possession lists, hops=%+v", conseq.Hops)
	}
	if conseq.NeedsEnumeration {
		t.Fatalf("consequence must not enumerate, intents=%v", conseq.Intents)
	}

	kidsCount := PlanQuery("How many children does Riley have?", nil)
	foundKidsFam := false
	for _, hop := range kidsCount.Hops {
		if hop.Predicate == PredicateFamilyMember {
			foundKidsFam = true
		}
	}
	if !foundKidsFam {
		t.Fatalf("how-many children must hop family, hops=%+v", kidsCount.Hops)
	}

	ev := PlanQuery("What events is Maria planning for the homeless shelter fundraiser?", nil)
	foundEv := false
	for _, hop := range ev.Hops {
		if hop.Predicate == PredicateEvent || hop.Predicate == PredicatePlan {
			foundEv = true
		}
	}
	if !foundEv {
		t.Fatalf("planning-for events must hop event/plan, hops=%+v", ev.Hops)
	}
	if toks := forClauseTokens("What events is Maria planning for the homeless shelter fundraiser?"); len(toks) == 0 {
		t.Fatal("expected for-clause tokens")
	}

	dualAct := PlanQuery("Which community activities have Riley and Casey participated in?", nil)
	dualEnts := map[string]bool{}
	foundDualAct := false
	for _, hop := range dualAct.Hops {
		if hop.Kind == "resolve_entity" {
			dualEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateActivity {
			foundDualAct = true
		}
	}
	if !dualEnts["riley"] || !dualEnts["casey"] {
		t.Fatalf("dual community must hop both people, hops=%+v", dualAct.Hops)
	}
	if !foundDualAct {
		t.Fatalf("dual community must hop activity, hops=%+v", dualAct.Hops)
	}
	if toks := inCommunityTokens("Which community activities have Riley and Casey participated in?"); len(toks) != 0 {
		t.Fatalf("dual community list is not a named in-the-X community, toks=%v", toks)
	}

	namedCom := PlanQuery("In what ways is Riley participating in the civic community?", nil)
	foundNamedAct, foundNamedAff := false, false
	namedEnts := map[string]bool{}
	for _, hop := range namedCom.Hops {
		if hop.Kind == "resolve_entity" {
			namedEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateActivity {
			foundNamedAct = true
		}
		if hop.Predicate == PredicateAffiliation {
			foundNamedAff = true
		}
	}
	if !namedEnts["riley"] {
		t.Fatalf("named community must hop the person, hops=%+v", namedCom.Hops)
	}
	if !foundNamedAct {
		t.Fatalf("named community must hop activity, hops=%+v", namedCom.Hops)
	}
	if !foundNamedAff {
		t.Fatalf("named community must hop affiliation, hops=%+v", namedCom.Hops)
	}
	if toks := inCommunityTokens("In what ways is Riley participating in the civic community?"); len(toks) != 1 || toks[0] != "civic" {
		t.Fatalf("expected civic community token, toks=%v", toks)
	}
	if toks := inCommunityTokens("In what ways is Riley participating in the community?"); len(toks) != 0 {
		t.Fatalf("unnamed community must not filter, toks=%v", toks)
	}

	told := PlanQuery("Who did Evan tell about his marriage?", nil)
	foundToldFam := false
	toldEnts := map[string]bool{}
	for _, hop := range told.Hops {
		if hop.Kind == "resolve_entity" {
			toldEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateFamilyMember {
			foundToldFam = true
		}
	}
	if !toldEnts["evan"] {
		t.Fatalf("who-told must hop the speaker, hops=%+v", told.Hops)
	}
	if !foundToldFam {
		t.Fatalf("who-told must hop family, hops=%+v", told.Hops)
	}

	journey := PlanQuery("What are some changes Riley has faced during her journey?", nil)
	foundJourneyID := false
	journeyEnts := map[string]bool{}
	for _, hop := range journey.Hops {
		if hop.Kind == "resolve_entity" {
			journeyEnts[strings.ToLower(hop.Entity)] = true
		}
		if hop.Predicate == PredicateIdentity {
			foundJourneyID = true
		}
	}
	if !journeyEnts["riley"] {
		t.Fatalf("journey changes must hop the person, hops=%+v", journey.Hops)
	}
	if !foundJourneyID {
		t.Fatalf("journey changes must hop identity, hops=%+v", journey.Hops)
	}
	if toks := duringClauseTokens("What are some changes Riley has faced during her journey?"); len(toks) != 0 {
		t.Fatalf("unnamed journey must not filter, toks=%v", toks)
	}
	if toks := duringClauseTokens("What are some changes Riley has faced during her recovery journey?"); len(toks) != 1 || toks[0] != "recovery" {
		t.Fatalf("expected recovery during-token, toks=%v", toks)
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
