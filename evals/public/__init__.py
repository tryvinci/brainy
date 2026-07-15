"""Proveable public evaluation framework for Brainy.

Compatible shape with mem0ai/memory-benchmarks UnifiedResult (schema_version 1.0)
so LOCOMO / LongMemEval runs can be compared apples-to-apples.

Design rules (proveability):
- Pin dataset URL + content SHA when downloaded
- Pin Brainy base URL / commit, judge model, temperature
- Every question emits retrieval + generation + judgment traces
- Record search_latency_ms and optional token counts
- Never invent scores; missing LLM judge is an explicit mode
"""
