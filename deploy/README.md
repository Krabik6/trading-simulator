# Deployment Foundation (Stage 1)

This directory contains deployment assets for three environments:
- `dev`
- `stage`
- `prod`

## Compose Manifests

- `deploy/docker/docker-compose.dev.yml`
- `deploy/docker/docker-compose.stage.yml`
- `deploy/docker/docker-compose.prod.yml`

`deploy/docker/docker-compose.yml` is kept as a compatibility alias for `dev`.

## Environment Profiles

- `deploy/env/dev.env`
- `deploy/env/stage.env`
- `deploy/env/prod.env.example`

For production usage, create `deploy/env/prod.env` from the example and inject sensitive values from your secret manager.

## Kafka Bootstrap

Kafka topics are explicitly created by `kafka-init` service using:
- `deploy/docker/kafka/create-topics.sh`
- `deploy/docker/kafka/topic-specs.env`

Auto topic creation is disabled in all manifests.

## PgBouncer

Stage and prod manifests route DB traffic through PgBouncer:
- `deploy/docker/pgbouncer/pgbouncer.ini`
- `deploy/docker/pgbouncer/userlist.txt`

## Usage

From repository root:

```bash
make up ENV=dev
make up ENV=stage
make up ENV=prod

make down ENV=dev
make down ENV=stage
make down ENV=prod
```

Validation:

```bash
make validate-compose
```
