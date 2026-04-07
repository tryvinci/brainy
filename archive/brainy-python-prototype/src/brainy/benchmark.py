from __future__ import annotations

import os
from dataclasses import dataclass
from statistics import mean
from time import perf_counter

from brainy.service import BrainyService
from brainy.types import ContextQuery, InputType, ReflectionJob


@dataclass(slots=True)
class BenchmarkResult:
    name: str
    score: float
    latency_ms: float
    details: dict[str, object]


class CompetitorAdapter:
    name = 'unknown'

    def run(self) -> BenchmarkResult:
        raise NotImplementedError


class APIKeyCompetitorAdapter(CompetitorAdapter):
    env_var = ''

    def run(self) -> BenchmarkResult:
        key = os.getenv(self.env_var)
        if not key:
            return BenchmarkResult(
                name=self.name,
                score=0.0,
                latency_ms=0.0,
                details={'status': 'skipped', 'reason': f'missing {self.env_var}'},
            )
        return BenchmarkResult(
            name=self.name,
            score=0.0,
            latency_ms=0.0,
            details={'status': 'not_implemented', 'reason': 'adapter stub'},
        )


class Mem0Adapter(APIKeyCompetitorAdapter):
    name = 'mem0'
    env_var = 'MEM0_API_KEY'


class SupermemoryAdapter(APIKeyCompetitorAdapter):
    name = 'supermemory'
    env_var = 'SUPERMEMORY_API_KEY'


class ZepAdapter(APIKeyCompetitorAdapter):
    name = 'zep'
    env_var = 'ZEP_API_KEY'


class LettaAdapter(APIKeyCompetitorAdapter):
    name = 'letta'
    env_var = 'LETTA_API_KEY'


class MemobaseAdapter(APIKeyCompetitorAdapter):
    name = 'memobase'
    env_var = 'MEMOBASE_API_KEY'


class CogneeAdapter(APIKeyCompetitorAdapter):
    name = 'cognee'
    env_var = 'COGNEE_API_KEY'


class BenchmarkRunner:
    def __init__(self) -> None:
        self.service = BrainyService()
        self.competitors: list[CompetitorAdapter] = [
            Mem0Adapter(),
            SupermemoryAdapter(),
            ZepAdapter(),
            LettaAdapter(),
            MemobaseAdapter(),
            CogneeAdapter(),
        ]

    def run_public_memory_track(self) -> list[BenchmarkResult]:
        start = perf_counter()

        self.service.ingestion.ingest(
            tenant_id='benchmark',
            actor_id='system',
            input_type=InputType.MESSAGE,
            content='Creator-led messaging is highest performing when identity coherence is high.',
            source='benchmark',
            source_type='fixture',
            tags=['taste:identity'],
        )
        self.service.consolidation.consolidate('benchmark')
        self.service.belief_graph.build('benchmark')
        bundle = self.service.retrieval.retrieve(
            ContextQuery(
                tenant_id='benchmark',
                actor_id='system',
                question='What performs best for creators?',
                tags=['taste'],
                limit=5,
            )
        )

        latency_ms = (perf_counter() - start) * 1000
        score = 1.0 if bundle.artifacts and bundle.beliefs else 0.0
        brainy_result = BenchmarkResult(
            name='brainy-public-track',
            score=score,
            latency_ms=latency_ms,
            details={'artifact_count': len(bundle.artifacts), 'belief_count': len(bundle.beliefs)},
        )

        results = [brainy_result]
        for competitor in self.competitors:
            results.append(competitor.run())
        return results

    def run_cognitive_track(self) -> list[BenchmarkResult]:
        start = perf_counter()

        tenant_id = 'cognitive-track'
        event = self.service.ingestion.ingest(
            tenant_id=tenant_id,
            actor_id='system',
            input_type=InputType.MESSAGE,
            content='Premium visual language always wins for this audience.',
            source='benchmark',
            source_type='fixture',
        )
        self.service.consolidation.consolidate(tenant_id)
        beliefs = self.service.belief_graph.build(tenant_id)
        target = beliefs[0]

        self.service.reflection.record_outcome(
            tenant_id=tenant_id,
            belief_id=target.belief_id,
            expected=0.9,
            observed=0.3,
            context={'event_id': event.event_id},
        )
        self.service.reflection.record_outcome(
            tenant_id=tenant_id,
            belief_id=target.belief_id,
            expected=0.85,
            observed=0.35,
            context={'event_id': event.event_id},
        )
        reflection_result = self.service.reflection.run_reflection(
            ReflectionJob(
                job_id='bench-reflect',
                tenant_id=tenant_id,
                scope='belief-updates',
                cadence='event-triggered',
            )
        )

        latency_ms = (perf_counter() - start) * 1000
        score = 1.0 if reflection_result['challenged'] >= 1 else 0.0
        return [
            BenchmarkResult(
                name='brainy-cognitive-track',
                score=score,
                latency_ms=latency_ms,
                details=reflection_result,
            )
        ]

    @staticmethod
    def summarize(results: list[BenchmarkResult]) -> dict[str, object]:
        if not results:
            return {'average_score': 0.0, 'average_latency_ms': 0.0}
        return {
            'average_score': round(mean(result.score for result in results), 3),
            'average_latency_ms': round(mean(result.latency_ms for result in results), 3),
        }

    def run_all(self) -> dict[str, object]:
        public_results = self.run_public_memory_track()
        cognitive_results = self.run_cognitive_track()
        return {
            'public_track': public_results,
            'cognitive_track': cognitive_results,
            'public_summary': self.summarize(public_results),
            'cognitive_summary': self.summarize(cognitive_results),
        }
