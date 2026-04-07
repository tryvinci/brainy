from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository
from brainy.types import InputType


def test_ingestion_normalizes_content_and_attaches_provenance() -> None:
    repository = InMemoryRepository()
    engine = IngestionEngine(repository)

    event = engine.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='  Line one.\n\n  Line two.  ',
        source='chat',
        source_type='conversation',
        tags=['brief'],
    )

    assert event.content == 'Line one. Line two.'
    assert event.provenance.source == 'chat'
    assert event.provenance.source_type == 'conversation'
    assert event.event_id in repository.events
    assert repository.audit_log[-1]['event_type'] == 'ingest'
