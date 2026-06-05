package models

// Book статус
const (
	StatusAvailable = "Available"
	StatusIssued    = "Issued"
)

// Book представляет книгу в библиотеке.
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Year   int    `json:"year"`
	Status string `json:"status"`
}

// User представляет читателя библиотеки.
type User struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	RegistrationDate string `json:"registration_date"`
}

// Issue представляет факт выдачи книги читателю.
type Issue struct {
	ID         string  `json:"id"`
	BookID     string  `json:"book_id"`
	UserID     string  `json:"user_id"`
	IssueDate  string  `json:"issue_date"`
	DueDate    string  `json:"due_date"`
	ReturnDate *string `json:"return_date,omitempty"`
}

// --- Request / Response DTO ---

type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Year   int    `json:"year"`
}

type UpdateBookRequest struct {
	Title  *string `json:"title"`
	Author *string `json:"author"`
	ISBN   *string `json:"isbn"`
	Year   *int    `json:"year"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type IssueBookRequest struct {
	BookID string `json:"book_id"`
	UserID string `json:"user_id"`
}

type ReturnBookRequest struct {
	BookID string `json:"book_id"`
}
