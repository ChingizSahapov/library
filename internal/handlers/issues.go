package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"library/internal/models"
)

// IssuesHandler обрабатывает выдачу и возврат книг.
type IssuesHandler struct {
	db *sql.DB
}

func NewIssuesHandler(db *sql.DB) *IssuesHandler {
	return &IssuesHandler{db: db}
}

// POST /issues — выдать книгу читателю.
func (h *IssuesHandler) Issue(w http.ResponseWriter, r *http.Request) {
	var req models.IssueBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BookID == "" || req.UserID == "" {
		writeError(w, http.StatusBadRequest, "book_id and user_id are required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		slog.Error("begin tx", "err", err)
		writeError(w, http.StatusInternalServerError, "transaction error")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	// Проверяем статус книги (с блокировкой строки внутри транзакции).
	var status string
	err = tx.QueryRowContext(r.Context(),
		`SELECT status FROM books WHERE id = ?`, req.BookID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	if err != nil {
		slog.Error("check book status", "err", err)
		writeError(w, http.StatusInternalServerError, "DB error")
		return
	}
	if status != models.StatusAvailable {
		writeError(w, http.StatusConflict, "book is already issued")
		return
	}

	// Проверяем существование читателя.
	var userExists bool
	err = tx.QueryRowContext(r.Context(),
		`SELECT 1 FROM users WHERE id = ?`, req.UserID,
	).Scan(&userExists)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		slog.Error("check user", "err", err)
		writeError(w, http.StatusInternalServerError, "DB error")
		return
	}

	now := time.Now().UTC()
	issue := models.Issue{
		ID:        uuid.NewString(),
		BookID:    req.BookID,
		UserID:    req.UserID,
		IssueDate: now.Format(time.RFC3339),
		DueDate:   now.Add(14 * 24 * time.Hour).Format(time.RFC3339),
	}

	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO issues (id, book_id, user_id, issue_date, due_date) VALUES (?,?,?,?,?)`,
		issue.ID, issue.BookID, issue.UserID, issue.IssueDate, issue.DueDate,
	)
	if err != nil {
		slog.Error("insert issue", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create issue record")
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE books SET status = ? WHERE id = ?`, models.StatusIssued, req.BookID,
	)
	if err != nil {
		slog.Error("update book status", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update book status")
		return
	}

	if err = tx.Commit(); err != nil {
		slog.Error("commit tx", "err", err)
		writeError(w, http.StatusInternalServerError, "transaction commit error")
		return
	}

	writeJSON(w, http.StatusCreated, issue)
}

// POST /returns — вернуть книгу.
func (h *IssuesHandler) Return(w http.ResponseWriter, r *http.Request) {
	var req models.ReturnBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BookID == "" {
		writeError(w, http.StatusBadRequest, "book_id is required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		slog.Error("begin tx", "err", err)
		writeError(w, http.StatusInternalServerError, "transaction error")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	// Находим активную выдачу (без return_date).
	var issueID string
	err = tx.QueryRowContext(r.Context(),
		`SELECT id FROM issues WHERE book_id = ? AND return_date IS NULL`, req.BookID,
	).Scan(&issueID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no active issue found for this book")
		return
	}
	if err != nil {
		slog.Error("find issue", "err", err)
		writeError(w, http.StatusInternalServerError, "DB error")
		return
	}

	returnDate := time.Now().UTC().Format(time.RFC3339)

	_, err = tx.ExecContext(r.Context(),
		`UPDATE issues SET return_date = ? WHERE id = ?`, returnDate, issueID,
	)
	if err != nil {
		slog.Error("update issue return_date", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update issue")
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE books SET status = ? WHERE id = ?`, models.StatusAvailable, req.BookID,
	)
	if err != nil {
		slog.Error("update book status", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update book status")
		return
	}

	if err = tx.Commit(); err != nil {
		slog.Error("commit tx", "err", err)
		writeError(w, http.StatusInternalServerError, "transaction commit error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"issue_id":    issueID,
		"return_date": returnDate,
		"message":     "book returned successfully",
	})
}

// decodeJSON — вспомогательная функция декодирования тела запроса.
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
