package domain

import (
	"errors"
	"strings"
	"time"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("сумма транзакции должна быть положительным числом")
	}
	if strings.TrimSpace(t.Category) == "" {
		return errors.New("категория транзакции не может быть пустой")
	}
	if t.Date.IsZero() {
		return errors.New("дата транзакции должна быть указана")
	}
	return nil
}

type Budget struct {
	Category string
	Limit    float64
	Period   string // "monthly", "weekly", "daily"
}

func (b Budget) Validate() error {
	if strings.TrimSpace(b.Category) == "" {
		return errors.New("категория бюджета не может быть пустой")
	}
	if b.Limit <= 0 {
		return errors.New("лимит бюджета должен быть положительным числом")
	}
	if b.Period != "" && b.Period != "monthly" && b.Period != "weekly" && b.Period != "daily" {
		return errors.New("период должен быть 'monthly', 'weekly', 'daily' или пустым")
	}
	return nil
}
