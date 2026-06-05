package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"library/internal/models"
)

// UsersHandler обрабатывает запросы, связанные с читателями.
type UsersHandler struct {
	db *sql.DB
}

func NewUsersHandler(db *sql.DB) *UsersHandler {
	return &UsersHandler{db: db}
}

// POST /users
func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}

	user := models.User{
		ID:               uuid.NewString(),
		Name:             req.Name,
		Email:            req.Email,
		RegistrationDate: nowUTC(),
	}

	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO users (id, name, email, registration_date) VALUES (?,?,?,?)`,
		user.ID, user.Name, user.Email, user.RegistrationDate,
	)
	if err != nil {
		slog.Error("create user", "err", err)
		writeError(w, http.StatusConflict, "user with this email already exists or DB error")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// GET /users/{id}/books — книги, которые сейчас на руках у читателя.
func (h *UsersHandler) ActiveBooks(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Проверяем, что пользователь существует.
	var exists bool
	err := h.db.QueryRowContext(r.Context(), `SELECT 1 FROM users WHERE id=?`, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		slog.Error("check user", "err", err)
		writeError(w, http.StatusInternalServerError, "DB error")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT b.id, b.title, b.author, b.isbn, b.year, b.status
		FROM books b
		INNER JOIN issues i ON i.book_id = b.id
		WHERE i.user_id = ? AND i.return_date IS NULL
	`, userID)
	if err != nil {
		slog.Error("user active books", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status); err != nil {
			slog.Error("scan book", "err", err)
			continue
		}
		books = append(books, b)
	}
	writeJSON(w, http.StatusOK, books)
}
