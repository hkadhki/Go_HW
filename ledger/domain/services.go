package domain

import (
	"context"
	"errors"
)

var (
	ErrBudgetNotFound = errors.New("budget not found")
	ErrBudgetExceeded = errors.New("budget exceeded")
)

type BudgetService struct {
	budgetRepo BudgetRepository
	transRepo  TransactionRepository
}

func NewBudgetService(budgetRepo BudgetRepository, transRepo TransactionRepository) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
		transRepo:  transRepo,
	}
}

func (s *BudgetService) CanSpend(ctx context.Context, category string, amount float64) (bool, error) {
	budget, err := s.budgetRepo.GetByCategory(ctx, category)
	if err != nil {
		return false, err
	}
	if budget == nil {
		return false, ErrBudgetNotFound
	}

	spent, err := s.transRepo.GetTotalByCategory(ctx, category)
	if err != nil {
		return false, err
	}

	return spent+amount <= budget.Limit, nil
}

func (s *BudgetService) GetRemainingBudget(ctx context.Context, category string) (float64, error) {
	budget, err := s.budgetRepo.GetByCategory(ctx, category)
	if err != nil {
		return 0, err
	}
	if budget == nil {
		return 0, ErrBudgetNotFound
	}

	spent, err := s.transRepo.GetTotalByCategory(ctx, category)
	if err != nil {
		return 0, err
	}

	remaining := budget.Limit - spent
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

type TransactionService struct {
	transRepo     TransactionRepository
	budgetService *BudgetService
}

func NewTransactionService(transRepo TransactionRepository, budgetService *BudgetService) *TransactionService {
	return &TransactionService{
		transRepo:     transRepo,
		budgetService: budgetService,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, transaction Transaction) (int, error) {
	if err := transaction.Validate(); err != nil {
		return 0, err
	}

	canSpend, err := s.budgetService.CanSpend(ctx, transaction.Category, transaction.Amount)
	if err != nil {
		return 0, err
	}
	if !canSpend {
		return 0, ErrBudgetExceeded
	}

	return s.transRepo.Create(ctx, transaction)
}

func (s *TransactionService) GetTransactionHistory(ctx context.Context) ([]Transaction, error) {
	return s.transRepo.List(ctx)
}
