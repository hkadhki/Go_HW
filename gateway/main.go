package main

import (
	"context"
	"encoding/json"
	"gateway/internal/api"
	"ledger/app"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app, err := app.New(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	log.Println("Application initialized successfully")

	handler := api.NewHandler(app.Service)

	r := setupRouter(handler)

	startServer(r)
}

func setupRouter(handler *api.Handler) *mux.Router {
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()

	apiRouter.Use(api.TimeoutMiddleware()) // Таймаут 2 секунды (первым!)
	apiRouter.Use(api.JSONMiddleware)      // JSON responses
	apiRouter.Use(api.LoggingMiddleware)   // Логирование

	apiRouter.HandleFunc("/transactions", handler.CreateTransactionHandler).Methods("POST")
	apiRouter.HandleFunc("/transactions", handler.ListTransactions).Methods("GET")
	apiRouter.HandleFunc("/budgets", handler.CreateBudget).Methods("POST")
	apiRouter.HandleFunc("/budgets", handler.ListBudgets).Methods("GET")
	apiRouter.HandleFunc("/ping", handler.Ping).Methods("GET")
	apiRouter.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	apiRouter.HandleFunc("/timeout-test", handler.TimeoutTest).Methods("GET")
	apiRouter.HandleFunc("/reports/summary", handler.GetSpendingSummary).Methods("GET")
	apiRouter.HandleFunc("/transactions/bulk", handler.CreateTransactionsBulk).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	return r
}

func startServer(r *mux.Router) {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
