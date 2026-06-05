package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"library/internal/config"
	"library/internal/db"
	"library/internal/handlers"
	"library/internal/middleware"
)

func main() {
	// Структурированный JSON-логгер.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	booksH := handlers.NewBooksHandler(database)
	usersH := handlers.NewUsersHandler(database)
	issuesH := handlers.NewIssuesHandler(database)

	r := chi.NewRouter()

	// Глобальные middleware.
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)

	// Книги.
	r.Post("/books", booksH.Create)
	r.Get("/books", booksH.List)
	r.Get("/books/{id}", booksH.Get)
	r.Put("/books/{id}", booksH.Update)

	// Читатели.
	r.Post("/users", usersH.Create)
	r.Get("/users/{id}/books", usersH.ActiveBooks)

	// Выдача / возврат.
	r.Post("/issues", issuesH.Issue)
	r.Post("/returns", issuesH.Return)

	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
