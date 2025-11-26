package pg

import (
	"context"
	"database/sql"
	"fmt"
	domain2 "ledger/domain"
)

type budgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) domain2.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Save(ctx context.Context, budget domain2.Budget) error {
	query := `
		INSERT INTO budgets (category, limit_amount) 
		VALUES ($1, $2) 
		ON CONFLICT (category) 
		DO UPDATE SET limit_amount = EXCLUDED.limit_amount
	`

	_, err := r.db.ExecContext(ctx, query, budget.Category, budget.Limit)
	if err != nil {
		return fmt.Errorf("failed to save budget: %w", err)
	}

	return nil
}

func (r *budgetRepository) GetByCategory(ctx context.Context, category string) (*domain2.Budget, error) {
	query := `
		SELECT category, limit_amount 
		FROM budgets 
		WHERE category = $1
	`

	var budget domain2.Budget
	err := r.db.QueryRowContext(ctx, query, category).Scan(&budget.Category, &budget.Limit)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get budget by category: %w", err)
	}

	// Устанавливаем период по умолчанию
	budget.Period = "monthly"

	return &budget, nil
}

func (r *budgetRepository) List(ctx context.Context) ([]domain2.Budget, error) {
	query := `
		SELECT category, limit_amount 
		FROM budgets 
		ORDER BY category
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	defer rows.Close()

	var budgets []domain2.Budget
	for rows.Next() {
		var budget domain2.Budget
		err := rows.Scan(&budget.Category, &budget.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}

		// Устанавливаем период по умолчанию
		budget.Period = "monthly"
		budgets = append(budgets, budget)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating budgets: %w", err)
	}

	return budgets, nil
}

func (r *budgetRepository) Exists(ctx context.Context, category string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM budgets WHERE category = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, category).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check budget existence: %w", err)
	}

	return exists, nil
}
