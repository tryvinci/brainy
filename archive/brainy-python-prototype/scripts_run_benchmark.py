from __future__ import annotations

import json
from pathlib import Path
import sys
from dataclasses import asdict, is_dataclass

sys.path.insert(0, str(Path(__file__).parent / 'src'))

from brainy.benchmark import BenchmarkRunner


if __name__ == '__main__':
    runner = BenchmarkRunner()
    result = runner.run_all()

    output_path = Path('docs/brainy/benchmarks/latest-report.json')
    output_path.parent.mkdir(parents=True, exist_ok=True)
    def serializer(obj: object) -> object:
        if is_dataclass(obj):
            return asdict(obj)
        return str(obj)

    output_path.write_text(json.dumps(result, indent=2, default=serializer), encoding='utf-8')

    markdown_path = Path('docs/brainy/benchmarks/latest-report.md')
    lines = [
        '# Brainy Benchmark Report',
        '',
        '## Public Track',
    ]
    for row in result['public_track']:
        lines.append(f"- {row.name}: score={row.score}, latency_ms={round(row.latency_ms, 3)}, details={row.details}")

    lines += ['', '## Cognitive Track']
    for row in result['cognitive_track']:
        lines.append(f"- {row.name}: score={row.score}, latency_ms={round(row.latency_ms, 3)}, details={row.details}")

    lines += [
        '',
        '## Summaries',
        f"- public: {result['public_summary']}",
        f"- cognitive: {result['cognitive_summary']}",
        '',
        '## Caveat',
        '- Competitor adapter calls are skipped unless API keys are configured.',
    ]
    markdown_path.write_text('\n'.join(lines) + '\n', encoding='utf-8')
