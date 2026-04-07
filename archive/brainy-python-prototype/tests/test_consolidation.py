from brainy.consolidation import ConsolidationEngine
from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository
from brainy.types import ArtifactType, InputType


def test_consolidation_extracts_expected_artifacts() -> None:
    repository = InMemoryRepository()
    ingestion = IngestionEngine(repository)
    consolidation = ConsolidationEngine(repository)

    ingestion.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Brand voice must stay consistent. This style performs with creators.',
        source='brief',
        source_type='doc',
        tags=['taste:bold'],
    )

    artifacts = consolidation.consolidate('tenant-a')

    types = {artifact.artifact_type for artifact in artifacts}
    assert ArtifactType.EPISODE in types
    assert ArtifactType.PRINCIPLE in types
    assert ArtifactType.TASTE_SIGNAL in types

    assert any(artifact.content == 'aesthetic-consideration' for artifact in artifacts if artifact.artifact_type == ArtifactType.TASTE_SIGNAL)
    assert repository.audit_log[-1]['event_type'] == 'consolidate'
