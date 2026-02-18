from __future__ import annotations

from uuid import uuid4

from brainy.repository import InMemoryRepository
from brainy.types import ArtifactType, MemoryArtifact, MemoryIngestEvent


class ConsolidationEngine:
    def __init__(self, repository: InMemoryRepository) -> None:
        self.repository = repository

    def consolidate(self, tenant_id: str) -> list[MemoryArtifact]:
        events = self.repository.list_events(tenant_id)
        artifacts: list[MemoryArtifact] = []
        for event in events:
            if self._already_consolidated(event.event_id):
                continue
            artifacts.extend(self._artifacts_from_event(event))
        for artifact in artifacts:
            self.repository.add_artifact(artifact)
        if artifacts:
            self.repository.add_audit_event(
                'consolidate',
                {'tenant_id': tenant_id, 'artifact_ids': [artifact.artifact_id for artifact in artifacts]},
            )
        return artifacts

    def _already_consolidated(self, event_id: str) -> bool:
        return any(event_id in artifact.evidence_event_ids for artifact in self.repository.artifacts.values())

    def _artifacts_from_event(self, event: MemoryIngestEvent) -> list[MemoryArtifact]:
        artifacts = [
            self._create_artifact(
                tenant_id=event.tenant_id,
                artifact_type=ArtifactType.EPISODE,
                content=event.content,
                confidence=0.7,
                evidence_event_ids=[event.event_id],
                tags=['episode', *event.tags],
            )
        ]

        for sentence in self._split_sentences(event.content):
            if self._looks_like_fact(sentence):
                artifacts.append(
                    self._create_artifact(
                        tenant_id=event.tenant_id,
                        artifact_type=ArtifactType.FACT,
                        content=sentence,
                        confidence=0.6,
                        evidence_event_ids=[event.event_id],
                        tags=['fact'],
                    )
                )
            if self._looks_like_principle(sentence):
                artifacts.append(
                    self._create_artifact(
                        tenant_id=event.tenant_id,
                        artifact_type=ArtifactType.PRINCIPLE,
                        content=sentence,
                        confidence=0.8,
                        evidence_event_ids=[event.event_id],
                        tags=['principle'],
                    )
                )

        taste_markers = self._extract_taste_markers(event.content, event.tags)
        for marker in taste_markers:
            artifacts.append(
                self._create_artifact(
                    tenant_id=event.tenant_id,
                    artifact_type=ArtifactType.TASTE_SIGNAL,
                    content=marker,
                    confidence=0.65,
                    evidence_event_ids=[event.event_id],
                    tags=['taste_signal'],
                )
            )
        return artifacts

    @staticmethod
    def _create_artifact(
        *,
        tenant_id: str,
        artifact_type: ArtifactType,
        content: str,
        confidence: float,
        evidence_event_ids: list[str],
        tags: list[str],
    ) -> MemoryArtifact:
        return MemoryArtifact(
            artifact_id=f'art_{uuid4().hex}',
            tenant_id=tenant_id,
            artifact_type=artifact_type,
            content=content,
            confidence=confidence,
            evidence_event_ids=evidence_event_ids,
            tags=tags,
        )

    @staticmethod
    def _split_sentences(content: str) -> list[str]:
        chunks = [chunk.strip() for chunk in content.replace('?', '.').replace('!', '.').split('.')]
        return [chunk for chunk in chunks if chunk]

    @staticmethod
    def _looks_like_fact(sentence: str) -> bool:
        lower = sentence.lower()
        return ' is ' in lower or ' are ' in lower or lower.startswith('the ')

    @staticmethod
    def _looks_like_principle(sentence: str) -> bool:
        lower = sentence.lower()
        return 'always' in lower or 'never' in lower or 'must' in lower

    @staticmethod
    def _extract_taste_markers(content: str, tags: list[str]) -> list[str]:
        lower = content.lower()
        markers: list[str] = []
        if any(token in lower for token in ['taste', 'aesthetic', 'voice', 'style']):
            markers.append('aesthetic-consideration')
        if any(token in lower for token in ['perform', 'resonate', 'engage']):
            markers.append('performance-signal')
        if 'identity' in lower:
            markers.append('identity-prior')
        for tag in tags:
            if tag.startswith('taste:'):
                markers.append(tag)
        return markers
