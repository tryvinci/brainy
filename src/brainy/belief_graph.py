from __future__ import annotations

from uuid import uuid4

from brainy.repository import InMemoryRepository
from brainy.types import ArtifactType, BeliefNode, BeliefStatus


class BeliefGraphEngine:
    def __init__(self, repository: InMemoryRepository) -> None:
        self.repository = repository

    def build(self, tenant_id: str) -> list[BeliefNode]:
        existing_claims = {belief.claim for belief in self.repository.list_beliefs(tenant_id)}
        artifacts = self.repository.list_artifacts(tenant_id)
        created: list[BeliefNode] = []
        for artifact in artifacts:
            if artifact.artifact_type not in {ArtifactType.FACT, ArtifactType.PRINCIPLE, ArtifactType.PATTERN}:
                continue
            claim = artifact.content
            if claim in existing_claims:
                continue
            belief = BeliefNode(
                belief_id=f'blf_{uuid4().hex}',
                tenant_id=tenant_id,
                claim=claim,
                rank=self._rank_from_artifact(artifact),
                conviction=artifact.confidence,
                status=BeliefStatus.ACTIVE,
                evidence_artifact_ids=[artifact.artifact_id],
                supporting_event_ids=artifact.evidence_event_ids,
            )
            self.repository.add_belief(belief)
            created.append(belief)
            existing_claims.add(claim)

        self._link_conflicts(tenant_id)
        if created:
            self.repository.add_audit_event('build_beliefs', {'tenant_id': tenant_id, 'count': len(created)})
        return created

    def reconcile_conflicts(self, tenant_id: str) -> list[tuple[str, str]]:
        beliefs = self.repository.list_beliefs(tenant_id)
        experiments: list[tuple[str, str]] = []
        for belief in beliefs:
            for conflict_id in belief.conflicting_belief_ids:
                conflict = self.repository.beliefs[conflict_id]
                if self._can_complement(belief.claim, conflict.claim):
                    continue
                belief.status = BeliefStatus.CHALLENGED
                conflict.status = BeliefStatus.CHALLENGED
                experiments.append((belief.belief_id, conflict.belief_id))
        if experiments:
            self.repository.add_audit_event('reconcile_conflicts', {'tenant_id': tenant_id, 'pairs': experiments})
        return experiments

    @staticmethod
    def _rank_from_artifact(artifact: object) -> float:
        confidence = getattr(artifact, 'confidence', 0.5)
        tags = getattr(artifact, 'tags', [])
        identity_bonus = 0.1 if any('identity' in tag for tag in tags) else 0.0
        principle_bonus = 0.1 if any('principle' in tag for tag in tags) else 0.0
        return round(min(1.0, confidence + identity_bonus + principle_bonus), 3)

    def _link_conflicts(self, tenant_id: str) -> None:
        beliefs = self.repository.list_beliefs(tenant_id)
        for left in beliefs:
            left.conflicting_belief_ids = []
        for index, left in enumerate(beliefs):
            for right in beliefs[index + 1 :]:
                if self._is_conflict(left.claim, right.claim):
                    left.conflicting_belief_ids.append(right.belief_id)
                    right.conflicting_belief_ids.append(left.belief_id)

    @staticmethod
    def _is_conflict(left: str, right: str) -> bool:
        left_l = left.lower()
        right_l = right.lower()
        if left_l == right_l:
            return False
        neg_tokens = [' not ', ' never ', ' no ']
        left_neg = any(token in f' {left_l} ' for token in neg_tokens)
        right_neg = any(token in f' {right_l} ' for token in neg_tokens)
        overlap = set(left_l.split()) & set(right_l.split())
        return bool(overlap) and left_neg != right_neg

    @staticmethod
    def _can_complement(left: str, right: str) -> bool:
        neg_tokens = [' not ', ' never ', ' no ']
        left_l = f' {left.lower()} '
        right_l = f' {right.lower()} '
        if any(token in left_l for token in neg_tokens) != any(token in right_l for token in neg_tokens):
            return False

        contextual_tokens = {'when', 'for', 'with', 'in'}
        left_tokens = set(left.lower().split())
        right_tokens = set(right.lower().split())
        return bool((left_tokens & contextual_tokens) and (right_tokens & contextual_tokens))
