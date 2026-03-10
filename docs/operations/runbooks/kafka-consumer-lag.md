# Runbook: Kafka Consumer Lag (Trading)

## Triggered Alerts

- `TradingKafkaConsumerLagHigh`
- `TradingKafkaConsumerLagCritical`

## Immediate Actions

1. Confirm lag trend in Grafana (`trading_kafka_consumer_lag`, queue utilization).
2. Check `trading` service CPU/memory saturation and restart loops.
3. Validate Kafka cluster health and ISR status for `crypto-prices`.

## Diagnostics

1. App-side bottleneck:
- `trading_kafka_consumer_queue_length` near capacity indicates processing bottleneck.
- Check DB wait metrics and slow command latency overlap.

2. Broker-side bottleneck:
- High produce latency or broker instability can increase lag.
- Confirm partition leadership and replication health.

## Mitigation

1. Scale `trading` consumers horizontally.
2. Temporarily reduce downstream processing load if non-critical paths exist.
3. Tune consumer throughput settings after incident (batching/parallelism).

## Recovery Criteria

- Lag drops below warning threshold and keeps decreasing.
- Queue utilization stays below 70% for 30 minutes.

## Follow-up

1. Capture lag growth rate and time-to-recovery.
2. Update capacity model if sustained lag happens under expected load.
