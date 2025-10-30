package models

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
	Date        string
}

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("сумма транзакции должна быть положительным числом")
	}
	if strings.TrimSpace(t.Category) == "" {
		return errors.New("категория транзакции не может быть пустой")
	}
	if t.Date != "" {
		if _, err := time.Parse("2006-01-02 15:04:05", t.Date); err != nil {
			return errors.New("некорректный формат даты")
		}
	}
	return nil
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

func (b Budget) Validate() error {
	if strings.TrimSpace(b.Category) == "" {
		return errors.New("категория бюджета не может быть пустой")
	}
	if b.Limit <= 0 {
		return errors.New("лимит бюджета должен быть положительным числом")
	}
	return nil
}

func CheckValid(v Validatable) error {
	if err := v.Validate(); err != nil {
		return errors.New("валидация не пройдена")
	}
	return nil
}
