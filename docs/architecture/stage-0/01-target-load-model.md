# Stage 0: Target Load Model

Date: 2026-03-10
Status: Baseline for planning

## Goal

Define load targets for both product tracks:
- Live trading simulation.
- Bot training/testing and replay workloads.

All values are planning targets used for sizing, architecture decisions, and test scenarios.

## Load Profiles

### Profile A: Live Trading (Interactive)

- High fan-out real-time market updates.
- Short command latency required for order operations.
- Burst behavior around volatile market periods.

### Profile B: Bot Training/Testing (Batch + Stream)

- High sustained event consumption.
- Large replay workloads with deterministic outputs.
- Lower interactive latency requirements, higher throughput requirements.

## Capacity Targets

| Metric | Current (Observed Approx) | Target P1 (Scale-Ready) | Target P2 (Bot Platform) |
|---|---:|---:|---:|
| Active users (monthly) | 1,000 | 10,000 | 50,000 |
| Concurrent logged-in users | 150 | 2,000 | 10,000 |
| Concurrent WS connections | 200 | 5,000 | 25,000 |
| API command peak RPS (`POST /orders`, close, TP/SL) | 15 | 150 | 600 |
| API read peak RPS (`/positions`, `/orders`, `/trades`) | 30 | 300 | 1,000 |
| Market tick ingest peak (all symbols) | 300/s | 3,000/s | 10,000/s |
| Open positions in write model | 5,000 | 100,000 | 500,000 |
| Daily trade events produced | 50,000 | 2,000,000 | 10,000,000 |
| Concurrent simulation runs | 0 | 100 | 1,000 |
| Replay event rate per simulation | 0 | 500/s | 2,000/s |

## Non-Functional Capacity Constraints

- Trading event loss budget: zero tolerated for `trading_events`.
- Tick processing may be at-least-once with deterministic dedup handling.
- No unbounded goroutine growth in hot paths.
- DB write path must remain within SLO under P1 peak load.

## Load Test Scenarios To Cover In Later Stages

1. Live burst: 10x command spike for 10 minutes.
2. Tick storm: sustained 10,000 ticks/s for 30 minutes.
3. Mixed mode: 70% live load + 30% replay load.
4. Long-run stability: 24h soak with background bot simulations.
