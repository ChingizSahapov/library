package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"library/internal/models"
)

// BooksHandler обрабатывает запросы, связанные с книгами.
type BooksHandler struct {
	db *sql.DB
}

func NewBooksHandler(db *sql.DB) *BooksHandler {
	return &BooksHandler{db: db}
}

// POST /books
func (h *BooksHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Author == "" || req.ISBN == "" || req.Year == 0 {
		writeError(w, http.StatusBadRequest, "title, author, isbn and year are required")
		return
	}

	book := models.Book{
		ID:     uuid.NewString(),
		Title:  req.Title,
		Author: req.Author,
		ISBN:   req.ISBN,
		Year:   req.Year,
		Status: models.StatusAvailable,
	}

	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO books (id, title, author, isbn, year, status) VALUES (?,?,?,?,?,?)`,
		book.ID, book.Title, book.Author, book.ISBN, book.Year, book.Status,
	)
	if err != nil {
		slog.Error("create book", "err", err)
		writeError(w, http.StatusConflict, "book with this ISBN already exists or DB error")
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

// GET /books
func (h *BooksHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	author := q.Get("author")
	status := q.Get("status")

	pageStr := q.Get("page")
	limitStr := q.Get("limit")
	page := 1
	limit := 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit

	query := `SELECT id, title, author, isbn, year, status FROM books WHERE 1=1`
	args := []any{}
	if author != "" {
		query += ` AND author LIKE ?`
		args = append(args, "%"+author+"%")
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		slog.Error("list books", "err", err)
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
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  books,
		"page":  page,
		"limit": limit,
	})
}

// GET /books/{id}
func (h *BooksHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b models.Book
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, title, author, isbn, year, status FROM books WHERE id = ?`, id,
	).Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	if err != nil {
		slog.Error("get book", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch book")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// PUT /books/{id}
func (h *BooksHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var existing models.Book
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, title, author, isbn, year, status FROM books WHERE id = ?`, id,
	).Scan(&existing.ID, &existing.Title, &existing.Author, &existing.ISBN, &existing.Year, &existing.Status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	if err != nil {
		slog.Error("update book fetch", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch book")
		return
	}

	var req models.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Author != nil {
		existing.Author = *req.Author
	}
	if req.ISBN != nil {
		existing.ISBN = *req.ISBN
	}
	if req.Year != nil {
		existing.Year = *req.Year
	}

	_, err = h.db.ExecContext(r.Context(),
		`UPDATE books SET title=?, author=?, isbn=?, year=? WHERE id=?`,
		existing.Title, existing.Author, existing.ISBN, existing.Year, id,
	)
	if err != nil {
		slog.Error("update book", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update book")
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// nowUTC возвращает текущее время в формате RFC3339.
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
