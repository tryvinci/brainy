from __future__ import annotations

from brainy.repository import InMemoryRepository
from brainy.types import ContextBundle, ContextQuery


class RetrievalEngine:
    def __init__(self, repository: InMemoryRepository) -> None:
        self.repository = repository

    def retrieve(self, query: ContextQuery) -> ContextBundle:
        artifacts = self.repository.list_artifacts(query.tenant_id)
        beliefs = self.repository.list_beliefs(query.tenant_id)

        ranked_artifacts = sorted(
            artifacts,
            key=lambda artifact: self._artifact_score(artifact.content, query.question, artifact.tags, query.tags),
            reverse=True,
        )[: query.limit]

        ranked_beliefs = sorted(
            beliefs,
            key=lambda belief: self._belief_score(belief.claim, query.question, belief.rank, query.tags),
            reverse=True,
        )[: query.limit]

        explanation = (
            f"Context assembled with {len(ranked_artifacts)} artifacts and {len(ranked_beliefs)} beliefs "
            f"for actor {query.actor_id}."
        )
        self.repository.add_audit_event(
            'retrieve',
            {'tenant_id': query.tenant_id, 'actor_id': query.actor_id, 'query': query.question},
        )

        return ContextBundle(
            tenant_id=query.tenant_id,
            query=query.question,
            artifacts=ranked_artifacts,
            beliefs=ranked_beliefs,
            explanation=explanation,
        )

    @staticmethod
    def _artifact_score(content: str, question: str, tags: list[str], query_tags: list[str]) -> float:
        content_tokens = set(content.lower().split())
        question_tokens = set(question.lower().split())
        overlap = len(content_tokens & question_tokens)
        taste_bonus = 0.8 if any('taste' in tag for tag in tags + query_tags) else 0.0
        return overlap + taste_bonus

    @staticmethod
    def _belief_score(claim: str, question: str, rank: float, query_tags: list[str]) -> float:
        claim_tokens = set(claim.lower().split())
        question_tokens = set(question.lower().split())
        overlap = len(claim_tokens & question_tokens)
        identity_bonus = 0.5 if any('identity' in tag for tag in query_tags) else 0.0
        return rank + overlap + identity_bonus
