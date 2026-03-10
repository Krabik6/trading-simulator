# Stage 0: Risk Register

Date: 2026-03-10
Status: Active

## Risk Table

| ID | Risk | Probability | Impact | Early Signal | Mitigation Plan | Owner Role | Status |
|---|---|---|---|---|---|---|---|
| R-01 | Dual-write inconsistency between DB and Kafka | High | Critical | Missing downstream state after successful command | Transactional outbox and relay with monitoring | Backend Lead | Open |
| R-02 | Event schema drift breaks consumers | Medium | High | Consumer deserialization failures | Schema versioning + compatibility checks in CI | Platform Lead | Open |
| R-03 | Message loss under consumer backpressure | Medium | Critical | Lag spikes + dropped counters | Bounded queues, no commit-before-accept policy, DLQ | Backend Lead | Open |
| R-04 | Command path latency regression after refactor | Medium | High | p95 over SLO | Isolate write path, perf tests per release | Backend Lead | Open |
| R-05 | Replay nondeterminism invalidates bot testing | Medium | Critical | Same input gives different result | Deterministic clocking and version-pinned rules | Simulation Lead | Open |
| R-06 | WS fan-out bottlenecks at high connection count | High | High | Message queue growth, drop rate | Dedicated ws-gateway + broker-based fan-out | Realtime Lead | Open |
| R-07 | Postgres saturation on write bursts | Medium | High | Pool waits, lock contention | Tx optimization, indexing, partition strategy, pooling | DBA/Platform | Open |
| R-08 | Kafka cluster single-point failure in production-like load | Medium | Critical | Broker unavailability impact | 3-broker topology, replication, ISR policy | Platform Lead | Open |
| R-09 | Security gap from token-in-query and permissive WS origin | Medium | High | Unauthorized connection attempts | Header-based auth, strict origin allowlist | Security Lead | Open |
| R-10 | Underestimated bot workload causes infra under-provisioning | Medium | High | Queue depth and CPU saturation | Capacity review each milestone + load test gates | Program Lead | Open |
| R-11 | Scope creep delays critical path | High | Medium | Milestones slip by >1 sprint | Scope lock and change control board | Program Lead | Open |
| R-12 | Insufficient observability blocks incident response | Medium | High | Long MTTR and unknown failure modes | OTel + runbooks + alert ownership | SRE Lead | Open |

## Risk Handling Policy

1. All `Critical` risks require mitigation task in active backlog.
2. Risks without owner are treated as blocked decisions.
3. Risk review cadence: weekly in architecture sync.
