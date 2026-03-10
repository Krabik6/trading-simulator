# Runbook: Platform Target Down

## Triggered Alerts

- `PrometheusTargetDown`
- `PrometheusConfigReloadFailed`
- `PrometheusRuleEvaluationSlow`
- `PrometheusScrapePoolFailures`

## Immediate Actions

1. Identify affected target from alert labels (`job`, `instance`).
2. Check container status and restart count in compose environment.
3. Verify service-level `/metrics` endpoint reachability from Prometheus container.

## Diagnostics

1. Target down:
- Validate network path and DNS name inside Docker network.
- Check service logs for startup failures or panic loops.

2. Config reload failed:
- Validate Prometheus config and rule syntax with `promtool`.
- Check mounted file paths in compose manifests.

3. Slow rule evaluation:
- Identify expensive rule group and reduce expression complexity.
- Increase evaluation interval if needed.

## Mitigation

1. Restart failed target and confirm `up == 1`.
2. Revert last Prometheus/rules change if failures began after config update.
3. Split heavy rule groups into smaller groups with longer intervals.

## Recovery Criteria

- All monitored targets in `market-data|trading` are `up == 1`.
- Prometheus config reload status returns to successful.
- Rule evaluation duration is below 80% of interval.

## Follow-up

1. Add regression validation (`promtool`) to CI for any missed rule/config issue.
2. Document persistent flaky network/target behavior and ownership.
