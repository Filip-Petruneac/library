package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(dbService *TestDBService) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/books", GetAllBooks(dbService.DB)).Methods("GET")
	r.HandleFunc("/search_books", SearchBooks(dbService.DB)).Methods("GET")
	r.HandleFunc("/search_authors", SearchAuthors(dbService.DB)).Methods("GET")
	r.HandleFunc("/authors", GetAuthors(dbService.DB)).Methods("GET")
	return r
}
// Test for GetAllBooks handler
func TestSearchBooks(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Amintiri din copilărie", 1, "amintiri.jpg", false, "A novel by Ion Creangă", "Creangă", "Ion").
		AddRow(2, "Chemarea străbunilor", 2, "chemarea.jpg", true, "A novel by Jack London", "London", "Jack")

	dbService.Mock.ExpectQuery(`SELECT books.id AS book_id, books.title AS book_title, books.author_id AS author_id, books.photo AS book_photo, books.is_borrowed AS is_borrowed, books.details AS book_details, authors.Lastname AS author_lastname, authors.Firstname AS author_firstname FROM books JOIN authors ON books.author_id = authors.id WHERE books.title LIKE \? OR authors.Firstname LIKE \? OR authors.Lastname LIKE \?`).
		WithArgs("%Amintiri%", "%Amintiri%", "%Amintiri%").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/search_books?query=Amintiri", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBooks := []BookAuthorInfo{
		{BookID: 1, BookTitle: "Amintiri din copilărie", AuthorID: 1, BookPhoto: "amintiri.jpg", IsBorrowed: false, BookDetails: "A novel by Ion Creangă", AuthorLastname: "Creangă", AuthorFirstname: "Ion"},
		{BookID: 2, BookTitle: "Chemarea străbunilor", AuthorID: 2, BookPhoto: "chemarea.jpg", IsBorrowed: true, BookDetails: "A novel by Jack London", AuthorLastname: "London", AuthorFirstname: "Jack"},
	}
	var actualBooks []BookAuthorInfo
	err = json.NewDecoder(rr.Body).Decode(&actualBooks)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if !reflect.DeepEqual(expectedBooks, actualBooks) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedBooks, actualBooks)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for SearchAuthors handler
func TestSearchAuthors(t *testing.T) {
    dbService, err := NewTestDBService()
    if err != nil {
        t.Fatalf("An error '%s' was not expected when creating the test DB service", err)
    }
    defer dbService.DB.Close()

    columns := []string{"id", "Firstname", "Lastname", "photo"}
    rows := sqlmock.NewRows(columns).
        AddRow(1, "Ion", "Creangă", "ion.jpg").
        AddRow(2, "Jack", "London", "jack.jpg")

    dbService.Mock.ExpectQuery(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \? OR Lastname LIKE \?`).
        WithArgs("%Ion%", "%Ion%").
        WillReturnRows(rows)

    req, _ := http.NewRequest("GET", "/search_authors?query=Ion", nil)
    rr := httptest.NewRecorder()
    setupTestRouter(dbService).ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    expectedAuthors := []AuthorInfo{
        {ID: 1, Firstname: "Ion", Lastname: "Creangă", Photo: "ion.jpg"},
        {ID: 2, Firstname: "Jack", Lastname: "London", Photo: "jack.jpg"},
    }
    var actualAuthors []AuthorInfo
    err = json.NewDecoder(rr.Body).Decode(&actualAuthors)
    if err != nil {
        t.Fatalf("Failed to decode response body: %v", err)
    }
    assert.ElementsMatch(t, expectedAuthors, actualAuthors)

    err = dbService.Mock.ExpectationsWereMet()
    assert.NoError(t, err)
}

// Test for GetAuthors handler
func TestGetAuthors(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when creating the test DB service", err)
	}
	defer dbService.DB.Close()

	columns := []string{"id", "lastname", "firstname", "photo"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Creangă", "Ion", "ion.jpg").
		AddRow(2, "London", "Jack", "jack.jpg")

	dbService.Mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/authors", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedAuthors := []Author{
		{ID: 1, Lastname: "Creangă", Firstname: "Ion", Photo: "ion.jpg"},
		{ID: 2, Lastname: "London", Firstname: "Jack", Photo: "jack.jpg"},
	}
	var actualAuthors []Author
	err = json.NewDecoder(rr.Body).Decode(&actualAuthors)
	assert.NoError(t, err)
	assert.Equal(t, expectedAuthors, actualAuthors)

	err = dbService.Mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
