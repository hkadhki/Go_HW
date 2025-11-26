package services

import (
	"database/sql"
	"errors"
	"fmt"
	"ledger/db"
	"ledger/models"
	"time"
)

type LedgerService struct {
	db *sql.DB
}

func NewLedgerService() *LedgerService {
	return &LedgerService{
		db: db.DB,
	}
}

func (ls *LedgerService) AddTransaction(transaction models.Transaction) (int, error) {
	if err := transaction.Validate(); err != nil {
		return 0, err
	}

	// Check if budget exists and validate limit
	var limit float64
	err := ls.db.QueryRow("SELECT limit_amount FROM budgets WHERE category = $1", transaction.Category).Scan(&limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("budget not found")
		}
		return 0, fmt.Errorf("database error: %v", err)
	}

	// Calculate spent amount
	var spent float64
	err = ls.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1", transaction.Category).Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("database error: %v", err)
	}

	// Check budget
	if spent+transaction.Amount > limit {
		return 0, errors.New("budget exceeded")
	}

	// Set date if empty
	if transaction.Date == "" {
		transaction.Date = time.Now().Format("2006-01-02 15:04:05")
	}

	// Insert transaction
	var id int
	err = ls.db.QueryRow(
		"INSERT INTO expenses (amount, category, description, date) VALUES ($1, $2, $3, $4) RETURNING id",
		transaction.Amount, transaction.Category, transaction.Description, transaction.Date,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %v", err)
	}

	return id, nil
}

func (ls *LedgerService) ListTransactions() ([]models.Transaction, error) {
	rows, err := ls.db.Query("SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &tx.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (ls *LedgerService) SetBudget(budget models.Budget) error {
	if err := budget.Validate(); err != nil {
		return err
	}

	_, err := ls.db.Exec(
		"INSERT INTO budgets (category, limit_amount) VALUES ($1, $2) ON CONFLICT (category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount",
		budget.Category, budget.Limit,
	)

	if err != nil {
		return fmt.Errorf("failed to set budget: %v", err)
	}

	return nil
}

func (ls *LedgerService) ListBudgets() ([]models.Budget, error) {
	rows, err := ls.db.Query("SELECT category, limit_amount FROM budgets ORDER BY category")
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %v", err)
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var budget models.Budget
		err := rows.Scan(&budget.Category, &budget.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %v", err)
		}
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

func (ls *LedgerService) Reset() {
}
