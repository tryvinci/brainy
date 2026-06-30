from brainy.benchmark import BenchmarkRunner


def test_benchmark_runner_produces_reproducible_structure() -> None:
    runner = BenchmarkRunner()
    result = runner.run_all()

    assert 'public_track' in result
    assert 'cognitive_track' in result
    assert result['public_track']
    assert result['cognitive_track']
    assert 'average_score' in result['public_summary']
    assert 'average_latency_ms' in result['cognitive_summary']
