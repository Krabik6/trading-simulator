# Scripts

## CI Scripts

- `scripts/ci/check-go-format.sh`: fails if Go code is not gofmt-formatted.
- `scripts/ci/check-go-tests.sh`: runs module-level Go tests (trading integration package excluded).
- `scripts/ci/check-migration-pairs.sh`: verifies migration up/down pair completeness.
- `scripts/ci/check-compose-config.sh`: validates compose manifests for dev/stage/prod.
