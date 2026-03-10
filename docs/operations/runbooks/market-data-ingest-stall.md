# Runbook: Market-Data Ingest Stall

## Triggered Alerts

- `MarketDataIngestStalled`
- `MarketDataKafkaSendErrors`
- `MarketDataProducerBacklogHigh`
- `MarketDataSourceDisconnected`
- `MarketDataHTTP5xxRateHigh`

## Immediate Actions (first 10 minutes)

1. Verify source connection metric `market_data_client_connected`.
2. Inspect `market-data` logs for upstream WebSocket/client errors.
3. Check Kafka broker health and topic availability (`crypto-prices`).
4. Inspect producer backlog metric `market_data_kafka_producer_in_flight`.

## Diagnostics

1. Ingest stalled with source disconnected:
- Validate external source connectivity.
- Restart market-data service if reconnect loop is stuck.

2. Kafka errors or backlog growth:
- Check broker availability and network latency.
- Verify Kafka topic exists and partitions are online.

3. Health endpoint errors:
- Inspect `/ready` failures and producer health checks.

## Mitigation

1. Restart `market-data` if client cannot recover connection.
2. Switch to backup data source if available.
3. Reduce symbol set temporarily to lower throughput pressure.
4. Coordinate with Kafka platform owner if broker issues persist.

## Recovery Criteria

- `market_data_prices_processed_total` rate restored to baseline.
- `market_data_kafka_producer_in_flight` returns to steady low value.
- No Kafka send error growth for 15 minutes.

## Follow-up

1. Record outage window and data gaps.
2. Backfill missed interval data if required by downstream consumers.
3. Tune reconnect and producer retry settings if needed.
