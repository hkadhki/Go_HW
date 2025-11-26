package pg

import (
	"context"
	"database/sql"
	"fmt"
	domain2 "ledger/domain"
	"time"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain2.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction domain2.Transaction) (int, error) {
	var id int

	// Если дата не установлена, используем текущее время
	date := transaction.Date
	if date.IsZero() {
		date = time.Now()
	}

	query := `
		INSERT INTO expenses (amount, category, description, date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		transaction.Amount,
		transaction.Category,
		transaction.Description,
		date.Format("2006-01-02 15:04:05"),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return id, nil
}

func (r *transactionRepository) List(ctx context.Context) ([]domain2.Transaction, error) {
	query := `
		SELECT id, amount, category, description, date 
		FROM expenses 
		ORDER BY date DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []domain2.Transaction
	for rows.Next() {
		var tx domain2.Transaction
		var dateStr string

		err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		// Парсим дату из строки
		if date, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
			tx.Date = date
		} else {
			// Если парсинг не удался, используем текущее время
			tx.Date = time.Now()
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) GetTotalByCategory(ctx context.Context, category string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE category = $1
	`

	var total float64
	err := r.db.QueryRowContext(ctx, query, category).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total by category: %w", err)
	}

	return total, nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int) (*domain2.Transaction, error) {
	query := `
		SELECT id, amount, category, description, date 
		FROM expenses 
		WHERE id = $1
	`

	var tx domain2.Transaction
	var dateStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &dateStr,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by id: %w", err)
	}

	// Парсим дату
	if date, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		tx.Date = date
	} else {
		tx.Date = time.Now()
	}

	return &tx, nil
}
