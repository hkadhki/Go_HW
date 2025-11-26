package app

import (
	"context"
	"database/sql"
	"fmt"
	pg2 "ledger/repository/pg"
	service2 "ledger/service"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	Service service2.LedgerService
	closeFn func() error
}

func (a *App) Close() error {
	if a.closeFn != nil {
		return a.closeFn()
	}
	return nil
}

func New(ctx context.Context) (*App, error) {
	config := LoadConfig()

	db, err := initDatabase(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	transactionRepo := pg2.NewTransactionRepository(db)
	budgetRepo := pg2.NewBudgetRepository(db)

	ledgerService := service2.NewLedgerService(transactionRepo, budgetRepo)

	closeFn := func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		return nil
	}

	return &App{
		Service: ledgerService,
		closeFn: closeFn,
	}, nil
}

func initDatabase(ctx context.Context, config *Config) (*sql.DB, error) {
	dsn := config.DSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
