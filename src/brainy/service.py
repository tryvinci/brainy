from __future__ import annotations

from brainy.ingestion import IngestionEngine
from brainy.repository import InMemoryRepository


class BrainyService:
    def __init__(self) -> None:
        self.repository = InMemoryRepository()
        self.ingestion = IngestionEngine(self.repository)
