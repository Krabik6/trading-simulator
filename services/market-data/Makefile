.PHONY: run build tidy clean logs health metrics restart docker-build

# ============ Development ============

run:
	go run ./cmd

build:
	go build -o bin/market-data ./cmd

tidy:
	go mod tidy

clean:
	rm -rf bin/

# ============ Docker ============

docker-build:
	docker build -t market-data:latest .

logs:
	docker compose -f ../../deploy/docker/docker-compose.yml logs -f market-data

restart:
	docker compose -f ../../deploy/docker/docker-compose.yml restart market-data

# ============ Health & Metrics ============

health:
	@curl -s http://localhost:8080/health | jq .

metrics:
	@curl -s http://localhost:8080/metrics | grep market_data
