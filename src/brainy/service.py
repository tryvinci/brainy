from __future__ import annotations

from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.governance import GovernanceEngine
from brainy.ingestion import IngestionEngine
from brainy.reflection import ReflectionEngine
from brainy.repository import InMemoryRepository
from brainy.retrieval import RetrievalEngine


class BrainyService:
    def __init__(self) -> None:
        self.repository = InMemoryRepository()
        self.ingestion = IngestionEngine(self.repository)
        self.consolidation = ConsolidationEngine(self.repository)
        self.belief_graph = BeliefGraphEngine(self.repository)
        self.retrieval = RetrievalEngine(self.repository)
        self.reflection = ReflectionEngine(self.repository)
        self.governance = GovernanceEngine(self.repository)
