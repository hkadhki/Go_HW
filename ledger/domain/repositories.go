package domain

import (
	"context"
	"time"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction Transaction) (int, error)
	List(ctx context.Context) ([]Transaction, error)
	GetTotalByCategory(ctx context.Context, category string) (float64, error)
	GetByID(ctx context.Context, id int) (*Transaction, error)
	GetSpendingByPeriod(ctx context.Context, from, to time.Time) (SpendingSummary, error)
	GetSpendingByCategoryAndPeriod(ctx context.Context, category string, from, to time.Time) (float64, error) // ДОБАВЛЕН
}

type BudgetRepository interface {
	Save(ctx context.Context, budget Budget) error
	GetByCategory(ctx context.Context, category string) (*Budget, error)
	List(ctx context.Context) ([]Budget, error)
	Exists(ctx context.Context, category string) (bool, error)
}
