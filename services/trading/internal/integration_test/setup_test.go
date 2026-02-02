package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"trading/internal/auth"
	httpdelivery "trading/internal/delivery/http"
	"trading/internal/delivery/http/handler"
	"trading/internal/delivery/http/middleware"
	"trading/internal/domain"
	"trading/internal/engine"
	"trading/internal/logger"
	"trading/internal/repository/postgres"
	accountuc "trading/internal/usecase/account"
	authuc "trading/internal/usecase/auth"
	orderuc "trading/internal/usecase/order"
	positionuc "trading/internal/usecase/position"
)

const (
	testJWTSecret      = "test-secret-key-for-integration-tests"
	testJWTExpiry      = 24
	testInitialBalance = 10000.0
	testMaxLeverage    = 100
	testMaintenanceRate = 0.005
)

var (
	testDB     *sql.DB
	testServer *httptest.Server
	testRouter *httpdelivery.Router
	testCtx    context.Context
	testCancel context.CancelFunc

	// Repositories
	userRepo     *postgres.UserRepository
	accountRepo  *postgres.AccountRepository
	orderRepo    *postgres.OrderRepository
	positionRepo *postgres.PositionRepository
	tradeRepo    *postgres.TradeRepository

	// Services
	jwtService *auth.JWTService
	eng        *engine.Engine
	priceCache *MockPriceCache

	// Use cases
	authUseCase     *authuc.UseCase
	accountUseCase  *accountuc.UseCase
	orderUseCase    *orderuc.UseCase
	positionUseCase *positionuc.UseCase
)

// MockPriceCache implements domain.PriceCache for testing
type MockPriceCache struct {
	mu     sync.RWMutex
	prices map[string]*domain.Price
}

func NewMockPriceCache() *MockPriceCache {
	return &MockPriceCache{
		prices: map[string]*domain.Price{
			"BTCUSDT": {
				Symbol:    "BTCUSDT",
				Bid:       50000,
				Ask:       50010,
				Timestamp: time.Now(),
				Source:    "mock",
			},
			"ETHUSDT": {
				Symbol:    "ETHUSDT",
				Bid:       3000,
				Ask:       3002,
				Timestamp: time.Now(),
				Source:    "mock",
			},
			"SOLUSDT": {
				Symbol:    "SOLUSDT",
				Bid:       100,
				Ask:       100.1,
				Timestamp: time.Now(),
				Source:    "mock",
			},
		},
	}
}

func (m *MockPriceCache) Get(symbol string) (*domain.Price, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	price, ok := m.prices[symbol]
	return price, ok
}

func (m *MockPriceCache) Set(symbol string, price *domain.Price) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prices[symbol] = price
}

func (m *MockPriceCache) GetAll() map[string]*domain.Price {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*domain.Price)
	for k, v := range m.prices {
		result[k] = v
	}
	return result
}

func (m *MockPriceCache) SetPrice(symbol string, bid, ask float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prices[symbol] = &domain.Price{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now(),
		Source:    "mock",
	}
}

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.Init("info")

	testCtx, testCancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer testCancel()

	// Start PostgreSQL container
	pgContainer, err := tcpostgres.RunContainer(testCtx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("trading_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("Failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := pgContainer.Terminate(testCtx); err != nil {
			fmt.Printf("Failed to terminate postgres container: %v\n", err)
		}
	}()

	// Get connection string
	connStr, err := pgContainer.ConnectionString(testCtx, "sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		if err := testDB.PingContext(testCtx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	// Apply migrations
	if err := applyMigrations(testDB); err != nil {
		fmt.Printf("Failed to apply migrations: %v\n", err)
		os.Exit(1)
	}

	// Setup test infrastructure
	if err := setupTestInfrastructure(); err != nil {
		fmt.Printf("Failed to setup test infrastructure: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func applyMigrations(db *sql.DB) error {
	// Find migrations directory
	migrationsDir := "../../migrations"

	// Read migration file
	migrationPath := filepath.Join(migrationsDir, "001_init.up.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	return nil
}

func setupTestInfrastructure() error {
	// Create database wrapper
	db := &postgres.DB{DB: testDB}

	// Create repositories
	userRepo = postgres.NewUserRepository(db)
	accountRepo = postgres.NewAccountRepository(db)
	orderRepo = postgres.NewOrderRepository(db)
	positionRepo = postgres.NewPositionRepository(db)
	tradeRepo = postgres.NewTradeRepository(db)

	// Create services
	jwtService = auth.NewJWTService(testJWTSecret, testJWTExpiry)
	eng = engine.NewEngine(testMaxLeverage, testMaintenanceRate)
	priceCache = NewMockPriceCache()

	// Create use cases
	authUseCase = authuc.NewUseCase(userRepo, accountRepo, jwtService, testInitialBalance)
	accountUseCase = accountuc.NewUseCase(accountRepo, positionRepo)
	orderUseCase = orderuc.NewUseCase(
		orderRepo,
		positionRepo,
		accountRepo,
		tradeRepo,
		priceCache,
		eng,
		[]string{"BTCUSDT", "ETHUSDT", "SOLUSDT"},
	)
	positionUseCase = positionuc.NewUseCase(
		positionRepo,
		accountRepo,
		tradeRepo,
		orderRepo,
		priceCache,
		eng,
	)

	// Create handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	accountHandler := handler.NewAccountHandler(accountUseCase)
	orderHandler := handler.NewOrderHandler(orderUseCase)
	positionHandler := handler.NewPositionHandler(positionUseCase)
	tradeHandler := handler.NewTradeHandler(tradeRepo)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Create router
	testRouter = httpdelivery.NewRouter(httpdelivery.RouterDeps{
		AuthMiddleware:  authMiddleware,
		AuthHandler:     authHandler,
		AccountHandler:  accountHandler,
		OrderHandler:    orderHandler,
		PositionHandler: positionHandler,
		TradeHandler:    tradeHandler,
		HealthChecker:   func() error { return testDB.Ping() },
	})

	// Create test server
	testServer = httptest.NewServer(testRouter)

	return nil
}

// Helper functions for tests

type testUser struct {
	Email    string
	Password string
	UserID   int64
	Token    string
}

func registerUser(t *testing.T, email, password string) *testUser {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := makeRequest(t, "POST", "/auth/register", body, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("register failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		UserID int64  `json:"user_id"`
		Token  string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	return &testUser{
		Email:    email,
		Password: password,
		UserID:   result.UserID,
		Token:    result.Token,
	}
}

func loginUser(t *testing.T, email, password string) string {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := makeRequest(t, "POST", "/auth/login", body, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("login failed: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	return result.Token
}

func makeRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, testServer.URL+path, reqBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}

	return resp
}

func parseResponse(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func parseErrorResponse(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	var result struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return string(bodyBytes)
	}
	return result.Error
}

func cleanupDatabase(t *testing.T) {
	t.Helper()

	tables := []string{"trades", "positions", "orders", "accounts", "users"}
	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("cleanup table %s: %v", table, err)
		}
	}
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.com", prefix, time.Now().UnixNano())
}
