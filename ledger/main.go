package main

import (
	"ledger/db"
	"ledger/models"
	"ledger/services"
	"log"
)

func main() {
	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Ledger service started with PostgreSQL storage")

	ledgerService := services.NewLedgerService()

	// Set initial budgets
	initialBudgets := []models.Budget{
		{Category: "Продукты", Limit: 5000, Period: "monthly"},
		{Category: "Транспорт", Limit: 2000, Period: "monthly"},
		{Category: "Развлечения", Limit: 3000, Period: "monthly"},
	}

	for _, budget := range initialBudgets {
		if err := ledgerService.SetBudget(budget); err != nil {
			log.Printf("Ошибка установки бюджета для категории %s: %v", budget.Category, err)
		} else {
			log.Printf("Бюджет установлен для категории: %s", budget.Category)
		}
	}

	// Add test transaction within budget
	_, err := ledgerService.AddTransaction(models.Transaction{
		Amount:      1000,
		Category:    "Продукты",
		Description: "Покупки в супермаркете",
		Date:        "2024-01-15 10:00:00",
	})
	if err != nil {
		log.Printf("Ошибка добавления транзакции: %v", err)
	} else {
		log.Println("Транзакция успешно добавлена")
	}

	// This should fail - exceeds budget
	_, err = ledgerService.AddTransaction(models.Transaction{
		Amount:      3000,
		Category:    "Транспорт",
		Description: "Такси",
		Date:        "2024-01-15 11:00:00",
	})
	if err != nil {
		log.Printf("Ожидаемая ошибка (превышение бюджета): %v", err)
	} else {
		log.Println("Транзакция добавлена (не должно было произойти)")
	}

	// List transactions to verify
	transactions, err := ledgerService.ListTransactions()
	if err != nil {
		log.Printf("Ошибка получения списка транзакций: %v", err)
	} else {
		log.Printf("Всего транзакций: %d", len(transactions))
		for _, tx := range transactions {
			log.Printf("Транзакция: %+v", tx)
		}
	}

	// List budgets to verify
	budgets, err := ledgerService.ListBudgets()
	if err != nil {
		log.Printf("Ошибка получения списка бюджетов: %v", err)
	} else {
		log.Printf("Всего бюджетов: %d", len(budgets))
		for _, budget := range budgets {
			log.Printf("Бюджет: %+v", budget)
		}
	}

	log.Println("Ledger service is running...")

	// Keep the service running
	select {}
}
