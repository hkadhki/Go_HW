package pg

import (
	"context"
	"database/sql"
	"fmt"
	"ledger/domain"
	"time"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction domain.Transaction) (int, error) {
	var id int

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

func (r *transactionRepository) List(ctx context.Context) ([]domain.Transaction, error) {
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

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var dateStr string

		err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if date, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
			tx.Date = date
		} else {
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

func (r *transactionRepository) GetByID(ctx context.Context, id int) (*domain.Transaction, error) {
	query := `
		SELECT id, amount, category, description, date 
		FROM expenses 
		WHERE id = $1
	`

	var tx domain.Transaction
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

	if date, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		tx.Date = date
	} else {
		tx.Date = time.Now()
	}

	return &tx, nil
}

func (r *transactionRepository) GetSpendingByPeriod(ctx context.Context, from, to time.Time) (domain.SpendingSummary, error) {
	query := `
		SELECT category, COALESCE(SUM(amount), 0) as total
		FROM expenses 
		WHERE date BETWEEN $1 AND $2
		GROUP BY category
		ORDER BY total DESC
	`

	rows, err := r.db.QueryContext(ctx, query,
		from.Format("2006-01-02 15:04:05"),
		to.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query spending by period: %w", err)
	}
	defer rows.Close()

	summary := make(domain.SpendingSummary)
	for rows.Next() {
		var category string
		var total float64

		err := rows.Scan(&category, &total)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spending: %w", err)
		}

		summary[category] = total
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating spending data: %w", err)
	}

	return summary, nil
}

func (r *transactionRepository) GetSpendingByCategoryAndPeriod(ctx context.Context, category string, from, to time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE category = $1 AND date BETWEEN $2 AND $3
	`

	var total float64
	err := r.db.QueryRowContext(ctx, query,
		category,
		from.Format("2006-01-02 15:04:05"),
		to.Format("2006-01-02 15:04:05"),
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("failed to get spending for category %s: %w", category, err)
	}

	return total, nil
}
