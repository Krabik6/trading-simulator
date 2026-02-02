package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	cleanupDatabase(t)

	email := uniqueEmail("register_success")
	body := map[string]string{
		"email":    email,
		"password": "password123",
	}

	resp := makeRequest(t, "POST", "/auth/register", body, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		UserID int64  `json:"user_id"`
		Token  string `json:"token"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Greater(t, result.UserID, int64(0))
	assert.NotEmpty(t, result.Token)

	// Verify account was created with initial balance
	accountResp := makeRequest(t, "GET", "/account", nil, result.Token)
	defer accountResp.Body.Close()

	assert.Equal(t, http.StatusOK, accountResp.StatusCode)

	var accountInfo struct {
		Balance string `json:"balance"`
	}
	err = json.NewDecoder(accountResp.Body).Decode(&accountInfo)
	require.NoError(t, err)

	assert.Equal(t, "10000.00", accountInfo.Balance)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	cleanupDatabase(t)

	email := uniqueEmail("duplicate")
	body := map[string]string{
		"email":    email,
		"password": "password123",
	}

	// First registration
	resp1 := makeRequest(t, "POST", "/auth/register", body, "")
	resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Second registration with same email
	resp2 := makeRequest(t, "POST", "/auth/register", body, "")
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	errMsg := parseErrorResponse(t, resp2)
	assert.Contains(t, errMsg, "already exists")
}

func TestRegister_InvalidEmail(t *testing.T) {
	cleanupDatabase(t)

	testCases := []struct {
		name  string
		email string
	}{
		{"no_at_sign", "invalidemail"},
		{"empty", ""},
		{"too_short", "a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]string{
				"email":    tc.email,
				"password": "password123",
			}

			resp := makeRequest(t, "POST", "/auth/register", body, "")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestRegister_InvalidPassword(t *testing.T) {
	cleanupDatabase(t)

	testCases := []struct {
		name     string
		password string
	}{
		{"too_short", "12345"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]string{
				"email":    uniqueEmail("invalid_pass"),
				"password": tc.password,
			}

			resp := makeRequest(t, "POST", "/auth/register", body, "")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestLogin_Success(t *testing.T) {
	cleanupDatabase(t)

	email := uniqueEmail("login_success")
	password := "password123"

	// Register first
	user := registerUser(t, email, password)

	// Login
	body := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := makeRequest(t, "POST", "/auth/login", body, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		UserID int64  `json:"user_id"`
		Token  string `json:"token"`
	}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, user.UserID, result.UserID)
	assert.NotEmpty(t, result.Token)

	// Verify token works
	accountResp := makeRequest(t, "GET", "/account", nil, result.Token)
	defer accountResp.Body.Close()
	assert.Equal(t, http.StatusOK, accountResp.StatusCode)
}

func TestLogin_WrongPassword(t *testing.T) {
	cleanupDatabase(t)

	email := uniqueEmail("wrong_password")
	password := "password123"

	// Register first
	registerUser(t, email, password)

	// Login with wrong password
	body := map[string]string{
		"email":    email,
		"password": "wrongpassword",
	}

	resp := makeRequest(t, "POST", "/auth/login", body, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	errMsg := parseErrorResponse(t, resp)
	assert.Contains(t, errMsg, "invalid credentials")
}

func TestLogin_NonExistentUser(t *testing.T) {
	cleanupDatabase(t)

	body := map[string]string{
		"email":    "nonexistent@test.com",
		"password": "password123",
	}

	resp := makeRequest(t, "POST", "/auth/login", body, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedEndpoint_NoToken(t *testing.T) {
	resp := makeRequest(t, "GET", "/account", nil, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	resp := makeRequest(t, "GET", "/account", nil, "invalid-token")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
