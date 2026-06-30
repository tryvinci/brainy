from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository
from brainy.types import BeliefStatus, InputType


def test_belief_graph_builds_and_flags_conflicts() -> None:
    repository = InMemoryRepository()
    ingestion = IngestionEngine(repository)
    consolidation = ConsolidationEngine(repository)
    beliefs = BeliefGraphEngine(repository)

    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Creators are the best channel for launch.',
        source='note',
        source_type='doc',
    )
    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Creators are not the best channel for launch.',
        source='note',
        source_type='doc',
    )

    consolidation.consolidate('tenant-a')
    built = beliefs.build('tenant-a')
    assert built

    experiments = beliefs.reconcile_conflicts('tenant-a')
    assert experiments

    challenged = [belief for belief in repository.list_beliefs('tenant-a') if belief.status == BeliefStatus.CHALLENGED]
    assert challenged
