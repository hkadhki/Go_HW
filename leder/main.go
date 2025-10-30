package main

import (
	"errors"
	"fmt"
	"log"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

func main() {
	fmt.Println("Ledger service started")

	err := AddTransaction(Transaction{
		Amount:      1,
		Category:    "aa1",
		Description: "aaaaaaaaaaaaaaaaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}

	err = AddTransaction(Transaction{
		Amount:      1,
		Category:    "aa2",
		Description: "aaaaaaaaaaaaaaaaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}

	err = AddTransaction(Transaction{
		Amount:      1,
		Category:    "aa3",
		Description: "aaaaaaaaaaaaaaaaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}

	fmt.Println(ListTransactions())

}

func AddTransaction(transaction Transaction) error {
	if transaction.Amount == 0 {
		return errors.New("Invalid amount")
	}

	transaction.ID = len(transactions)

	transactions = append(transactions, transaction)

	return nil
}

func ListTransactions() []Transaction {
	result := make([]Transaction, len(transactions))
	copy(result, transactions)
	return result
}
