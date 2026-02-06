package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"trading/internal/delivery/http/handler"
	"trading/internal/delivery/http/middleware"
	"trading/internal/domain"
)

type Router struct {
	chi.Router
}

type RouterDeps struct {
	AuthMiddleware   *middleware.AuthMiddleware
	AuthHandler      *handler.AuthHandler
	AccountHandler   *handler.AccountHandler
	OrderHandler     *handler.OrderHandler
	PositionHandler  *handler.PositionHandler
	TradeHandler     *handler.TradeHandler
	UserHandler      *handler.UserHandler
	PriceHandler     *handler.PriceHandler
	CandleHandler    *handler.CandleHandler
	TickerHandler    *handler.TickerHandler
	WebSocketHandler *handler.WebSocketHandler
	UserRepo         domain.UserRepository
	HealthChecker    func() error
}

func NewRouter(deps RouterDeps) *Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// Health endpoints (no auth)
	r.Get("/health", healthHandler)
	r.Get("/ready", readyHandler(deps.HealthChecker))
	r.Handle("/metrics", promhttp.Handler())

	// WebSocket endpoint (auth via query param)
	if deps.WebSocketHandler != nil {
		r.Get("/ws", deps.WebSocketHandler.HandleWs)
	}

	// Public endpoints (no auth)
	if deps.PriceHandler != nil {
		r.Get("/prices", deps.PriceHandler.GetPrices)
		r.Get("/symbols", deps.PriceHandler.GetSymbols)
	}
	if deps.CandleHandler != nil {
		r.Get("/candles", deps.CandleHandler.GetCandles)
	}
	if deps.TickerHandler != nil {
		r.Get("/ticker24h", deps.TickerHandler.GetTicker24h)
	}

	// Auth endpoints (no auth)
	r.Post("/auth/register", deps.AuthHandler.Register)
	r.Post("/auth/login", deps.AuthHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(deps.AuthMiddleware.Authenticate)

		// Auth
		r.Post("/auth/refresh", deps.AuthHandler.Refresh)

		// User
		if deps.UserHandler != nil {
			r.Get("/user/me", deps.UserHandler.GetMe)
		}

		// Account
		r.Get("/account", deps.AccountHandler.GetAccount)

		// Orders
		r.Post("/orders", deps.OrderHandler.PlaceOrder)
		r.Get("/orders", deps.OrderHandler.GetOrders)
		r.Get("/orders/{id}", deps.OrderHandler.GetOrder)
		r.Patch("/orders/{id}", deps.OrderHandler.UpdateOrder)
		r.Delete("/orders/{id}", deps.OrderHandler.CancelOrder)

		// Positions
		r.Get("/positions", deps.PositionHandler.GetPositions)
		r.Get("/positions/{id}", deps.PositionHandler.GetPosition)
		r.Post("/positions/{id}/close", deps.PositionHandler.ClosePosition)
		r.Patch("/positions/{id}", deps.PositionHandler.UpdateTPSL)

		// Trades
		r.Get("/trades", deps.TradeHandler.GetTrades)
	})

	return &Router{r}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func readyHandler(healthChecker func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := healthChecker(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready","error":"` + err.Error() + `"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	}
}
