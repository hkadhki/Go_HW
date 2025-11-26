package service

import (
	"context"
	"fmt"
	"ledger/domain"
	"log"
	"sync"
	"time"
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

func (s *ledgerService) GetSpendingSummary(ctx context.Context, req domain.GetSpendingSummaryRequest) (domain.SpendingSummary, error) {
	if req.From.IsZero() || req.To.IsZero() {
		return nil, fmt.Errorf("both from and to dates are required")
	}

	if req.From.After(req.To) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}

	if req.To.Sub(req.From) > 365*24*time.Hour {
		return nil, fmt.Errorf("period cannot exceed 1 year")
	}

	categories, err := s.getCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	if len(categories) == 0 {
		return make(domain.SpendingSummary), nil
	}

	summary, err := s.calculateSpendingParallel(ctx, categories, req.From, req.To)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate spending: %w", err)
	}

	return summary, nil
}

func (s *ledgerService) getCategories(ctx context.Context) ([]string, error) {
	transactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	budgets, err := s.budgetRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	categorySet := make(map[string]bool)

	for _, tx := range transactions {
		if tx.Category != "" {
			categorySet[tx.Category] = true
		}
	}

	for _, budget := range budgets {
		if budget.Category != "" {
			categorySet[budget.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for category := range categorySet {
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *ledgerService) calculateSpendingParallel(ctx context.Context, categories []string, from, to time.Time) (domain.SpendingSummary, error) {
	var wg sync.WaitGroup
	results := make(chan struct {
		category string
		amount   float64
		err      error
	}, len(categories))

	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()

	go s.heartbeat(heartbeatCtx, len(categories))

	log.Printf("Starting parallel calculation for %d categories", len(categories))

	for _, category := range categories {
		wg.Add(1)

		go func(cat string) {
			defer wg.Done()

			if ctx.Err() != nil {
				results <- struct {
					category string
					amount   float64
					err      error
				}{cat, 0, ctx.Err()}
				return
			}

			amount, err := s.transactionRepo.GetSpendingByCategoryAndPeriod(ctx, cat, from, to)
			results <- struct {
				category string
				amount   float64
				err      error
			}{cat, amount, err}
		}(category)
	}

	go func() {
		wg.Wait()
		close(results)
		stopHeartbeat()
		log.Printf("All calculation goroutines completed")
	}()

	summary := make(domain.SpendingSummary)
	processed := 0

	for result := range results {
		processed++

		if ctx.Err() != nil {
			log.Printf("Calculation cancelled after processing %d/%d categories", processed, len(categories))
			return nil, ctx.Err()
		}

		if result.err != nil && result.err != context.Canceled {
			return nil, fmt.Errorf("failed to calculate spending for category %s: %w", result.category, result.err)
		}

		if result.amount > 0 {
			summary[result.category] = result.amount
		}
	}

	log.Printf("Parallel calculation completed: processed %d categories", processed)
	return summary, nil
}

func (s *ledgerService) heartbeat(ctx context.Context, totalCategories int) {
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	beatCount := 0

	log.Printf("Heartbeat started for calculation of %d categories", totalCategories)

	for {
		select {
		case <-ctx.Done():
			duration := time.Since(startTime)
			log.Printf("Heartbeat stopped after %v (%d beats)", duration, beatCount)
			return
		case <-ticker.C:
			beatCount++
			duration := time.Since(startTime)
			log.Printf("Calculation in progress... [%v elapsed, beat #%d]", duration.Truncate(time.Millisecond), beatCount)
		}
	}
}

func (s *ledgerService) calculateCategorySpending(ctx context.Context, category string, from, to time.Time) (float64, error) {
	transactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	var total float64
	for _, tx := range transactions {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		if tx.Category == category && !tx.Date.Before(from) && !tx.Date.After(to) {
			total += tx.Amount
		}
	}

	return total, nil
}
