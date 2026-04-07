from __future__ import annotations

from dataclasses import dataclass
from uuid import uuid4

from brainy.repository import InMemoryRepository
from brainy.types import BeliefStatus, OutcomeEvent, ReflectionJob


@dataclass(slots=True)
class StopLossPolicy:
    min_observations: int = 1
    delta_threshold: float = 0.2
    persistence_window: int = 2
    max_conviction_drop_per_cycle: float = 0.25


class ReflectionEngine:
    def __init__(self, repository: InMemoryRepository, policy: StopLossPolicy | None = None) -> None:
        self.repository = repository
        self.policy = policy or StopLossPolicy()

    def record_outcome(
        self,
        *,
        tenant_id: str,
        belief_id: str,
        expected: float,
        observed: float,
        context: dict[str, object] | None = None,
    ) -> OutcomeEvent:
        outcome = OutcomeEvent(
            outcome_id=f'out_{uuid4().hex}',
            tenant_id=tenant_id,
            belief_id=belief_id,
            expected=expected,
            observed=observed,
            context=context or {},
        )
        self.repository.add_outcome(outcome)
        self.repository.add_audit_event('record_outcome', {'outcome_id': outcome.outcome_id, 'belief_id': belief_id})
        return outcome

    def run_reflection(self, job: ReflectionJob) -> dict[str, int]:
        outcomes = [
            outcome
            for outcome in self.repository.list_outcomes(job.tenant_id)
            if self.repository.beliefs.get(outcome.belief_id) is not None
        ]

        updated = 0
        challenged = 0
        for belief in self.repository.list_beliefs(job.tenant_id):
            related = [outcome for outcome in outcomes if outcome.belief_id == belief.belief_id]
            if len(related) < self.policy.min_observations:
                continue

            failing = [outcome for outcome in related if (outcome.expected - outcome.observed) >= self.policy.delta_threshold]
            if not failing:
                continue

            avg_delta = sum((outcome.expected - outcome.observed) for outcome in failing) / len(failing)
            conviction_drop = min(self.policy.max_conviction_drop_per_cycle, avg_delta)
            belief.conviction = max(0.0, round(belief.conviction - conviction_drop, 3))
            belief.rank = max(0.0, round(belief.rank - conviction_drop, 3))
            updated += 1

            if len(failing) >= self.policy.persistence_window:
                belief.status = BeliefStatus.CHALLENGED
                challenged += 1

        self.repository.add_audit_event(
            'run_reflection',
            {'tenant_id': job.tenant_id, 'updated': updated, 'challenged': challenged, 'job_id': job.job_id},
        )
        return {'updated': updated, 'challenged': challenged}
