package domain

import "context"

type TransactionRepository interface {
	Create(ctx context.Context, transaction Transaction) (int, error)

	List(ctx context.Context) ([]Transaction, error)

	GetTotalByCategory(ctx context.Context, category string) (float64, error)

	GetByID(ctx context.Context, id int) (*Transaction, error)
}

type BudgetRepository interface {
	Save(ctx context.Context, budget Budget) error

	GetByCategory(ctx context.Context, category string) (*Budget, error)

	List(ctx context.Context) ([]Budget, error)

	Exists(ctx context.Context, category string) (bool, error)
}
