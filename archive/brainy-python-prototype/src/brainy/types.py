from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from enum import Enum
from typing import Any


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


class InputType(str, Enum):
    MESSAGE = 'message'
    DOCUMENT = 'document'
    EVENT = 'event'
    OUTCOME = 'outcome'


class ArtifactType(str, Enum):
    EPISODE = 'episode'
    FACT = 'fact'
    PRINCIPLE = 'principle'
    TASTE_SIGNAL = 'taste_signal'
    PATTERN = 'pattern'


class BeliefStatus(str, Enum):
    CANDIDATE = 'candidate'
    ACTIVE = 'active'
    CHALLENGED = 'challenged'
    RETIRED = 'retired'


@dataclass(slots=True)
class Provenance:
    source: str
    source_type: str
    captured_at: datetime = field(default_factory=utc_now)
    model_version: str = 'n/a'
    actor: str = 'system'
    transform_chain: list[str] = field(default_factory=list)


@dataclass(slots=True)
class MemoryIngestEvent:
    event_id: str
    tenant_id: str
    actor_id: str
    input_type: InputType
    content: str
    tags: list[str] = field(default_factory=list)
    metadata: dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=utc_now)
    schema_version: str = '1.0'
    provenance: Provenance = field(default_factory=lambda: Provenance(source='unknown', source_type='unknown'))


@dataclass(slots=True)
class MemoryArtifact:
    artifact_id: str
    tenant_id: str
    artifact_type: ArtifactType
    content: str
    confidence: float
    evidence_event_ids: list[str]
    tags: list[str] = field(default_factory=list)
    metadata: dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=utc_now)
    updated_at: datetime = field(default_factory=utc_now)
    schema_version: str = '1.0'


@dataclass(slots=True)
class BeliefNode:
    belief_id: str
    tenant_id: str
    claim: str
    rank: float
    conviction: float
    status: BeliefStatus
    evidence_artifact_ids: list[str]
    supporting_event_ids: list[str]
    conflicting_belief_ids: list[str] = field(default_factory=list)
    metadata: dict[str, Any] = field(default_factory=dict)
    updated_at: datetime = field(default_factory=utc_now)
    schema_version: str = '1.0'


@dataclass(slots=True)
class OutcomeEvent:
    outcome_id: str
    tenant_id: str
    belief_id: str
    expected: float
    observed: float
    context: dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=utc_now)
    schema_version: str = '1.0'


@dataclass(slots=True)
class ReflectionJob:
    job_id: str
    tenant_id: str
    scope: str
    cadence: str
    constraints: dict[str, Any] = field(default_factory=dict)


@dataclass(slots=True)
class ContextQuery:
    tenant_id: str
    actor_id: str
    question: str
    tags: list[str] = field(default_factory=list)
    limit: int = 8


@dataclass(slots=True)
class ContextBundle:
    tenant_id: str
    query: str
    artifacts: list[MemoryArtifact]
    beliefs: list[BeliefNode]
    explanation: str
    schema_version: str = '1.0'


@dataclass(slots=True)
class HypothesisRecord:
    hypothesis_id: str
    title: str
    description: str
    component: str
    confidence: float
    evidence_refs: list[str]
    risk_level: str
    falsification_test: str
    owner: str
    status: str
    created_at: datetime = field(default_factory=utc_now)
    updated_at: datetime = field(default_factory=utc_now)
