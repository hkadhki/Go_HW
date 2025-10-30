package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"ledger/models"
	"time"
)

type LedgerService struct {
	transactions  []models.Transaction
	Budgets       map[string]models.Budget
	BudgetsAmount map[string]float64
	id            int
}

func NewLedgerService() *LedgerService {
	return &LedgerService{
		transactions:  make([]models.Transaction, 0),
		Budgets:       make(map[string]models.Budget),
		BudgetsAmount: make(map[string]float64),
		id:            0,
	}
}

func (ls *LedgerService) AddTransaction(transaction models.Transaction) (int, error) {
	if err := models.CheckValid(transaction); err != nil {
		return 0, err
	}

	budget, exists := ls.Budgets[transaction.Category]
	if exists {
		if ls.BudgetsAmount[transaction.Category]+transaction.Amount > budget.Limit {
			return 0, errors.New("budget exceeded")
		}
	} else {
		return 0, errors.New("budget not found")
	}

	if transaction.Date == "" {
		transaction.Date = time.Now().Format("2006-01-02 15:04:05")
	}

	transaction.ID = ls.id
	ls.id++

	ls.transactions = append(ls.transactions, transaction)

	if exists {
		ls.BudgetsAmount[transaction.Category] += transaction.Amount
	}

	return transaction.ID, nil
}

func (ls *LedgerService) ListTransactions() []models.Transaction {
	result := make([]models.Transaction, len(ls.transactions))
	copy(result, ls.transactions)
	return result
}

func (ls *LedgerService) SetBudget(b models.Budget) error {
	if err := models.CheckValid(b); err != nil {
		return err
	}

	ls.Budgets[b.Category] = b

	return nil
}

func (ls *LedgerService) LoadBudgets(r io.Reader) error {
	reader := bufio.NewReader(r)

	data, err := io.ReadAll(reader)
	if err != nil {
		return errors.New("ошибка чтения")
	}

	var budgetList []models.Budget
	if err := json.Unmarshal(data, &budgetList); err != nil {
		return errors.New("ошибка парсинга в json")
	}

	for _, budget := range budgetList {
		if err := ls.SetBudget(budget); err != nil {
			return errors.New("ошибка сохранения бюджета")
		}
	}

	return nil
}

func (ls *LedgerService) ListBudgets() []models.Budget {
	result := make([]models.Budget, 0, len(ls.Budgets))
	for _, budget := range ls.Budgets {
		result = append(result, budget)
	}
	return result
}
