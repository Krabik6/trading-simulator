# Stage 0 Design: Architecture Foundation And Program Control

Date: 2026-03-10
Status: Accepted

## Purpose

Stage 0 establishes shared technical baselines before execution work starts. The goal is to remove ambiguity around load targets, reliability targets, service ownership, and program risks so later implementation stages can move quickly without architecture churn.

This stage does not change runtime behavior. It defines the operating contract for all next stages.

## Scope

Stage 0 delivers:
- Target load model and capacity assumptions.
- SLO, SLI, and error budget policy.
- Service boundaries, ownership, and data flow contracts.
- ADR package for core architectural decisions.
- Risk register and critical path with dependency order.
- Execution roadmap at epic level.

## Artifacts

- `docs/architecture/stage-0/01-target-load-model.md`
- `docs/architecture/stage-0/02-slo-sli-and-error-budget.md`
- `docs/architecture/stage-0/03-service-boundaries.md`
- `docs/architecture/adr/ADR-0001-event-driven-core.md`
- `docs/architecture/adr/ADR-0002-transactional-outbox.md`
- `docs/architecture/adr/ADR-0003-idempotency-and-deduplication.md`
- `docs/architecture/adr/ADR-0004-read-write-separation.md`
- `docs/architecture/stage-0/04-risk-register.md`
- `docs/architecture/stage-0/05-critical-path.md`
- `docs/architecture/stage-0/06-epic-roadmap.md`

## Working Assumptions

- One codebase serves two product tracks: live trading simulation and bot training/testing.
- Trading correctness takes priority over throughput.
- Trading events are never lost; duplicates are tolerated and deduplicated.
- Replay determinism is mandatory for bot validation.

## Exit Criteria

Stage 0 is complete when:
1. Load and SLO targets are explicitly documented and traceable to metrics.
2. Service ownership and data boundaries are documented.
3. Core architectural decisions are captured in accepted ADRs.
4. Risks and mitigation owners are assigned.
5. Critical path and epic sequencing are documented.

## Out Of Scope

- Runtime refactors.
- New production services.
- Schema migrations.
- Performance tuning implementation.
