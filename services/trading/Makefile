.PHONY: run build tidy clean logs health metrics restart docker-build migrate-up migrate-down test test-integration

# ============ Development ============

run:
	go run ./cmd

build:
	go build -o bin/trading ./cmd

tidy:
	go mod tidy

clean:
	rm -rf bin/

# ============ Testing ============

test:
	go test -v ./...

test-integration:
	go test -v -timeout 5m ./internal/integration_test/...

# ============ Docker ============

docker-build:
	docker build -t trading:latest .

logs:
	docker compose -f ../../deploy/docker/docker-compose.yml logs -f trading

restart:
	docker compose -f ../../deploy/docker/docker-compose.yml restart trading

# ============ Database ============

migrate-up:
	migrate -path migrations -database "postgres://trading:trading@localhost:5432/trading?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://trading:trading@localhost:5432/trading?sslmode=disable" down

# ============ Health & Metrics ============

health:
	@curl -s http://localhost:8081/health | jq .

ready:
	@curl -s http://localhost:8081/ready | jq .

metrics:
	@curl -s http://localhost:8081/metrics | grep trading
