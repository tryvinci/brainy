from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from brainy.types import MemoryArtifact, MemoryIngestEvent


@dataclass(slots=True)
class InMemoryRepository:
    events: dict[str, MemoryIngestEvent] = field(default_factory=dict)
    artifacts: dict[str, MemoryArtifact] = field(default_factory=dict)
    audit_log: list[dict[str, Any]] = field(default_factory=list)

    def add_event(self, event: MemoryIngestEvent) -> None:
        self.events[event.event_id] = event

    def list_events(self, tenant_id: str) -> list[MemoryIngestEvent]:
        return [event for event in self.events.values() if event.tenant_id == tenant_id]

    def add_artifact(self, artifact: MemoryArtifact) -> None:
        self.artifacts[artifact.artifact_id] = artifact

    def list_artifacts(self, tenant_id: str) -> list[MemoryArtifact]:
        return [artifact for artifact in self.artifacts.values() if artifact.tenant_id == tenant_id]

    def add_audit_event(self, event_type: str, payload: dict[str, Any]) -> None:
        self.audit_log.append({'event_type': event_type, 'payload': payload})
