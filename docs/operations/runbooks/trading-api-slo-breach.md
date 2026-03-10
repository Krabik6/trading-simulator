# Runbook: Trading API SLO Breach

## Triggered Alerts

- `TradingAPIHigh5xxRate`
- `TradingCommandLatencyP95High`
- `TradingCommandLatencyP99Critical`
- `TradingDBPoolSaturationHigh`
- `TradingDBConnectionWaitSpike`
- `TradingWSMessagesDropped`

## Immediate Actions (first 10 minutes)

1. Verify alert scope in Grafana dashboard `Trading Service Overview`.
2. Check `trading` container logs for errors and panics.
3. Confirm DB health and pool state (`trading_db_pool_in_use`, `trading_db_pool_open`, wait metrics).
4. Confirm Kafka consumer lag is not spiking at the same time.

## Diagnostics

1. High 5xx:
- Inspect route-level error spikes in `trading_http_requests_total{status=~"5.."}`.
- Correlate with deploy/restart events.

2. High latency:
- Inspect p95/p99 query against command routes.
- Check DB wait growth and Kafka lag correlation.

3. WS drops:
- Inspect `trading_ws_messages_dropped_total` by `reason`.
- Verify active WS connections and queue utilization.

## Mitigation

1. Roll back latest deploy if errors started after release.
2. Temporarily scale `trading` replicas (if deployment mode supports scaling).
3. Increase DB pool limits only if PostgreSQL/pgbouncer headroom is confirmed.
4. Reduce incoming load for command routes if SLO burn is critical.

## Recovery Criteria

- 5xx ratio < 1% for at least 30 minutes.
- p95 command latency < 150ms and p99 < 400ms for at least 30 minutes.
- DB wait count growth returns to baseline.

## Follow-up

1. Create incident summary with root cause and timeline.
2. Add missing alert threshold tuning if noise observed.
3. Add regression test for root-cause path.
