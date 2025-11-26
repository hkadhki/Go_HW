package api

import (
	"encoding/json"
	"ledger/domain"
	"ledger/service"
	"net/http"
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

	// Преобразуем в доменный DTO
	domainReq := domain.CreateTransactionRequest{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
	}

	// Парсим дату, если указана
	if req.Date != "" {
		if date, err := time.Parse("2006-01-02 15:04:05", req.Date); err == nil {
			domainReq.Date = date
		} else if date, err := time.Parse("2006-01-02", req.Date); err == nil {
			domainReq.Date = date
		}
	}

	// Вызываем метод сервиса
	response, err := h.ledgerService.CreateTransaction(r.Context(), domainReq)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Преобразуем ответ обратно в API DTO
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
	responses, err := h.ledgerService.ListTransactions(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		return
	}

	// Преобразуем в API DTO
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

	// Преобразуем в доменный DTO
	domainReq := domain.CreateBudgetRequest{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   req.Period,
	}

	// Вызываем метод сервиса
	response, err := h.ledgerService.CreateBudget(r.Context(), domainReq)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Преобразуем ответ обратно в API DTO
	apiResponse := BudgetResponse{
		Category: response.Category,
		Limit:    response.Limit,
		Period:   response.Period,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	responses, err := h.ledgerService.ListBudgets(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		return
	}

	// Преобразуем в API DTO
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
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
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
