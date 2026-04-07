from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.ingestion import IngestionEngine
from brainy.reflection import ReflectionEngine
from brainy.repository import InMemoryRepository
from brainy.types import BeliefStatus, InputType, ReflectionJob


def test_reflection_downgrades_conviction_on_outcome_failure() -> None:
    repository = InMemoryRepository()
    ingestion = IngestionEngine(repository)
    consolidation = ConsolidationEngine(repository)
    belief_graph = BeliefGraphEngine(repository)
    reflection = ReflectionEngine(repository)

    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Creators are the best launch channel.',
        source='brief',
        source_type='doc',
    )
    consolidation.consolidate('tenant-a')
    beliefs = belief_graph.build('tenant-a')
    target = beliefs[0]

    reflection.record_outcome(
        tenant_id='tenant-a',
        belief_id=target.belief_id,
        expected=0.9,
        observed=0.4,
    )
    reflection.record_outcome(
        tenant_id='tenant-a',
        belief_id=target.belief_id,
        expected=0.8,
        observed=0.2,
    )

    result = reflection.run_reflection(
        ReflectionJob(
            job_id='job-1',
            tenant_id='tenant-a',
            scope='belief_updates',
            cadence='event-triggered',
        )
    )

    updated = repository.beliefs[target.belief_id]
    assert result['updated'] == 1
    assert result['challenged'] == 1
    assert updated.status == BeliefStatus.CHALLENGED
    assert updated.conviction < 1.0
