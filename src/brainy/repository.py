from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from brainy.types import BeliefNode, MemoryArtifact, MemoryIngestEvent, OutcomeEvent


@dataclass(slots=True)
class InMemoryRepository:
    events: dict[str, MemoryIngestEvent] = field(default_factory=dict)
    artifacts: dict[str, MemoryArtifact] = field(default_factory=dict)
    beliefs: dict[str, BeliefNode] = field(default_factory=dict)
    outcomes: dict[str, OutcomeEvent] = field(default_factory=dict)
    audit_log: list[dict[str, Any]] = field(default_factory=list)

    def add_event(self, event: MemoryIngestEvent) -> None:
        self.events[event.event_id] = event

    def list_events(self, tenant_id: str) -> list[MemoryIngestEvent]:
        return [event for event in self.events.values() if event.tenant_id == tenant_id]

    def add_artifact(self, artifact: MemoryArtifact) -> None:
        self.artifacts[artifact.artifact_id] = artifact

    def list_artifacts(self, tenant_id: str) -> list[MemoryArtifact]:
        return [artifact for artifact in self.artifacts.values() if artifact.tenant_id == tenant_id]

    def add_belief(self, belief: BeliefNode) -> None:
        self.beliefs[belief.belief_id] = belief

    def list_beliefs(self, tenant_id: str) -> list[BeliefNode]:
        return [belief for belief in self.beliefs.values() if belief.tenant_id == tenant_id]

    def add_outcome(self, outcome: OutcomeEvent) -> None:
        self.outcomes[outcome.outcome_id] = outcome

    def list_outcomes(self, tenant_id: str) -> list[OutcomeEvent]:
        return [outcome for outcome in self.outcomes.values() if outcome.tenant_id == tenant_id]

    def add_audit_event(self, event_type: str, payload: dict[str, Any]) -> None:
        self.audit_log.append({'event_type': event_type, 'payload': payload})
