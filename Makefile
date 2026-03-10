.PHONY: up down logs ps clean urls \
	up-dev down-dev up-stage down-stage up-prod down-prod \
	kafka-topics kafka-consume kafka-consume-live \
	validate-compose ci-local check-go-format check-go-tests check-migrations

ENV ?= dev
COMPOSE_FILE ?= deploy/docker/docker-compose.$(ENV).yml
ENV_FILE ?= deploy/env/$(ENV).env

KAFKA_SERVICE ?= kafka
KAFKA_BOOTSTRAP ?= localhost:9092

ifeq ($(ENV),stage)
KAFKA_SERVICE := kafka-1
KAFKA_BOOTSTRAP := kafka-1:9092
endif

ifeq ($(ENV),prod)
KAFKA_SERVICE := kafka-1
KAFKA_BOOTSTRAP := kafka-1:9092
endif

ifneq ("$(wildcard $(ENV_FILE))","")
DC := docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE)
else
DC := docker compose -f $(COMPOSE_FILE)
endif

ensure-compose:
	@test -f $(COMPOSE_FILE) || (echo "Compose file not found: $(COMPOSE_FILE)" && exit 1)

# ============ Docker Compose ============

up: ensure-compose
	$(DC) up -d --build

down: ensure-compose
	$(DC) down

ps: ensure-compose
	$(DC) ps

logs: ensure-compose
	$(DC) logs -f

clean: ensure-compose
	$(DC) down -v --remove-orphans

up-dev:
	@$(MAKE) up ENV=dev

down-dev:
	@$(MAKE) down ENV=dev

up-stage:
	@$(MAKE) up ENV=stage

down-stage:
	@$(MAKE) down ENV=stage

up-prod:
	@$(MAKE) up ENV=prod

down-prod:
	@$(MAKE) down ENV=prod

# ============ Kafka ============

kafka-topics: ensure-compose
	$(DC) exec $(KAFKA_SERVICE) /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server $(KAFKA_BOOTSTRAP)

kafka-consume: ensure-compose
	$(DC) exec $(KAFKA_SERVICE) /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server $(KAFKA_BOOTSTRAP) \
		--topic crypto-prices \
		--from-beginning \
		--max-messages 10

kafka-consume-live: ensure-compose
	$(DC) exec $(KAFKA_SERVICE) /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server $(KAFKA_BOOTSTRAP) \
		--topic crypto-prices

# ============ Validation ============

validate-compose:
	./scripts/ci/check-compose-config.sh

check-go-format:
	./scripts/ci/check-go-format.sh

check-go-tests:
	./scripts/ci/check-go-tests.sh

check-migrations:
	./scripts/ci/check-migration-pairs.sh

ci-local: check-go-format check-go-tests check-migrations validate-compose

# ============ Info ============

urls:
	@echo "Environment:  $(ENV)"
	@echo "Compose file: $(COMPOSE_FILE)"
	@echo "Frontend:     http://localhost:3001"
	@echo "Trading API:  http://localhost:8081"
	@echo "Market Data:  http://localhost:8080"
	@echo "Prometheus:   http://localhost:9090"
	@echo "Grafana:      http://localhost:3000"

prometheus-targets:
	@curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
