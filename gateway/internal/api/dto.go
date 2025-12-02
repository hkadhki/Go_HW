package api

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type TransactionResponse struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

type SpendingSummaryResponse map[string]float64

type BulkTransactionResponse struct {
	Total    int                     `json:"total"`
	Accepted int                     `json:"accepted"`
	Rejected int                     `json:"rejected"`
	Errors   []BulkTransactionResult `json:"errors"`
}

type BulkTransactionResult struct {
	Index int    `json:"index"`
	ID    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}
