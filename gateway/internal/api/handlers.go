package api

import (
	"context"
	"encoding/json"
	"ledger/domain"
	"ledger/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	ledgerService service.LedgerService
}

func NewHandler(ledgerService service.LedgerService) *Handler {
	return &Handler{
		ledgerService: ledgerService,
	}
}

func (h *Handler) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if r.Context().Err() != nil {
		return
	}

	domainReq := domain.CreateTransactionRequest{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
	}

	if req.Date != "" {
		if date, err := time.Parse("2006-01-02 15:04:05", req.Date); err == nil {
			domainReq.Date = date
		} else if date, err := time.Parse("2006-01-02", req.Date); err == nil {
			domainReq.Date = date
		}
	}

	response, err := h.ledgerService.CreateTransaction(r.Context(), domainReq)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	apiResponse := TransactionResponse{
		ID:          response.ID,
		Amount:      response.Amount,
		Category:    response.Category,
		Description: response.Description,
		Date:        response.Date.Format("2006-01-02 15:04:05"),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		return
	}

	responses, err := h.ledgerService.ListTransactions(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		return
	}

	apiResponses := make([]TransactionResponse, len(responses))
	for i, response := range responses {
		apiResponses[i] = TransactionResponse{
			ID:          response.ID,
			Amount:      response.Amount,
			Category:    response.Category,
			Description: response.Description,
			Date:        response.Date.Format("2006-01-02 15:04:05"),
		}
	}

	json.NewEncoder(w).Encode(apiResponses)
}

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if r.Context().Err() != nil {
		return
	}

	// Преобразуем в доменный DTO
	domainReq := domain.CreateBudgetRequest{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   req.Period,
	}

	response, err := h.ledgerService.CreateBudget(r.Context(), domainReq)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	apiResponse := BudgetResponse{
		Category: response.Category,
		Limit:    response.Limit,
		Period:   response.Period,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		return
	}

	responses, err := h.ledgerService.ListBudgets(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		return
	}

	apiResponses := make([]BudgetResponse, len(responses))
	for i, response := range responses {
		apiResponses[i] = BudgetResponse{
			Category: response.Category,
			Limit:    response.Limit,
			Period:   response.Period,
		}
	}

	json.NewEncoder(w).Encode(apiResponses)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	// Проверяем, не отменен ли уже контекст
	if r.Context().Err() != nil {
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		return
	}

	if err := h.ledgerService.HealthCheck(r.Context()); err != nil {
		http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	errorMsg := err.Error()

	switch errorMsg {
	case "budget not found":
		http.Error(w, `{"error":"budget not found"}`, http.StatusBadRequest)
	case "budget exceeded":
		http.Error(w, `{"error":"budget exceeded"}`, http.StatusConflict)
	case "validation failed: amount must be positive",
		"validation failed: category is required":
		http.Error(w, `{"error":"`+errorMsg+`"}`, http.StatusBadRequest)
	default:
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
	}
}

func (h *Handler) TimeoutTest(w http.ResponseWriter, r *http.Request) {
	time.Sleep(10 * time.Second)

	json.NewEncoder(w).Encode(map[string]string{"status": "completed after 10s"})
}

func (h *Handler) GetSpendingSummary(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		http.Error(w, `{"error":"both from and to parameters are required"}`, http.StatusBadRequest)
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		http.Error(w, `{"error":"invalid from date format, expected YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		http.Error(w, `{"error":"invalid to date format, expected YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	req := domain.GetSpendingSummaryRequest{
		From: from,
		To:   to,
	}

	summary, err := h.ledgerService.GetSpendingSummary(r.Context(), req)
	if err != nil {
		h.handleReportServiceError(w, err)
		return
	}

	if summary == nil {
		summary = make(domain.SpendingSummary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *Handler) handleReportServiceError(w http.ResponseWriter, err error) {
	errorMsg := err.Error()

	switch {
	case strings.Contains(errorMsg, "dates are required"),
		strings.Contains(errorMsg, "from date cannot be after to date"),
		strings.Contains(errorMsg, "period cannot exceed"):
		http.Error(w, `{"error":"`+errorMsg+`"}`, http.StatusBadRequest)
	default:
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
	}
}

func (h *Handler) CreateTransactionsBulk(w http.ResponseWriter, r *http.Request) {
	if r.Context().Err() != nil {
		h.handleTimeoutError(w, r.Context().Err())
		return
	}

	workers := 4
	if workersStr := r.URL.Query().Get("workers"); workersStr != "" {
		if w, err := strconv.Atoi(workersStr); err == nil && w > 0 {
			workers = w
		}
	}

	var req struct {
		Transactions []CreateTransactionRequest `json:"transactions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if len(req.Transactions) > 1000 {
		http.Error(w, `{"error":"Maximum 1000 transactions per request"}`, http.StatusBadRequest)
		return
	}

	bulkReq := domain.BulkTransactionRequest{
		Transactions: make([]domain.CreateTransactionRequest, len(req.Transactions)),
	}

	for i, tx := range req.Transactions {
		domainReq := domain.CreateTransactionRequest{
			Amount:      tx.Amount,
			Category:    tx.Category,
			Description: tx.Description,
		}

		if tx.Date != "" {
			if date, err := time.Parse("2006-01-02 15:04:05", tx.Date); err == nil {
				domainReq.Date = date
			} else if date, err := time.Parse("2006-01-02", tx.Date); err == nil {
				domainReq.Date = date
			}
		}

		bulkReq.Transactions[i] = domainReq
	}

	response, err := h.ledgerService.CreateTransactionsBulk(r.Context(), bulkReq, workers)

	if err != nil {
		if err == context.DeadlineExceeded {
			h.handleTimeoutError(w, err)
		} else {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	apiResponse := BulkTransactionResponse{
		Total:    response.Total,
		Accepted: response.Accepted,
		Rejected: response.Rejected,
		Errors:   make([]BulkTransactionResult, len(response.Errors)),
	}

	for i, err := range response.Errors {
		apiResponse.Errors[i] = BulkTransactionResult{
			Index: err.Index,
			Error: err.Error,
		}
	}

	json.NewEncoder(w).Encode(apiResponse)
}

func (h *Handler) handleTimeoutError(w http.ResponseWriter, err error) {
	if err == context.DeadlineExceeded {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Request timeout",
		})
	} else if err == context.Canceled {
		w.Header().Set("Content-Type", "/json")
		w.WriteHeader(499) // Client Closed Request
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Client closed request",
		})
	}
}
