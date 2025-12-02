package domain

import "time"

type BulkTransactionRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions"`
}

type BulkTransactionResult struct {
	Index int    `json:"index"`
	ID    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type BulkTransactionResponse struct {
	Total    int                     `json:"total"`
	Accepted int                     `json:"accepted"`
	Rejected int                     `json:"rejected"`
	Errors   []BulkTransactionResult `json:"errors"`
}

type CreateTransactionRequest struct {
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (dto CreateTransactionRequest) ToEntity() Transaction {
	return Transaction{
		Amount:      dto.Amount,
		Category:    dto.Category,
		Description: dto.Description,
		Date:        dto.Date,
	}
}

type TransactionResponse struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func TransactionResponseFromEntity(entity Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          entity.ID,
		Amount:      entity.Amount,
		Category:    entity.Category,
		Description: entity.Description,
		Date:        entity.Date,
	}
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

func (dto CreateBudgetRequest) ToEntity() Budget {
	period := dto.Period
	if period == "" {
		period = "monthly"
	}

	return Budget{
		Category: dto.Category,
		Limit:    dto.Limit,
		Period:   period,
	}
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

func BudgetResponseFromEntity(entity Budget) BudgetResponse {
	return BudgetResponse{
		Category: entity.Category,
		Limit:    entity.Limit,
		Period:   entity.Period,
	}
}

type SpendingSummary map[string]float64

type GetSpendingSummaryRequest struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
