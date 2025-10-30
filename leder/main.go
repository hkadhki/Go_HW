package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

var transactions []Transaction

var budgets map[string]Budget
var budgetsAmount map[string]float64

func main() {
	fmt.Println("Ledger service started")

	budgets = make(map[string]Budget)
	budgetsAmount = make(map[string]float64)

	//initialBudgets := []Budget{
	//	{Category: "Продукты", Limit: 5000, Period: "2024-01"},
	//	{Category: "Транспорт", Limit: 2000, Period: "2024-01"},
	//	{Category: "Развлечения", Limit: 3000, Period: "2024-01"},
	//}
	//
	//for _, budget := range initialBudgets {
	//	if err := SetBudget(budget); err != nil {
	//		log.Printf("Ошибка установки бюджета: %v", err)
	//	}
	//}

	file, err := os.Open("budgets.json")
	if err != nil {
		log.Printf("Ошибка открытия файла бюджетов: %v", err)
	} else {
		defer file.Close()

		if err := LoadBudgets(file); err != nil {
			log.Printf("Ошибка загрузки бюджетов: %v", err)
		}
	}

	// успешный тест
	err = AddTransaction(Transaction{
		Amount:      1,
		Category:    "Продукты",
		Description: "aaaaaaaaaaaaaaaaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}

	// ожидается ошибка
	err = AddTransaction(Transaction{
		Amount:      1000000,
		Category:    "Транспорт",
		Description: "aaaaaaaaaaaaaaaaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}
}

func AddTransaction(transaction Transaction) error {
	if transaction.Amount == 0 {
		return errors.New("invalid amount")
	}

	budget, exists := budgets[transaction.Category]
	if !exists {
		return nil
	}

	if budgetsAmount[transaction.Category]+transaction.Amount > budget.Limit {
		return errors.New("budget exceeded")
	}

	transaction.ID = len(transactions)

	transactions = append(transactions, transaction)

	budgetsAmount[transaction.Category] += transaction.Amount
	
	fmt.Println(budgets)
	fmt.Println(budgetsAmount)
	fmt.Println(transactions)

	return nil
}

func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)
	return result
}

func SetBudget(b Budget) error {
	if b.Category == "" {
		return errors.New("категория бюджета не может быть пустой")
	}
	if b.Limit <= 0 {
		return errors.New("лимит бюджета должен быть положительным числом")
	}
	if b.Period == "" {
		b.Period = "monthly"
	}

	budgets[b.Category] = b

	return nil
}

func LoadBudgets(r io.Reader) error {
	reader := bufio.NewReader(r)

	data, err := io.ReadAll(reader)
	if err != nil {
		return errors.New("ошибка чтения")
	}

	var budgetList []Budget
	if err := json.Unmarshal(data, &budgetList); err != nil {
		return errors.New("ошибка парсинга в json")
	}

	for _, budget := range budgetList {
		if err := SetBudget(budget); err != nil {
			return errors.New("ошибка сохранения бюджета")
		}
	}

	return nil
}
