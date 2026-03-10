package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	OrdersPlaced = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "orders_placed_total",
			Help:      "Total number of orders placed",
		},
		[]string{"symbol", "side", "type"},
	)

	OrdersFilled = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "orders_filled_total",
			Help:      "Total number of orders filled",
		},
		[]string{"symbol", "side"},
	)

	OrdersCancelled = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "orders_cancelled_total",
			Help:      "Total number of orders cancelled",
		},
		[]string{"symbol"},
	)

	PositionsOpened = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "positions_opened_total",
			Help:      "Total number of positions opened",
		},
		[]string{"symbol", "side"},
	)

	PositionsClosed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "positions_closed_total",
			Help:      "Total number of positions closed",
		},
		[]string{"symbol", "side", "reason"},
	)

	Liquidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "liquidations_total",
			Help:      "Total number of liquidations",
		},
		[]string{"symbol"},
	)

	ActivePositions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "active_positions",
			Help:      "Number of active positions",
		},
		[]string{"symbol", "side"},
	)

	TotalPnL = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "total_pnl",
			Help:      "Total PnL by symbol",
		},
		[]string{"symbol"},
	)

	PriceUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "price_updates_total",
			Help:      "Total number of price updates processed",
		},
		[]string{"symbol"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "trading",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"method", "path", "status"},
	)

	KafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "kafka_consumer_lag",
			Help:      "Kafka consumer lag",
		},
		[]string{"topic", "partition"},
	)

	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_connections_open",
			Help:      "Number of open database connections",
		},
	)

	DBConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_connections_in_use",
			Help:      "Number of database connections currently in use",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_connections_idle",
			Help:      "Number of idle database connections",
		},
	)

	DBWaitCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_wait_count",
			Help:      "Total number of waits for a new DB connection",
		},
	)

	DBWaitDurationSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_wait_duration_seconds",
			Help:      "Total time blocked waiting for a DB connection",
		},
	)

	DBPoolOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_pool_open",
			Help:      "Number of open database connections (SLI-compatible naming)",
		},
	)

	DBPoolInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_pool_in_use",
			Help:      "Number of in-use database connections (SLI-compatible naming)",
		},
	)

	DBPoolIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_pool_idle",
			Help:      "Number of idle database connections (SLI-compatible naming)",
		},
	)

	DBPoolWaitCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_pool_wait_count",
			Help:      "Total number of waits for a DB connection (SLI-compatible naming)",
		},
	)

	DBPoolWaitDurationSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "db_pool_wait_duration_seconds",
			Help:      "Total blocked time waiting for DB connection (SLI-compatible naming)",
		},
	)

	KafkaConsumerQueueLength = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "kafka_consumer_queue_length",
			Help:      "Kafka consumer in-memory queue length",
		},
		[]string{"topic", "partition"},
	)

	KafkaConsumerQueueCapacity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "kafka_consumer_queue_capacity",
			Help:      "Kafka consumer in-memory queue capacity",
		},
		[]string{"topic", "partition"},
	)

	WSConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "trading",
			Name:      "ws_connections_active",
			Help:      "Number of active WebSocket connections",
		},
	)

	WSMessageDropped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trading",
			Name:      "ws_messages_dropped_total",
			Help:      "Total number of dropped WebSocket messages",
		},
		[]string{"reason"},
	)
)

func RecordOrderPlaced(symbol, side, orderType string) {
	OrdersPlaced.WithLabelValues(symbol, side, orderType).Inc()
}

func RecordOrderFilled(symbol, side string) {
	OrdersFilled.WithLabelValues(symbol, side).Inc()
}

func RecordOrderCancelled(symbol string) {
	OrdersCancelled.WithLabelValues(symbol).Inc()
}

func RecordPositionOpened(symbol, side string) {
	PositionsOpened.WithLabelValues(symbol, side).Inc()
	ActivePositions.WithLabelValues(symbol, side).Inc()
}

func RecordPositionClosed(symbol, side, reason string) {
	PositionsClosed.WithLabelValues(symbol, side, reason).Inc()
	ActivePositions.WithLabelValues(symbol, side).Dec()
}

func RecordLiquidation(symbol string) {
	Liquidations.WithLabelValues(symbol).Inc()
}

func RecordPriceUpdate(symbol string) {
	PriceUpdates.WithLabelValues(symbol).Inc()
}

func SetTotalPnL(symbol string, pnl float64) {
	TotalPnL.WithLabelValues(symbol).Set(pnl)
}

func ObserveHTTPRequest(method, path, status string, durationSeconds float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, status).Observe(durationSeconds)
}

func SetKafkaConsumerLag(topic, partition string, lag float64) {
	KafkaConsumerLag.WithLabelValues(topic, partition).Set(lag)
}

func SetKafkaConsumerQueue(topic, partition string, length, capacity float64) {
	KafkaConsumerQueueLength.WithLabelValues(topic, partition).Set(length)
	KafkaConsumerQueueCapacity.WithLabelValues(topic, partition).Set(capacity)
}

func SetDBConnectionStats(open, inUse, idle int, waitCount int64, waitDurationSeconds float64) {
	DBConnectionsOpen.Set(float64(open))
	DBConnectionsInUse.Set(float64(inUse))
	DBConnectionsIdle.Set(float64(idle))
	DBWaitCount.Set(float64(waitCount))
	DBWaitDurationSeconds.Set(waitDurationSeconds)

	DBPoolOpen.Set(float64(open))
	DBPoolInUse.Set(float64(inUse))
	DBPoolIdle.Set(float64(idle))
	DBPoolWaitCount.Set(float64(waitCount))
	DBPoolWaitDurationSeconds.Set(waitDurationSeconds)
}

func SetWSConnectionsActive(count int) {
	WSConnectionsActive.Set(float64(count))
}

func RecordWSMessageDrop(reason string) {
	WSMessageDropped.WithLabelValues(reason).Inc()
}
