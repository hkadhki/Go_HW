package main

import (
	"ledger/models"
	"ledger/services"
	"log"
	"os"
)

func main() {
	log.Printf("Ledger service started")

	ledgerService := services.NewLedgerService()

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

		if err := ledgerService.LoadBudgets(file); err != nil {
			log.Printf("Ошибка загрузки бюджетов: %v", err)
		}
	}

	// успешный тест
	_, err = ledgerService.AddTransaction(models.Transaction{
		Amount:      1,
		Category:    "Продукты",
		Description: "aaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}

	// ожидается ошибка
	_, err = ledgerService.AddTransaction(models.Transaction{
		Amount:      1000000,
		Category:    "Транспорт",
		Description: "aaa",
		Date:        "2022-02-22 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	}
}
