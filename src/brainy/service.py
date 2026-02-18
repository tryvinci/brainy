from __future__ import annotations

from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.governance import GovernanceEngine
from brainy.ingestion import IngestionEngine
from brainy.reflection import ReflectionEngine
from brainy.repository import InMemoryRepository
from brainy.retrieval import RetrievalEngine
from brainy.types import ContextBundle, ContextQuery, InputType, MemoryIngestEvent, OutcomeEvent, ReflectionJob


class BrainyService:
    def __init__(self) -> None:
        self.repository = InMemoryRepository()
        self.ingestion = IngestionEngine(self.repository)
        self.consolidation = ConsolidationEngine(self.repository)
        self.belief_graph = BeliefGraphEngine(self.repository)
        self.retrieval = RetrievalEngine(self.repository)
        self.reflection = ReflectionEngine(self.repository)
        self.governance = GovernanceEngine(self.repository)

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
        return self.ingestion.ingest(
            tenant_id=tenant_id,
            actor_id=actor_id,
            input_type=input_type,
            content=content,
            source=source,
            source_type=source_type,
            tags=tags,
            metadata=metadata,
            model_version=model_version,
        )

    def consolidate(self, scope: str) -> int:
        artifacts = self.consolidation.consolidate(scope)
        self.belief_graph.build(scope)
        return len(artifacts)

    def retrieve(self, query: ContextQuery) -> ContextBundle:
        return self.retrieval.retrieve(query)

    def rank_hypotheses(self, context: ContextQuery) -> list[tuple[str, float]]:
        bundle = self.retrieve(context)
        ranked = sorted(
            [(belief.claim, belief.rank) for belief in bundle.beliefs],
            key=lambda item: item[1],
            reverse=True,
        )
        return ranked

    def record_outcome(
        self,
        *,
        tenant_id: str,
        belief_id: str,
        expected: float,
        observed: float,
        context: dict[str, object] | None = None,
    ) -> OutcomeEvent:
        return self.reflection.record_outcome(
            tenant_id=tenant_id,
            belief_id=belief_id,
            expected=expected,
            observed=observed,
            context=context,
        )

    def run_reflection(self, job: ReflectionJob) -> dict[str, int]:
        return self.reflection.run_reflection(job)

    def explain(self, decision_id: str) -> dict[str, object]:
        return self.governance.explain_decision(decision_id)
