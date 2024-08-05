package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(db *sql.DB) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/books", GetAllBooks(db)).Methods("GET")
	return r
}

func TestGetAllBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	columns := []string{"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Amintiri din copilărie", 1, "amintiri.jpg", false, "A novel by Ion Creangă", "Creangă", "Ion").
		AddRow(2, "Chemarea străbunilor", 2, "chemarea.jpg", true, "A novel by Jack London", "London", "Jack")

	mock.ExpectQuery(`SELECT books.id AS book_id, books.title AS book_title, books.author_id AS author_id, books.photo AS book_photo, books.is_borrowed AS is_borrowed, books.details AS book_details, authors.Lastname AS author_lastname, authors.Firstname AS author_firstname FROM books JOIN authors ON books.author_id = authors.id`).WillReturnRows(rows)

	r := setupTestRouter(db)
	req, _ := http.NewRequest("GET", "/books", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBooks := []BookAuthorInfo{
		{BookID: 1, BookTitle: "Amintiri din copilărie", AuthorID: 1, BookPhoto: "amintiri.jpg", IsBorrowed: false, BookDetails: "A novel by Ion Creangă", AuthorLastname: "Creangă", AuthorFirstname: "Ion"},
		{BookID: 2, BookTitle: "Chemarea străbunilor", AuthorID: 2, BookPhoto: "chemarea.jpg", IsBorrowed: true, BookDetails: "A novel by Jack London", AuthorLastname: "London", AuthorFirstname: "Jack"},
	}
	var actualBooks []BookAuthorInfo
	err = json.NewDecoder(rr.Body).Decode(&actualBooks)
	assert.NoError(t, err)
	assert.Equal(t, expectedBooks, actualBooks)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
