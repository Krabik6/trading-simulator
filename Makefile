.PHONY: up down logs ps clean urls

COMPOSE_FILE := deploy/docker/docker-compose.yml
DC := docker compose -f $(COMPOSE_FILE)

# ============ Docker Compose ============

up:
	$(DC) up -d --build

down:
	$(DC) down

ps:
	$(DC) ps

logs:
	$(DC) logs -f

clean:
	$(DC) down -v --remove-orphans

# ============ Kafka ============

kafka-topics:
	$(DC) exec kafka /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092

kafka-consume:
	$(DC) exec kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic crypto-prices \
		--from-beginning \
		--max-messages 10

kafka-consume-live:
	$(DC) exec kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic crypto-prices

# ============ Info ============

urls:
	@echo "Frontend:     http://localhost:3001"
	@echo "Trading API:  http://localhost:8081"
	@echo "Market Data:  http://localhost:8080"
	@echo "Prometheus:   http://localhost:9090"
	@echo "Grafana:      http://localhost:3000 (admin/admin)"

prometheus-targets:
	@curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
