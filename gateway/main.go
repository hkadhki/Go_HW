package main

import (
	"fmt"
	"gateway/internal/api"
	"ledger/services"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	ledgerService := services.NewLedgerService()
	handler := api.NewHandler(ledgerService)

	r := mux.NewRouter()

	r.Use(api.JSONMiddleware)
	r.Use(api.LoggingMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/transactions", handler.CreateTransactionHandler).Methods("POST")
	api.HandleFunc("/transactions", handler.ListTransactions).Methods("GET")

	// Бюджеты
	api.HandleFunc("/budgets", handler.CreateBudget).Methods("POST")
	api.HandleFunc("/budgets", handler.ListBudgets).Methods("GET")

	// Health check
	r.HandleFunc("/ping", handler.Ping).Methods("GET")

	// Запуск сервера
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
