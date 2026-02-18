from brainy.service import BrainyService
from brainy.types import ContextQuery, InputType, ReflectionJob


def test_service_contract_end_to_end() -> None:
    service = BrainyService()

    service.ingest(
        tenant_id='tenant-a',
        actor_id='user-1',
        input_type=InputType.MESSAGE,
        content='Creator messaging is the strongest channel for this launch.',
        source='brief',
        source_type='doc',
    )

    artifact_count = service.consolidate('tenant-a')
    assert artifact_count > 0

    bundle = service.retrieve(
        ContextQuery(
            tenant_id='tenant-a',
            actor_id='user-1',
            question='What is the strongest channel?',
            tags=['taste'],
            limit=5,
        )
    )
    assert bundle.artifacts
    assert bundle.beliefs

    ranked = service.rank_hypotheses(
        ContextQuery(
            tenant_id='tenant-a',
            actor_id='user-1',
            question='Which channel should we prioritize?',
            limit=3,
        )
    )
    assert ranked

    belief_id = bundle.beliefs[0].belief_id
    service.record_outcome(
        tenant_id='tenant-a',
        belief_id=belief_id,
        expected=0.8,
        observed=0.4,
    )
    service.record_outcome(
        tenant_id='tenant-a',
        belief_id=belief_id,
        expected=0.8,
        observed=0.3,
    )
    reflection = service.run_reflection(
        ReflectionJob(
            job_id='job-1',
            tenant_id='tenant-a',
            scope='belief-updates',
            cadence='event-triggered',
        )
    )
    assert reflection['updated'] >= 1

    explanation = service.explain(belief_id)
    assert explanation['belief_id'] == belief_id
