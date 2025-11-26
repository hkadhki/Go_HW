package main

import (
	"ledger/models"
	"ledger/services"
	"testing"
)

// Test helper function to reset service state
func Reset(service *services.LedgerService) {
	// Since fields are private, we need to create a new service
	// In real implementation, you might want to add Reset method to LedgerService
}

func TestTransactionValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		transaction models.Transaction
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid transaction",
			transaction: models.Transaction{
				Amount:      100.0,
				Category:    "food",
				Description: "lunch",
				Date:        "2024-01-15 12:00:00",
			},
			shouldError: false,
		},
		{
			name: "zero amount",
			transaction: models.Transaction{
				Amount:      0,
				Category:    "food",
				Description: "lunch",
			},
			shouldError: true,
			errorMsg:    "сумма транзакции должна быть положительным числом",
		},
		{
			name: "negative amount",
			transaction: models.Transaction{
				Amount:      -50.0,
				Category:    "food",
				Description: "lunch",
			},
			shouldError: true,
			errorMsg:    "сумма транзакции должна быть положительным числом",
		},
		{
			name: "empty category",
			transaction: models.Transaction{
				Amount:      100.0,
				Category:    "",
				Description: "lunch",
			},
			shouldError: true,
			errorMsg:    "категория транзакции не может быть пустой",
		},
		{
			name: "whitespace category",
			transaction: models.Transaction{
				Amount:      100.0,
				Category:    "   ",
				Description: "lunch",
			},
			shouldError: true,
			errorMsg:    "категория транзакции не может быть пустой",
		},
		{
			name: "invalid date format",
			transaction: models.Transaction{
				Amount:      100.0,
				Category:    "food",
				Description: "lunch",
				Date:        "invalid-date",
			},
			shouldError: true,
			errorMsg:    "некорректный формат даты",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.transaction.Validate()

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tc.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%s'", err.Error())
				}
			}
		})
	}
}

func TestBudgetValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		budget      models.Budget
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid budget",
			budget: models.Budget{
				Category: "food",
				Limit:    5000.0,
				Period:   "monthly",
			},
			shouldError: false,
		},
		{
			name: "valid budget with empty period",
			budget: models.Budget{
				Category: "food",
				Limit:    5000.0,
				Period:   "",
			},
			shouldError: false,
		},
		{
			name: "zero limit",
			budget: models.Budget{
				Category: "food",
				Limit:    0,
				Period:   "monthly",
			},
			shouldError: true,
			errorMsg:    "лимит бюджета должен быть положительным числом",
		},
		{
			name: "negative limit",
			budget: models.Budget{
				Category: "food",
				Limit:    -100.0,
				Period:   "monthly",
			},
			shouldError: true,
			errorMsg:    "лимит бюджета должен быть положительным числом",
		},
		{
			name: "empty category",
			budget: models.Budget{
				Category: "",
				Limit:    5000.0,
				Period:   "monthly",
			},
			shouldError: true,
			errorMsg:    "категория бюджета не может быть пустой",
		},
		{
			name: "whitespace category",
			budget: models.Budget{
				Category: "   ",
				Limit:    5000.0,
				Period:   "monthly",
			},
			shouldError: true,
			errorMsg:    "категория бюджета не может быть пустой",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.budget.Validate()

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tc.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%s'", err.Error())
				}
			}
		})
	}
}

func TestBudgetExceeded(t *testing.T) {
	service := services.NewLedgerService()

	t.Run("transaction within budget", func(t *testing.T) {
		// Set up budget
		budget := models.Budget{
			Category: "food",
			Limit:    5000.0,
			Period:   "monthly",
		}
		err := service.SetBudget(budget)
		if err != nil {
			t.Fatalf("Failed to set budget: %v", err)
		}

		// Add transaction within budget
		transaction := models.Transaction{
			Amount:      1000.0,
			Category:    "food",
			Description: "groceries",
		}

		id, err := service.AddTransaction(transaction)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if id != 0 {
			t.Errorf("Expected ID 0, got: %d", id)
		}

		// Verify transaction was added
		transactions := service.ListTransactions()
		if len(transactions) != 1 {
			t.Errorf("Expected 1 transaction, got %d", len(transactions))
		}
	})

	t.Run("transaction exceeds budget", func(t *testing.T) {
		// Reset service for clean test
		service = services.NewLedgerService()

		// Set up budget
		budget := models.Budget{
			Category: "food",
			Limit:    1000.0,
			Period:   "monthly",
		}
		err := service.SetBudget(budget)
		if err != nil {
			t.Fatalf("Failed to set budget: %v", err)
		}

		// Add first transaction
		transaction1 := models.Transaction{
			Amount:      800.0,
			Category:    "food",
			Description: "groceries",
		}

		_, err = service.AddTransaction(transaction1)
		if err != nil {
			t.Fatalf("Failed to add first transaction: %v", err)
		}

		// Try to add second transaction that exceeds budget
		transaction2 := models.Transaction{
			Amount:      300.0,
			Category:    "food",
			Description: "restaurant",
		}

		_, err = service.AddTransaction(transaction2)
		if err == nil {
			t.Error("Expected error for exceeded budget, got nil")
		} else if err.Error() != "budget exceeded" {
			t.Errorf("Expected 'budget exceeded' error, got: %v", err)
		}

		// Verify second transaction was not added
		transactions := service.ListTransactions()
		if len(transactions) != 1 {
			t.Errorf("Expected 1 transaction after budget exceeded, got %d", len(transactions))
		}
	})

	t.Run("transaction without budget", func(t *testing.T) {
		// Reset service for clean test
		service = services.NewLedgerService()

		transaction := models.Transaction{
			Amount:      100.0,
			Category:    "entertainment", // No budget set for this category
			Description: "cinema",
		}

		_, err := service.AddTransaction(transaction)
		if err == nil {
			t.Error("Expected error for missing budget, got nil")
		} else if err.Error() != "budget not found" {
			t.Errorf("Expected 'budget not found' error, got: %v", err)
		}
	})
}
