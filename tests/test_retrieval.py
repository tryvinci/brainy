from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository
from brainy.retrieval import RetrievalEngine
from brainy.types import ContextQuery, InputType


def test_retrieval_returns_taste_aware_context_bundle() -> None:
    repository = InMemoryRepository()
    ingestion = IngestionEngine(repository)
    consolidation = ConsolidationEngine(repository)
    belief_graph = BeliefGraphEngine(repository)
    retrieval = RetrievalEngine(repository)

    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Our aesthetic style resonates with creator communities.',
        source='brief',
        source_type='doc',
        tags=['taste:editorial'],
    )
    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Email frequency is weekly for this campaign.',
        source='ops',
        source_type='doc',
    )

    consolidation.consolidate('tenant-a')
    belief_graph.build('tenant-a')

    bundle = retrieval.retrieve(
        ContextQuery(
            tenant_id='tenant-a',
            actor_id='user-1',
            question='What style resonates with creators?',
            tags=['taste'],
            limit=3,
        )
    )

    assert bundle.artifacts
    assert bundle.beliefs
    assert 'Context assembled' in bundle.explanation
    assert repository.audit_log[-1]['event_type'] == 'retrieve'
