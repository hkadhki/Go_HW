## Запуск сервисов



```
cd gateway
go run main.go
cd ledger
go run main.go 
```

### Создание/обновление бюджета

```
curl -X POST http://localhost:8080/api/budgets \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Продукты",
    "limit": 5000,
    "period": "2024-01"
  }' 
 ```

### Получение всех бюджетов

``` 
curl http://localhost:8080/api/budgets
```

#### Ошибка валидации (400)

``` 
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": -100, "category": "Продукты", "description": "Отрицательная сумма"}'
```
#### Пустая категория (400)

```
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": 500, "category": "", "description": "Пустая категория"}'
```

#### Превышение бюджета (409)
```
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": 4000, "category": "Продукты", "description": "Превышение бюджета"}'
```

### Создание тестовой транзакции

``` 
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": 500, "category": "Тест", "description": "Тестовая транзакция"}'
```

### Вывод транзакций
``` 
curl http://localhost:8080/api/transactions
```

#### Ошибка валидации (400)

``` 
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": -100, "category": "Продукты", "description": "Отрицательная сумма"}'
```

#### Превышение бюджета (409)
``` 
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount": 400000000, "category": "Продукты", "description": "Превышение бюджета"}'
```
