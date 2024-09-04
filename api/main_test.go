package main

import (
	"bytes"
	"mime/multipart"
	// "database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestSearchAuthors(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Test for missing query parameter
	req, err := http.NewRequest("GET", "/search_authors", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.SearchAuthors)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Query parameter is required\n", rr.Body.String())

	// Test for successful query with valid parameter
	mock.ExpectQuery("^SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \\? OR Lastname LIKE \\?$").
		WithArgs("%John%", "%John%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "Firstname", "Lastname", "photo"}).
			AddRow(1, "John", "Doe", "john_doe.jpg"))

	req, err = http.NewRequest("GET", "/search_authors?query=John", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody := `[{"id":1,"firstname":"John","lastname":"Doe","photo":"john_doe.jpg"}]`
	assert.JSONEq(t, expectedBody, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

func TestSearchBooks(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Test for missing query parameter
	req, err := http.NewRequest("GET", "/search_books", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.SearchBooks)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Query parameter is missing\n", rr.Body.String())

	// Test for successful query with valid parameter
	mock.ExpectQuery("^SELECT books.id AS book_id, books.title AS book_title, books.author_id AS author_id, books.photo AS book_photo, books.is_borrowed AS is_borrowed, books.details AS book_details, authors.Lastname AS author_lastname, authors.Firstname AS author_firstname FROM books JOIN authors ON books.author_id = authors.id WHERE books.title LIKE \\? OR authors.Firstname LIKE \\? OR authors.Lastname LIKE \\?$").
		WithArgs("%John%", "%John%", "%John%").
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname"}).
			AddRow(1, "Book Title", 1, "book.jpg", false, "Details", "Doe", "John"))

	req, err = http.NewRequest("GET", "/search_books?query=John", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody := `[{"book_id":1,"book_title":"Book Title","author_id":1,"book_photo":"book.jpg","is_borrowed":false,"book_details":"Details","author_lastname":"Doe","author_firstname":"John"}]`
	assert.JSONEq(t, expectedBody, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// createTestApp creates a test instance of the application with mocked dependencies.
func createTestApp(t *testing.T) (*App, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating sqlmock: %v", err)
	}

	logger := log.New(io.Discard, "", log.LstdFlags) // A logger that discards all output

	return &App{
		DB:     db,
		Logger: logger,
	}, mock
}

// TestGetAuthors tests the GetAuthors handler
func TestGetAuthors(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
		AddRow(1, "Doe", "John", "photo.jpg").
		AddRow(2, "Smith", "Jane", "photo2.jpg")

	mock.ExpectQuery("SELECT id, lastname, firstname, photo FROM authors").WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/authors", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetAuthors)
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var authors []Author
	err = json.NewDecoder(rr.Body).Decode(&authors)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 2, len(authors))
	assert.Equal(t, "Doe", authors[0].Lastname)
	assert.Equal(t, "Smith", authors[1].Lastname)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAllBooks tests the GetAllBooks handler
func TestGetAllBooks(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{
		"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
	}).AddRow(1, "Book Title 1", 1, "book1.jpg", false, "Details 1", "Doe", "John").
		AddRow(2, "Book Title 2", 2, "book2.jpg", true, "Details 2", "Smith", "Jane")

	mock.ExpectQuery("SELECT books.id AS book_id,").WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/books", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetAllBooks)
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var books []BookAuthorInfo
	err = json.NewDecoder(rr.Body).Decode(&books)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 2, len(books))
	assert.Equal(t, "Book Title 1", books[0].BookTitle)
	assert.Equal(t, "Book Title 2", books[1].BookTitle)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAuthorsAndBooks tests the GetAuthorsAndBooks handler
func TestGetAuthorsAndBooks(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "book_title", "book_photo"}).
		AddRow("John", "Doe", "Book Title 1", "book1.jpg").
		AddRow("Jane", "Smith", "Book Title 2", "book2.jpg")

	mock.ExpectQuery("SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo").
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/authorsbooks", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetAuthorsAndBooks)
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var authorsAndBooks []AuthorBook
	err = json.NewDecoder(rr.Body).Decode(&authorsAndBooks)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 2, len(authorsAndBooks))
	assert.Equal(t, "John", authorsAndBooks[0].AuthorFirstname)
	assert.Equal(t, "Doe", authorsAndBooks[0].AuthorLastname)
	assert.Equal(t, "Book Title 1", authorsAndBooks[0].BookTitle)
	assert.Equal(t, "book1.jpg", authorsAndBooks[0].BookPhoto)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAuthorBooksByID tests the GetAuthorBooksByID handler
func TestGetAuthorBooksByID(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	authorID := "1"

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "author_photo", "book_title", "book_photo"}).
		AddRow("John", "Doe", "john.jpg", "Book Title 1", "book1.jpg").
		AddRow("John", "Doe", "john.jpg", "Book Title 2", "book2.jpg")

	mock.ExpectQuery("SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, a.Photo AS author_photo, b.title AS book_title, b.photo AS book_photo").
		WithArgs(1).
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/authors/1", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetAuthorBooksByID)
	// Using mux.Vars to mock the ID parameter
	req = mux.SetURLVars(req, map[string]string{"id": authorID})
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var authorAndBooks struct {
		AuthorFirstname string       `json:"author_firstname"`
		AuthorLastname  string       `json:"author_lastname"`
		AuthorPhoto     string       `json:"author_photo"`
		Books           []AuthorBook `json:"books"`
	}
	err = json.NewDecoder(rr.Body).Decode(&authorAndBooks)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, "John", authorAndBooks.AuthorFirstname)
	assert.Equal(t, "Doe", authorAndBooks.AuthorLastname)
	assert.Equal(t, "john.jpg", authorAndBooks.AuthorPhoto)
	assert.Equal(t, 2, len(authorAndBooks.Books))
	assert.Equal(t, "Book Title 1", authorAndBooks.Books[0].BookTitle)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetBookByID tests the GetBookByID handler
func TestGetBookByID(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	bookID := "1"

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{
		"book_title", "author_id", "book_photo", "is_borrowed", "book_id", "book_details", "author_lastname", "author_firstname",
	}).AddRow("Book Title", 1, "book.jpg", false, 1, "Book details", "Doe", "John")

	mock.ExpectQuery("SELECT books.title AS book_title, books.author_id AS author_id").
		WithArgs(1).
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/books/1", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetBookByID)
	// Using mux.Vars to mock the ID parameter
	req = mux.SetURLVars(req, map[string]string{"id": bookID})
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var book BookAuthorInfo
	err = json.NewDecoder(rr.Body).Decode(&book)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, "Book Title", book.BookTitle)
	assert.Equal(t, "Doe", book.AuthorLastname)
	assert.Equal(t, "John", book.AuthorFirstname)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetSubscribersByBookID tests the GetSubscribersByBookID handler
func TestGetSubscribersByBookID(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	bookID := "1"

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{"Lastname", "Firstname", "Email"}).
		AddRow("Doe", "John", "john.doe@example.com").
		AddRow("Smith", "Jane", "jane.smith@example.com")

	mock.ExpectQuery("SELECT s.Lastname, s.Firstname, s.Email").
		WithArgs(bookID).
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/subscribers/1", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetSubscribersByBookID)
	// Using mux.Vars to mock the ID parameter
	req = mux.SetURLVars(req, map[string]string{"id": bookID})
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var subscribers []Subscriber
	err = json.NewDecoder(rr.Body).Decode(&subscribers)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 2, len(subscribers))
	assert.Equal(t, "Doe", subscribers[0].Lastname)
	assert.Equal(t, "john.doe@example.com", subscribers[0].Email)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAllSubscribers tests the GetAllSubscribers handler
func TestGetAllSubscribers(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
		AddRow("Doe", "John", "john.doe@example.com").
		AddRow("Smith", "Jane", "jane.smith@example.com")

	mock.ExpectQuery("SELECT lastname, firstname, email FROM subscribers").
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/subscribers", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.GetAllSubscribers)
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the JSON response
	var subscribers []Subscriber
	err = json.NewDecoder(rr.Body).Decode(&subscribers)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 2, len(subscribers))
	assert.Equal(t, "Doe", subscribers[0].Lastname)
	assert.Equal(t, "Smith", subscribers[1].Lastname)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestAddAuthorPhoto tests the AddAuthorPhoto handler
func TestAddAuthorPhoto(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	authorID := "1"

	// Set up the SQL mock expectations
	mock.ExpectExec("^UPDATE authors SET photo = \\? WHERE id = \\?$").
		WithArgs("./upload/1/fullsize.jpg", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with a mocked file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Could not create form file: %v", err)
	}
	part.Write([]byte("test image content"))
	writer.Close()

	req, err := http.NewRequest("POST", "/author/photo/1", body)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.AddAuthorPhoto)
	// Use mux.Vars to mock the ID parameter
	req = mux.SetURLVars(req, map[string]string{"id": authorID})
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "File uploaded successfully: ./upload/1/fullsize.jpg\n"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestAddAuthor tests the AddAuthor handler
func TestAddAuthor(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	mock.ExpectExec("INSERT INTO authors").
		WithArgs("Doe", "John", "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Creating a new HTTP request with JSON body
	author := Author{Firstname: "John", Lastname: "Doe"}
	body, err := json.Marshal(author)
	if err != nil {
		t.Fatalf("Could not marshal author: %v", err)
	}

	req, err := http.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.AddAuthor)
	handler.ServeHTTP(rr, req)

	// Ensuring the response status is 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Checking the JSON response
	var response map[string]int
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verifying the response data
	assert.Equal(t, 1, response["id"])

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestAddBookPhoto tests the AddBookPhoto handler
func TestAddBookPhoto(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	bookID := "1"

	// Set up the SQL mock expectations
	mock.ExpectExec("^UPDATE books SET photo = \\? WHERE id = \\?$").
		WithArgs("./upload/books/1/fullsize.jpg", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with a mocked file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Could not create form file: %v", err)
	}
	part.Write([]byte("test image content"))
	writer.Close()

	req, err := http.NewRequest("POST", "/books/photo/1", body)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.AddBookPhoto)
	// Use mux.Vars to mock the ID parameter
	req = mux.SetURLVars(req, map[string]string{"id": bookID})
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "File uploaded successfully: ./upload/books/1/fullsize.jpg\n"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestAddBook tests the AddBook handler
func TestAddBook(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Set up the SQL mock expectations
	mock.ExpectExec("INSERT INTO books").
		WithArgs("Test Book", "", "Some details", 1, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	book := Book{Title: "Test Book", AuthorID: 1, Details: "Some details", IsBorrowed: false}
	body, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Could not marshal book: %v", err)
	}

	req, err := http.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.AddBook)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Check the JSON response
	var response map[string]int
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verify the response data
	assert.Equal(t, 1, response["id"])

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestAddSubscriber tests the AddSubscriber handler
func TestAddSubscriber(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Set up the SQL mock expectations
	mock.ExpectExec("INSERT INTO subscribers").
		WithArgs("Doe", "John", "john.doe@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	subscriber := Subscriber{Firstname: "John", Lastname: "Doe", Email: "john.doe@example.com"}
	body, err := json.Marshal(subscriber)
	if err != nil {
		t.Fatalf("Could not marshal subscriber: %v", err)
	}

	req, err := http.NewRequest("POST", "/subscribers/new", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.AddSubscriber)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Check the JSON response
	var response map[string]int
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Verify the response data
	assert.Equal(t, 1, response["id"])

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestBorrowBook tests the BorrowBook handler
func TestBorrowBook(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Set up SQL mock expectations for checking if the book is borrowed
	mock.ExpectQuery("SELECT is_borrowed FROM books WHERE id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"is_borrowed"}).AddRow(false))

	// Set up SQL mock expectations for inserting into borrowed_books
	mock.ExpectExec("INSERT INTO borrowed_books").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Set up SQL mock expectations for updating the books table
	mock.ExpectExec("UPDATE books SET is_borrowed = TRUE WHERE id = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	requestBody := struct {
		SubscriberID int `json:"subscriber_id"`
		BookID       int `json:"book_id"`
	}{
		SubscriberID: 1,
		BookID:       1,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Could not marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "/book/borrow", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.BorrowBook)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 201 Created
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Check the response message
	expected := `{"message": "Book borrowed successfully"}`
	assert.JSONEq(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestReturnBorrowedBook tests the ReturnBorrowedBook handler
func TestReturnBorrowedBook(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Set up SQL mock expectations for checking if the book is borrowed
	mock.ExpectQuery("^SELECT is_borrowed FROM books WHERE id = \\?$").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"is_borrowed"}).AddRow(true))

	// Set up SQL mock expectations for updating the borrowed_books table
	mock.ExpectExec("^UPDATE borrowed_books SET return_date = NOW\\(\\) WHERE subscriber_id = \\? AND book_id = \\?$").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Set up SQL mock expectations for updating the books table
	mock.ExpectExec("^UPDATE books SET is_borrowed = FALSE WHERE id = \\?$").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	requestBody := struct {
		SubscriberID int `json:"subscriber_id"`
		BookID       int `json:"book_id"`
	}{
		SubscriberID: 1,
		BookID:       1,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Could not marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "/book/return", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.ReturnBorrowedBook)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "Book returned successfully"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestUpdateAuthor tests the UpdateAuthor handler
func TestUpdateAuthor(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	authorID := "1"

	// Set up SQL mock expectations for updating the author
	mock.ExpectExec("^UPDATE authors SET lastname = \\?, firstname = \\?, photo = \\? WHERE id = \\?$").
		WithArgs("Doe", "John", "john.jpg", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	author := Author{Firstname: "John", Lastname: "Doe", Photo: "john.jpg"}
	body, err := json.Marshal(author)
	if err != nil {
		t.Fatalf("Could not marshal author: %v", err)
	}

	req, err := http.NewRequest("PUT", "/authors/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": authorID})

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.UpdateAuthor)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "Author updated successfully"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestUpdateBook tests the UpdateBook handler
func TestUpdateBook(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	bookID := "1"

	// Set up SQL mock expectations for updating the book
	mock.ExpectExec("^UPDATE books SET title = \\?, author_id = \\?, photo = \\?, details = \\?, is_borrowed = \\? WHERE id = \\?$").
		WithArgs("New Title", 1, "newphoto.jpg", "Some details", false, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	book := struct {
		Title      string `json:"title"`
		AuthorID   int    `json:"author_id"`
		Photo      string `json:"photo"`
		Details    string `json:"details"`
		IsBorrowed bool   `json:"is_borrowed"`
	}{
		Title:      "New Title",
		AuthorID:   1,
		Photo:      "newphoto.jpg",
		Details:    "Some details",
		IsBorrowed: false,
	}
	body, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Could not marshal book: %v", err)
	}

	req, err := http.NewRequest("PUT", "/books/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": bookID})

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.UpdateBook)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "Book updated successfully"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestUpdateSubscriber tests the UpdateSubscriber handler
func TestUpdateSubscriber(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	subscriberID := "1"

	// Set up SQL mock expectations for updating the subscriber
	mock.ExpectExec("^UPDATE subscribers SET lastname = \\?, firstname = \\?, email = \\? WHERE id = \\?$").
		WithArgs("Doe", "John", "john.doe@example.com", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new HTTP request with JSON body
	subscriber := Subscriber{Firstname: "John", Lastname: "Doe", Email: "john.doe@example.com"}
	body, err := json.Marshal(subscriber)
	if err != nil {
		t.Fatalf("Could not marshal subscriber: %v", err)
	}

	req, err := http.NewRequest("PUT", "/subscribers/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": subscriberID})

	// Capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.UpdateSubscriber)
	handler.ServeHTTP(rr, req)

	// Ensure the response status is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response message
	expected := "Subscriber updated successfully"
	assert.Equal(t, expected, rr.Body.String())

	// Ensure all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}


