from __future__ import annotations

from copy import deepcopy
from dataclasses import dataclass, field
from datetime import datetime, timezone
from uuid import uuid4

from brainy.repository import InMemoryRepository


def utc_now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


@dataclass(slots=True)
class GovernanceEngine:
    repository: InMemoryRepository
    checkpoints: dict[str, dict[str, object]] = field(default_factory=dict)

    def create_checkpoint(self, label: str) -> str:
        checkpoint_id = f'chk_{uuid4().hex}'
        self.checkpoints[checkpoint_id] = {
            'label': label,
            'created_at': utc_now_iso(),
            'events': deepcopy(self.repository.events),
            'artifacts': deepcopy(self.repository.artifacts),
            'beliefs': deepcopy(self.repository.beliefs),
            'outcomes': deepcopy(self.repository.outcomes),
            'audit_log': deepcopy(self.repository.audit_log),
        }
        self.repository.add_audit_event('checkpoint_created', {'checkpoint_id': checkpoint_id, 'label': label})
        return checkpoint_id

    def rollback(self, checkpoint_id: str) -> None:
        snapshot = self.checkpoints[checkpoint_id]
        self.repository.events = deepcopy(snapshot['events'])
        self.repository.artifacts = deepcopy(snapshot['artifacts'])
        self.repository.beliefs = deepcopy(snapshot['beliefs'])
        self.repository.outcomes = deepcopy(snapshot['outcomes'])
        self.repository.add_audit_event('rollback', {'checkpoint_id': checkpoint_id})

    def explain_decision(self, belief_id: str) -> dict[str, object]:
        belief = self.repository.beliefs[belief_id]
        artifacts = [
            self.repository.artifacts[artifact_id]
            for artifact_id in belief.evidence_artifact_ids
            if artifact_id in self.repository.artifacts
        ]
        return {
            'belief_id': belief.belief_id,
            'claim': belief.claim,
            'status': belief.status.value,
            'conviction': belief.conviction,
            'rank': belief.rank,
            'evidence': [artifact.content for artifact in artifacts],
            'supporting_events': belief.supporting_event_ids,
            'conflicts': belief.conflicting_belief_ids,
        }

    def audit_events(self, event_type: str | None = None) -> list[dict[str, object]]:
        if event_type is None:
            return list(self.repository.audit_log)
        return [event for event in self.repository.audit_log if event['event_type'] == event_type]
