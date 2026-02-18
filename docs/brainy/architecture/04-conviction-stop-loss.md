# Conviction vs Outcome Stop-Loss Policy

## Goal
Formalize when conviction must be reduced because outcomes do not validate the belief.

## Policy Parameters
- `min_observations`: minimum outcome samples before hard adjustments.
- `delta_threshold`: expected-observed gap threshold.
- `persistence_window`: number of consecutive failures before escalation.
- `max_conviction_drop_per_cycle`: guardrail to avoid overreaction.

## Actions
- Soft trigger: reduce rank and conviction incrementally.
- Hard trigger: set belief to challenged and schedule experiment.
- Terminal trigger: retire belief after unresolved failure window.

## Configurability
Policy is per domain and per belief class.
