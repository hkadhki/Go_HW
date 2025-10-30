package api

import (
	"encoding/json"
	"ledger/models"
	"ledger/services"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	ledgerService *services.LedgerService
}

func NewHandler(ledgerService *services.LedgerService) *Handler {
	return &Handler{
		ledgerService: ledgerService,
	}
}

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (h *Handler) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	transaction := CreateRequestToTransaction(req)

	if err := models.CheckValid(transaction); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.ledgerService.AddTransaction(transaction)
	if err != nil {
		if err.Error() == "budget exceeded" {
			http.Error(w, `{"error":"budget exceeded"}`, http.StatusConflict)
		} else if err.Error() == "budget not found" {
			http.Error(w, `{"error":"budget not found"}`, http.StatusBadRequest)
		} else {
			http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := TransactionToResponse(transaction)
	response.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := h.ledgerService.ListTransactions()
	response := TransactionsToResponseList(transactions)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	budget := CreateRequestToBudget(req)

	if err := budget.Validate(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if err := h.ledgerService.SetBudget(budget); err != nil {
		http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(BudgetToResponse(budget))
	if err != nil {
		return
	}
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	budgets := h.ledgerService.ListBudgets()
	response := BudgetsToResponseList(budgets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}
