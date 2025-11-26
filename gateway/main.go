// main.go
package main

import (
	"context"
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
	// Инициализация приложения через фабрику
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app, err := app.New(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	log.Println("Application initialized successfully")

	// Создание HTTP обработчиков
	handler := api.NewHandler(app.Service)

	// Настройка маршрутов
	r := setupRouter(handler)

	// Запуск сервера
	startServer(r)
}

func setupRouter(handler *api.Handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(api.JSONMiddleware)
	r.Use(api.LoggingMiddleware)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/transactions", handler.CreateTransactionHandler).Methods("POST")
	apiRouter.HandleFunc("/transactions", handler.ListTransactions).Methods("GET")
	apiRouter.HandleFunc("/budgets", handler.CreateBudget).Methods("POST")
	apiRouter.HandleFunc("/budgets", handler.ListBudgets).Methods("GET")
	apiRouter.HandleFunc("/ping", handler.Ping).Methods("GET")
	apiRouter.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	return r
}

func startServer(r *mux.Router) {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
