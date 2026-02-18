from __future__ import annotations

from brainy.belief_graph import BeliefGraphEngine
from brainy.consolidation import ConsolidationEngine
from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository


class BrainyService:
    def __init__(self) -> None:
        self.repository = InMemoryRepository()
        self.ingestion = IngestionEngine(self.repository)
        self.consolidation = ConsolidationEngine(self.repository)
        self.belief_graph = BeliefGraphEngine(self.repository)
