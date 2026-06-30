from __future__ import annotations

from uuid import uuid4

from brainy.repository import InMemoryRepository
from brainy.types import InputType, MemoryIngestEvent, Provenance


class IngestionEngine:
    def __init__(self, repository: InMemoryRepository) -> None:
        self.repository = repository

    def ingest(
        self,
        *,
        tenant_id: str,
        actor_id: str,
        input_type: InputType,
        content: str,
        source: str,
        source_type: str,
        tags: list[str] | None = None,
        metadata: dict[str, object] | None = None,
        model_version: str = 'n/a',
    ) -> MemoryIngestEvent:
        normalized = self._normalize_content(content)
        event = MemoryIngestEvent(
            event_id=f'evt_{uuid4().hex}',
            tenant_id=tenant_id,
            actor_id=actor_id,
            input_type=input_type,
            content=normalized,
            tags=tags or [],
            metadata=metadata or {},
            provenance=Provenance(
                source=source,
                source_type=source_type,
                model_version=model_version,
                actor=actor_id,
                transform_chain=['normalize_content'],
            ),
        )
        self.repository.add_event(event)
        self.repository.add_audit_event('ingest', {'event_id': event.event_id, 'tenant_id': tenant_id})
        return event

    @staticmethod
    def _normalize_content(content: str) -> str:
        lines = [line.strip() for line in content.strip().splitlines() if line.strip()]
        return ' '.join(lines)
