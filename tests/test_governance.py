from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.governance import GovernanceEngine
from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository
from brainy.types import InputType


def test_governance_checkpoint_rollback_and_explain() -> None:
    repository = InMemoryRepository()
    ingestion = IngestionEngine(repository)
    consolidation = ConsolidationEngine(repository)
    belief_graph = BeliefGraphEngine(repository)
    governance = GovernanceEngine(repository)

    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Brand voice must stay premium and confident.',
        source='brief',
        source_type='doc',
    )
    consolidation.consolidate('tenant-a')
    beliefs = belief_graph.build('tenant-a')
    target = beliefs[0]

    checkpoint_id = governance.create_checkpoint('before-change')
    original_conviction = repository.beliefs[target.belief_id].conviction

    repository.beliefs[target.belief_id].conviction = 0.1
    governance.rollback(checkpoint_id)

    restored = repository.beliefs[target.belief_id]
    assert restored.conviction == original_conviction

    explanation = governance.explain_decision(target.belief_id)
    assert explanation['belief_id'] == target.belief_id
    assert explanation['evidence']

    checkpoint_events = governance.audit_events('checkpoint_created')
    assert checkpoint_events
