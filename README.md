# Библиотека — RESTful API на Go

Бэкенд для системы учёта книг городской библиотеки. Написан на Go 1.21+, использует SQLite в качестве базы данных.

## Стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.21+ |
| Роутер | [chi v5](https://github.com/go-chi/chi) |
| БД | SQLite (`mattn/go-sqlite3`) |
| UUID | `google/uuid` |
| .env | `joho/godotenv` |
| Логи | `log/slog` (JSON) |

## Структура проекта

```
library/
├── cmd/
│   └── server/
│       └── main.go          # точка входа
├── internal/
│   ├── config/
│   │   └── config.go        # чтение переменных окружения
│   ├── db/
│   │   └── db.go            # подключение к SQLite + миграции
│   ├── handlers/
│   │   ├── books.go         # CRUD книг
│   │   ├── users.go         # регистрация, список книг читателя
│   │   └── issues.go        # выдача / возврат
│   ├── middleware/
│   │   └── logger.go        # структурированное логирование запросов
│   └── models/
│       └── models.go        # структуры данных и DTO
├── migrations/
│   └── 001_init.sql         # исходная схема БД (справочно)
├── .env.example
├── go.mod
└── README.md
```

## Запуск

### 1. Зависимости

```bash
# CGO нужен для go-sqlite3
# Linux / macOS: обычно всё есть
# Windows: нужен MinGW (gcc)
go mod tidy
```

### 2. Конфигурация

```bash
cp .env.example .env
# при необходимости отредактируйте .env
```

### 3. Запуск сервера

```bash
go run ./cmd/server
```

Или с явными переменными окружения:

```bash
SERVER_PORT=9000 DB_PATH=./data.db go run ./cmd/server
```

### 4. Сборка бинарника

```bash
go build -o library ./cmd/server
./library
```

---

## API — примеры curl

### Книги

**Добавить книгу**
```bash
curl -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Мастер и Маргарита","author":"Михаил Булгаков","isbn":"978-5-17-090000-1","year":1967}' | jq
```

**Список книг (с пагинацией и фильтрами)**
```bash
# Все книги (страница 1, по 20 шт.)
curl -s "http://localhost:8080/books?page=1&limit=20" | jq

# Фильтр по автору
curl -s "http://localhost:8080/books?author=Булгаков" | jq

# Фильтр по статусу
curl -s "http://localhost:8080/books?status=Available" | jq
```

**Получить книгу по ID**
```bash
curl -s http://localhost:8080/books/<BOOK_ID> | jq
```

**Обновить книгу**
```bash
curl -s -X PUT http://localhost:8080/books/<BOOK_ID> \
  -H "Content-Type: application/json" \
  -d '{"year":1966}' | jq
```

---

### Читатели

**Зарегистрировать читателя**
```bash
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван Иванов","email":"ivan@example.com"}' | jq
```

**Книги, которые сейчас на руках у читателя**
```bash
curl -s http://localhost:8080/users/<USER_ID>/books | jq
```

---

### Выдача и возврат

**Выдать книгу читателю**
```bash
curl -s -X POST http://localhost:8080/issues \
  -H "Content-Type: application/json" \
  -d '{"book_id":"<BOOK_ID>","user_id":"<USER_ID>"}' | jq
```

> Книга должна иметь статус `Available`. Срок возврата — 14 дней.

**Вернуть книгу**
```bash
curl -s -X POST http://localhost:8080/returns \
  -H "Content-Type: application/json" \
  -d '{"book_id":"<BOOK_ID>"}' | jq
```

---

## Формат логов

Все логи пишутся в stdout в формате JSON:

```json
{"time":"2024-06-05T12:00:00Z","level":"INFO","msg":"request","method":"POST","path":"/books","status":201,"duration_ms":3,"remote_addr":"127.0.0.1:54321"}
```
