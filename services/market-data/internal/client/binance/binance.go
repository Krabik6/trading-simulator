package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
	"github.com/Krabik6/trading-simulator/market-data/internal/logger"
)

const (
	wsBaseURL      = "wss://stream.binance.com:9443/ws"
	maxRetries     = 10
	baseBackoff    = 1 * time.Second
	maxBackoff     = 60 * time.Second
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = 30 * time.Second
	throttleWindow = 200 * time.Millisecond // max 5 updates/sec per symbol
)

// bookTickerMsg is the Binance bookTicker WebSocket message format.
type bookTickerMsg struct {
	UpdateID int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

type BinanceClient struct {
	symbols []domain.Symbol
	conn    *websocket.Conn
	mu      sync.Mutex
	stop    chan struct{}
	done    chan struct{}

	// Throttling: store latest price per symbol, flush periodically.
	latestMu     sync.Mutex
	latestPrices map[string]domain.Price
}

func NewBinanceClient(symbols []domain.Symbol) *BinanceClient {
	logger.Info("binance client created", "symbols", symbols)
	return &BinanceClient{
		symbols:      symbols,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
		latestPrices: make(map[string]domain.Price),
	}
}

func (c *BinanceClient) StreamPrices(ctx context.Context) (<-chan domain.PriceWithError, error) {
	priceCh := make(chan domain.PriceWithError, 100)

	go func() {
		defer close(priceCh)
		defer close(c.done)

		retries := 0
		for {
			select {
			case <-ctx.Done():
				logger.Info("binance stream stopped by context")
				return
			case <-c.stop:
				logger.Info("binance stream stopped")
				return
			default:
			}

			err := c.connectAndStream(ctx, priceCh)
			if err == nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			default:
			}

			retries++
			if retries > maxRetries {
				logger.Error("binance: max retries reached, giving up", "retries", retries)
				priceCh <- domain.PriceWithError{Error: fmt.Errorf("binance: max retries (%d) reached: %w", maxRetries, err)}
				return
			}

			backoff := time.Duration(math.Min(
				float64(baseBackoff)*math.Pow(2, float64(retries-1)),
				float64(maxBackoff),
			))
			logger.Warn("binance: reconnecting", "retry", retries, "backoff", backoff, "error", err)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			}
		}
	}()

	return priceCh, nil
}

// startThrottledFlusher periodically flushes buffered prices to the output channel.
func (c *BinanceClient) startThrottledFlusher(ctx context.Context, priceCh chan<- domain.PriceWithError) {
	ticker := time.NewTicker(throttleWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.latestMu.Lock()
			for sym, price := range c.latestPrices {
				select {
				case priceCh <- domain.PriceWithError{Price: price}:
				default:
				}
				delete(c.latestPrices, sym)
			}
			c.latestMu.Unlock()
		case <-ctx.Done():
			return
		case <-c.stop:
			return
		}
	}
}

func (c *BinanceClient) connectAndStream(ctx context.Context, priceCh chan<- domain.PriceWithError) error {
	streams := make([]string, len(c.symbols))
	for i, s := range c.symbols {
		streams[i] = strings.ToLower(string(s)) + "@bookTicker"
	}
	url := wsBaseURL + "/" + strings.Join(streams, "/")

	logger.Info("binance: connecting", "url", url)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("binance: dial failed: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		conn.Close()
	}()

	logger.Info("binance: connected")

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Ping goroutine to keep connection alive.
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				if c.conn != nil {
					c.conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						c.mu.Unlock()
						logger.Warn("binance: ping failed", "error", err)
						return
					}
				}
				c.mu.Unlock()
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			}
		}
	}()

	// Throttled flusher — sends buffered prices every 200ms
	go c.startThrottledFlusher(ctx, priceCh)

	conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.stop:
			return nil
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			<-pingDone
			return fmt.Errorf("binance: read failed: %w", err)
		}

		conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg bookTickerMsg
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warn("binance: unmarshal failed", "error", err, "message", string(message))
			continue
		}

		if msg.Symbol == "" {
			continue
		}

		bid, err := strconv.ParseFloat(msg.BidPrice, 64)
		if err != nil {
			continue
		}
		ask, err := strconv.ParseFloat(msg.AskPrice, 64)
		if err != nil {
			continue
		}

		price := domain.Price{
			Symbol:    domain.Symbol(msg.Symbol),
			Bid:       bid,
			Ask:       ask,
			Timestamp: time.Now(),
			Source:    "binance",
		}

		// Buffer latest price — flusher will send it.
		c.latestMu.Lock()
		c.latestPrices[msg.Symbol] = price
		c.latestMu.Unlock()
	}
}

func (c *BinanceClient) Close() error {
	select {
	case <-c.stop:
	default:
		close(c.stop)
	}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn != nil {
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}

	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
	}

	logger.Info("binance client closed")
	return nil
}

func (c *BinanceClient) Name() string {
	return "binance"
}
