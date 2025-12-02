package service

import (
	"context"
	"ledger/domain"
)

type LedgerService interface {
	CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (*domain.TransactionResponse, error)
	ListTransactions(ctx context.Context) ([]domain.TransactionResponse, error)
	CreateBudget(ctx context.Context, req domain.CreateBudgetRequest) (*domain.BudgetResponse, error)
	ListBudgets(ctx context.Context) ([]domain.BudgetResponse, error)
	HealthCheck(ctx context.Context) error
	GetSpendingSummary(ctx context.Context, req domain.GetSpendingSummaryRequest) (domain.SpendingSummary, error)
	CreateTransactionsBulk(ctx context.Context, req domain.BulkTransactionRequest, workers int) (*domain.BulkTransactionResponse, error)
}
