# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### From repo root (`trading-simulator/`)
```bash
make up                 # start all services
make down               # stop all
make ps                 # container status
make logs               # all logs
make clean              # stop + remove volumes
make kafka-topics       # list Kafka topics
make kafka-consume      # read 10 messages from crypto-prices
make kafka-consume-live # stream messages (Ctrl+C to exit)
```

### From service dir (`services/trading/`)
```bash
make run          # run locally
make build        # build binary to bin/
make tidy         # go mod tidy
make logs         # service logs
make restart      # restart container
make health       # check /health
make ready        # check /ready
make metrics      # show Prometheus metrics
make migrate-up   # run migrations
make migrate-down # rollback migrations
```

## Architecture

Go microservice for crypto trading simulation with margin trading support.

```
cmd/main.go → config.Load() → logger.Init() → app.New()
                                               ├─ postgres.NewDB()
                                               ├─ kafka.NewPriceConsumer()
                                               ├─ kafka.NewTradeProducer()
                                               ├─ engine.NewEngine()
                                               └─ app.Run()
                                                    ├─ HTTP server (Chi router)
                                                    ├─ PriceConsumer.Start() → prices channel
                                                    └─ PriceProcessor.Start() → update PnL, TP/SL, liquidations
```

## Package Structure

- `cmd/` - entry point
- `config/` - env-based configuration
- `internal/app/` - application orchestration
- `internal/domain/` - entities and repository interfaces
- `internal/repository/postgres/` - PostgreSQL implementations
- `internal/usecase/` - business logic (auth, order, position, account, price)
- `internal/delivery/http/` - REST API handlers and router
- `internal/engine/` - trading calculations (margin, PnL, liquidation)
- `internal/auth/` - JWT and password utilities
- `internal/kafka/` - Kafka consumer and producer
- `internal/logger/` - structured logging (slog)
- `internal/metrics/` - Prometheus metrics
- `migrations/` - SQL migrations

## Configuration

Env vars:

| Variable | Default | Description |
|----------|---------|-------------|
| SERVICE_NAME | trading | service name |
| HTTP_PORT | 8081 | HTTP port |
| LOG_LEVEL | info | log level |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | trading | PostgreSQL user |
| DB_PASSWORD | trading | PostgreSQL password |
| DB_NAME | trading | PostgreSQL database |
| KAFKA_BROKERS | localhost:9092 | Kafka brokers |
| KAFKA_PRICES_TOPIC | crypto-prices | prices topic (consumer) |
| KAFKA_TRADES_TOPIC | trades | trades topic (producer) |
| JWT_SECRET | change-me | JWT signing secret |
| JWT_EXPIRY_HOURS | 24 | JWT token expiry |
| MAX_LEVERAGE | 100 | maximum leverage |
| INITIAL_BALANCE | 10000 | initial balance for new users (USDT) |
| SUPPORTED_SYMBOLS | BTCUSDT,ETHUSDT,SOLUSDT | supported trading pairs |

## REST API

### Health Endpoints
- `GET /health` - liveness
- `GET /ready` - readiness (checks DB, Kafka)
- `GET /metrics` - Prometheus metrics

### Public Endpoints
- `POST /auth/register` - register new user
- `POST /auth/login` - login, returns JWT
- `GET /prices` - current bid/ask/mid for all symbols
- `GET /symbols` - supported trading pairs with specs
- `GET /ws` - WebSocket (optional `?token=<jwt>` for authenticated updates)

### Protected Endpoints (require Bearer token)
- `POST /auth/refresh` - refresh JWT token
- `GET /user/me` - current user profile
- `GET /account` - account balance, equity, margin info
- `POST /orders` - place order
- `GET /orders` - list orders (query: limit, offset)
- `GET /orders/{id}` - get order by ID
- `DELETE /orders/{id}` - cancel pending order
- `GET /positions` - list open positions
- `GET /positions/{id}` - get position by ID
- `POST /positions/{id}/close` - close position
- `PATCH /positions/{id}` - update TP/SL
- `GET /trades` - trade history (query: limit, offset)

## Kafka

**Consumer:** `crypto-prices` (price updates from market-data)
**Producer:** `trades` (trade events for analytics)

## Trading Features

- Long/Short positions with up to 100x leverage
- Market and Limit orders
- One position per symbol per user (averaging when adding)
- Cross margin mode
- Automatic liquidation at calculated price
- Stop Loss / Take Profit
- Real-time PnL updates from price stream

## Key Formulas

```
Long PnL:  Quantity * (MarkPrice - EntryPrice)
Short PnL: Quantity * (EntryPrice - MarkPrice)

InitialMargin = (Quantity * EntryPrice) / Leverage

Long Liquidation:  EntryPrice * (1 - 1/Leverage + 0.005)
Short Liquidation: EntryPrice * (1 + 1/Leverage - 0.005)

Equity = Balance + sum(UnrealizedPnL)
AvailableMargin = Equity - UsedMargin
```

## Database

PostgreSQL with tables: `users`, `accounts`, `orders`, `positions`, `trades`

Key constraint - only one open position per user per symbol:
```sql
CREATE UNIQUE INDEX idx_positions_unique_open
    ON positions(user_id, symbol)
    WHERE status = 'OPEN';
```
