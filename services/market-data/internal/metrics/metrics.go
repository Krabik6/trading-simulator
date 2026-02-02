package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PricesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "market_data",
			Name:      "prices_processed_total",
			Help:      "Total number of prices processed",
		},
		[]string{"symbol", "source"},
	)

	PricesErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "market_data",
			Name:      "prices_errors_total",
			Help:      "Total number of price processing errors",
		},
		[]string{"symbol", "error_type"},
	)

	KafkaSendDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "market_data",
			Name:      "kafka_send_duration_seconds",
			Help:      "Duration of Kafka send operations",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"symbol"},
	)

	KafkaSendErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "market_data",
			Name:      "kafka_send_errors_total",
			Help:      "Total number of Kafka send errors",
		},
		[]string{"symbol"},
	)

	ClientConnectionStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "market_data",
			Name:      "client_connected",
			Help:      "Client connection status (1 = connected, 0 = disconnected)",
		},
	)

	CurrentPrice = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "market_data",
			Name:      "current_price",
			Help:      "Current price by symbol and type (bid/ask)",
		},
		[]string{"symbol", "type"},
	)
)

func RecordPrice(symbol, source string) {
	PricesProcessed.WithLabelValues(symbol, source).Inc()
}

func RecordError(symbol, errorType string) {
	PricesErrors.WithLabelValues(symbol, errorType).Inc()
}

func RecordKafkaSend(symbol string, duration float64) {
	KafkaSendDuration.WithLabelValues(symbol).Observe(duration)
}

func RecordKafkaError(symbol string) {
	KafkaSendErrors.WithLabelValues(symbol).Inc()
}

func SetClientConnected(connected bool) {
	if connected {
		ClientConnectionStatus.Set(1)
	} else {
		ClientConnectionStatus.Set(0)
	}
}

func SetCurrentPrice(symbol string, bid, ask float64) {
	CurrentPrice.WithLabelValues(symbol, "bid").Set(bid)
	CurrentPrice.WithLabelValues(symbol, "ask").Set(ask)
}
