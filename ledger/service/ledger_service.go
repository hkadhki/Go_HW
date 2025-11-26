package service

import (
	"context"
	"fmt"
	"ledger/domain"
)

type ledgerService struct {
	transactionRepo domain.TransactionRepository
	budgetRepo      domain.BudgetRepository
}

func NewLedgerService(
	transactionRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
) LedgerService {
	return &ledgerService{
		transactionRepo: transactionRepo,
		budgetRepo:      budgetRepo,
	}
}

func (s *ledgerService) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (*domain.TransactionResponse, error) {
	if err := s.validateTransactionRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	transaction := req.ToEntity()

	if err := s.checkBudgetRule(ctx, transaction.Category, transaction.Amount); err != nil {
		return nil, err
	}

	id, err := s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	transaction.ID = id
	response := domain.TransactionResponseFromEntity(transaction)
	return &response, nil
}

func (s *ledgerService) ListTransactions(ctx context.Context) ([]domain.TransactionResponse, error) {
	transactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	responses := make([]domain.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = domain.TransactionResponseFromEntity(transaction)
	}

	return responses, nil
}

func (s *ledgerService) CreateBudget(ctx context.Context, req domain.CreateBudgetRequest) (*domain.BudgetResponse, error) {
	if err := s.validateBudgetRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	budget := req.ToEntity()

	if err := s.budgetRepo.Save(ctx, budget); err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	response := domain.BudgetResponseFromEntity(budget)
	return &response, nil
}

func (s *ledgerService) ListBudgets(ctx context.Context) ([]domain.BudgetResponse, error) {
	budgets, err := s.budgetRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}

	responses := make([]domain.BudgetResponse, len(budgets))
	for i, budget := range budgets {
		responses[i] = domain.BudgetResponseFromEntity(budget)
	}

	return responses, nil
}

func (s *ledgerService) HealthCheck(ctx context.Context) error {
	if _, err := s.budgetRepo.List(ctx); err != nil {
		return fmt.Errorf("budget repository unavailable: %w", err)
	}

	if _, err := s.transactionRepo.List(ctx); err != nil {
		return fmt.Errorf("transaction repository unavailable: %w", err)
	}

	return nil
}

// Валидации
func (s *ledgerService) validateTransactionRequest(req domain.CreateTransactionRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}
	return nil
}

func (s *ledgerService) validateBudgetRequest(req domain.CreateBudgetRequest) error {
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}
	if req.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	return nil
}

func (s *ledgerService) checkBudgetRule(ctx context.Context, category string, amount float64) error {
	budget, err := s.budgetRepo.GetByCategory(ctx, category)
	if err != nil {
		return fmt.Errorf("failed to get budget: %w", err)
	}

	if budget == nil {
		return domain.ErrBudgetNotFound
	}

	spent, err := s.transactionRepo.GetTotalByCategory(ctx, category)
	if err != nil {
		return fmt.Errorf("failed to get spent amount: %w", err)
	}

	if spent+amount > budget.Limit {
		return domain.ErrBudgetExceeded
	}

	return nil
}
