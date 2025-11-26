package api

import (
	"bytes"
	"encoding/json"
	"ledger/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTest() (*Handler, *httptest.Server) {
	ledgerService := services.NewLedgerService()
	handler := NewHandler(ledgerService)

	// Setup router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/transactions", handler.CreateTransactionHandler)
	mux.HandleFunc("GET /api/transactions", handler.ListTransactions)
	mux.HandleFunc("POST /api/budgets", handler.CreateBudget)
	mux.HandleFunc("GET /api/budgets", handler.ListBudgets)
	mux.HandleFunc("GET /ping", handler.Ping)

	// Apply middleware
	handlerChain := JSONMiddleware(LoggingMiddleware(mux))

	server := httptest.NewServer(handlerChain)
	return handler, server
}

func TestBudgetHandlers(t *testing.T) {
	_, server := setupTest()
	defer server.Close()

	t.Run("create valid budget", func(t *testing.T) {
		budgetReq := CreateBudgetRequest{
			Category: "food",
			Limit:    5000.0,
			Period:   "monthly",
		}

		jsonData, _ := json.Marshal(budgetReq)
		resp, err := http.Post(server.URL+"/api/budgets", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		// Более гибкая проверка Content-Type
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected Content-Type to contain 'application/json', got '%s'", contentType)
		}

		var response BudgetResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Category != "food" {
			t.Errorf("Expected category 'food', got '%s'", response.Category)
		}
		if response.Limit != 5000.0 {
			t.Errorf("Expected limit 5000.0, got %f", response.Limit)
		}
	})

	t.Run("create budget with invalid JSON", func(t *testing.T) {
		invalidJSON := `{"category": "food", "limit": "invalid"}`
		resp, err := http.Post(server.URL+"/api/budgets", "application/json", bytes.NewBufferString(invalidJSON))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp["error"] != "Invalid JSON" {
			t.Errorf("Expected error 'Invalid JSON', got '%s'", errorResp["error"])
		}
	})

	t.Run("create budget with zero limit", func(t *testing.T) {
		budgetReq := CreateBudgetRequest{
			Category: "food",
			Limit:    0,
			Period:   "monthly",
		}

		jsonData, _ := json.Marshal(budgetReq)
		resp, err := http.Post(server.URL+"/api/budgets", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp["error"] == "" {
			t.Error("Expected error message in response")
		}
	})

	t.Run("list budgets", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/budgets")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var budgets []BudgetResponse
		if err := json.NewDecoder(resp.Body).Decode(&budgets); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(budgets) == 0 {
			t.Error("Expected at least one budget in response")
		}
	})
}

func TestTransactionHandlers(t *testing.T) {
	_, server := setupTest()
	defer server.Close()

	t.Run("complete flow: budget -> transaction -> list", func(t *testing.T) {
		// Create budget first
		budgetReq := CreateBudgetRequest{
			Category: "food",
			Limit:    5000.0,
			Period:   "monthly",
		}

		jsonData, _ := json.Marshal(budgetReq)
		resp, err := http.Post(server.URL+"/api/budgets", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create budget: %v", err)
		}
		resp.Body.Close()

		// Create transaction
		transactionReq := CreateTransactionRequest{
			Amount:      1000.0,
			Category:    "food",
			Description: "groceries",
			Date:        "2024-01-15 12:00:00",
		}

		jsonData, _ = json.Marshal(transactionReq)
		resp, err = http.Post(server.URL+"/api/transactions", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}

		var transactionResp TransactionResponse
		if err := json.NewDecoder(resp.Body).Decode(&transactionResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if transactionResp.Amount != 1000.0 {
			t.Errorf("Expected amount 1000.0, got %f", transactionResp.Amount)
		}
		if transactionResp.Category != "food" {
			t.Errorf("Expected category 'food', got '%s'", transactionResp.Category)
		}

		// List transactions
		resp, err = http.Get(server.URL + "/api/transactions")
		if err != nil {
			t.Fatalf("Failed to list transactions: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var transactions []TransactionResponse
		if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
			t.Fatalf("Failed to decode transactions: %v", err)
		}

		if len(transactions) != 1 {
			t.Errorf("Expected 1 transaction, got %d", len(transactions))
		}
	})

	t.Run("transaction exceeds budget", func(t *testing.T) {
		// Create budget with small limit
		budgetReq := CreateBudgetRequest{
			Category: "entertainment",
			Limit:    500.0,
			Period:   "monthly",
		}

		jsonData, _ := json.Marshal(budgetReq)
		resp, err := http.Post(server.URL+"/api/budgets", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to create budget: %v", err)
		}
		resp.Body.Close()

		// Try to create transaction that exceeds budget
		transactionReq := CreateTransactionRequest{
			Amount:      600.0,
			Category:    "entertainment",
			Description: "concert",
		}

		jsonData, _ = json.Marshal(transactionReq)
		resp, err = http.Post(server.URL+"/api/transactions", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}

		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp["error"] != "budget exceeded" {
			t.Errorf("Expected error 'budget exceeded', got '%s'", errorResp["error"])
		}
	})

	t.Run("transaction with invalid JSON", func(t *testing.T) {
		invalidJSON := `{"amount": "invalid", "category": "food"}`
		resp, err := http.Post(server.URL+"/api/transactions", "application/json", bytes.NewBufferString(invalidJSON))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp["error"] != "Invalid JSON" {
			t.Errorf("Expected error 'Invalid JSON', got '%s'", errorResp["error"])
		}
	})

	t.Run("transaction without budget", func(t *testing.T) {
		transactionReq := CreateTransactionRequest{
			Amount:      100.0,
			Category:    "nonexistent",
			Description: "test",
		}

		jsonData, _ := json.Marshal(transactionReq)
		resp, err := http.Post(server.URL+"/api/transactions", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Fatalf("Failed to decode error response: %v", err)
		}

		if errorResp["error"] != "budget not found" {
			t.Errorf("Expected error 'budget not found', got '%s'", errorResp["error"])
		}
	})
}

func TestPingHandler(t *testing.T) {
	_, server := setupTest()
	defer server.Close()

	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "pong" {
		t.Errorf("Expected message 'pong', got '%s'", response["message"])
	}
}
