package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"trading/internal/delivery/http/handler"
	"trading/internal/delivery/http/middleware"
)

type Router struct {
	chi.Router
}

type RouterDeps struct {
	AuthMiddleware  *middleware.AuthMiddleware
	AuthHandler     *handler.AuthHandler
	AccountHandler  *handler.AccountHandler
	OrderHandler    *handler.OrderHandler
	PositionHandler *handler.PositionHandler
	TradeHandler    *handler.TradeHandler
	HealthChecker   func() error
}

func NewRouter(deps RouterDeps) *Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// Health endpoints (no auth)
	r.Get("/health", healthHandler)
	r.Get("/ready", readyHandler(deps.HealthChecker))
	r.Handle("/metrics", promhttp.Handler())

	// Auth endpoints (no auth)
	r.Post("/auth/register", deps.AuthHandler.Register)
	r.Post("/auth/login", deps.AuthHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(deps.AuthMiddleware.Authenticate)

		// Account
		r.Get("/account", deps.AccountHandler.GetAccount)

		// Orders
		r.Post("/orders", deps.OrderHandler.PlaceOrder)
		r.Get("/orders", deps.OrderHandler.GetOrders)
		r.Delete("/orders/{id}", deps.OrderHandler.CancelOrder)

		// Positions
		r.Get("/positions", deps.PositionHandler.GetPositions)
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
