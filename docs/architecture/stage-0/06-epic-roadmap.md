# Stage 0: Epic Roadmap And Sequencing

Date: 2026-03-10
Status: Baseline

## Epic Sequence

1. `EPIC-PLATFORM-FOUNDATION`
- Production-like Kafka/Postgres topology, CI/CD guardrails, secret management.

2. `EPIC-OBSERVABILITY-AND-OPERATIONS`
- OTel instrumentation, SLI coverage, dashboards, alerting, runbooks.

3. `EPIC-CORE-CORRECTNESS`
- Transaction boundaries, concurrency control, idempotency keys, invariants.

4. `EPIC-EVENT-RELIABILITY`
- Outbox relay, schema versioning, DLQ strategy, consumer dedup.

5. `EPIC-SERVICE-BOUNDARY-REFACTOR`
- Split into trading-core, query-api, ws-gateway with clear ownership.

6. `EPIC-MARKET-DATA-HARDENING`
- Backpressure, bounded concurrency, multi-source readiness.

7. `EPIC-HISTORICAL-AND-REPLAY`
- Historical ingest, replay APIs, determinism checks.

8. `EPIC-BOT-SIMULATION`
- Session lifecycle, backtest/paper modes, result storage.

9. `EPIC-AI-RUNTIME`
- Model integration contracts, policy guards, drift monitoring.

10. `EPIC-PRODUCT-SURFACE`
- API contracts, UI workflows, access controls for simulation.

11. `EPIC-RESILIENCE-VALIDATION`
- Load, chaos, failover, DR, security hardening.

12. `EPIC-ROLLOUT-AND-STABILIZATION`
- Canary rollout, progressive traffic, post-launch control.

## Planning Notes

- Epics 3-5 form the mandatory path for scale safety.
- Epics 7-9 form the mandatory path for bot platform credibility.
- Each epic must define measurable done criteria tied to SLO/SLI.
