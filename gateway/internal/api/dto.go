package api

import "ledger/models"

type CreateTransactionRequest struct {
	Amount      float64
	Category    string
	Description string
	Date        string
}

type TransactionResponse struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

type CreateBudgetRequest struct {
	Category string
	Limit    float64
	Period   string
}

type BudgetResponse struct {
	Category string
	Limit    float64
	Period   string
}

func TransactionToCreateRequest(tx models.Transaction) CreateTransactionRequest {
	return CreateTransactionRequest{
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func CreateRequestToTransaction(req CreateTransactionRequest) models.Transaction {
	return models.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}
}

func TransactionToResponse(tx models.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func TransactionsToResponseList(transactions []models.Transaction) []TransactionResponse {
	response := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		response[i] = TransactionToResponse(tx)
	}
	return response
}

func BudgetToCreateRequest(budget models.Budget) CreateBudgetRequest {
	return CreateBudgetRequest{
		Category: budget.Category,
		Limit:    budget.Limit,
		Period:   budget.Period,
	}
}

func CreateRequestToBudget(req CreateBudgetRequest) models.Budget {

	period := req.Period
	if period == "" {
		period = "monthly"
	}

	return models.Budget{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   period,
	}
}

func BudgetToResponse(budget models.Budget) BudgetResponse {
	return BudgetResponse{
		Category: budget.Category,
		Limit:    budget.Limit,
		Period:   budget.Period,
	}
}

func BudgetsToResponseList(budgets []models.Budget) []BudgetResponse {
	response := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		response[i] = BudgetToResponse(budget)
	}
	return response
}
