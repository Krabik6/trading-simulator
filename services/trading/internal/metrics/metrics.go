package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
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
