package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
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
	r.HandleFunc("/authors_books", GetAuthorsAndBooks(dbService.DB)).Methods("GET")
	r.HandleFunc("/authors_books/{id}", GetAuthorBooksByID(dbService.DB)).Methods("GET")
	r.HandleFunc("/books/{id}", GetBookByID(dbService.DB)).Methods("GET")
	r.HandleFunc("/books/{id}/subscribers", GetSubscribersByBookID(dbService.DB)).Methods("GET")
	r.HandleFunc("/subscribers", GetAllSubscribers(dbService.DB)).Methods("GET")
	r.HandleFunc("/authors/{id}/photo", AddAuthorPhoto(dbService.DB)).Methods("POST")
	r.HandleFunc("/books/photo/{id}", AddBookPhoto(dbService.DB)).Methods("POST")
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

// Test for GetAuthorsAndBooks handler
func TestGetAuthorsAndBooks(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"author_firstname", "author_lastname", "book_title", "book_photo"}
	rows := sqlmock.NewRows(columns).
		AddRow("Ion", "Creangă", "Amintiri din copilărie", "amintiri.jpg").
		AddRow("Jack", "London", "Chemarea străbunilor", "chemarea.jpg")

	dbService.Mock.ExpectQuery(`
		SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo
		FROM authors_books ab
		JOIN authors a ON ab.author_id = a.id
		JOIN books b ON ab.book_id = b.id
	`).WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/authors_books", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedAuthorsAndBooks := []AuthorBook{
		{AuthorFirstname: "Ion", AuthorLastname: "Creangă", BookTitle: "Amintiri din copilărie", BookPhoto: "amintiri.jpg"},
		{AuthorFirstname: "Jack", AuthorLastname: "London", BookTitle: "Chemarea străbunilor", BookPhoto: "chemarea.jpg"},
	}
	var actualAuthorsAndBooks []AuthorBook
	err = json.NewDecoder(rr.Body).Decode(&actualAuthorsAndBooks)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if !reflect.DeepEqual(expectedAuthorsAndBooks, actualAuthorsAndBooks) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedAuthorsAndBooks, actualAuthorsAndBooks)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for GetAuthorBooksByID handler
func TestGetAuthorBooksByID(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"author_firstname", "author_lastname", "author_photo", "book_title", "book_photo"}
	rows := sqlmock.NewRows(columns).
		AddRow("Ion", "Creangă", "ion.jpg", "Amintiri din copilărie", "amintiri.jpg").
		AddRow("Ion", "Creangă", "ion.jpg", "Povești", "povesti.jpg")

	dbService.Mock.ExpectQuery(`
		SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, a.Photo AS author_photo, b.title AS book_title, b.photo AS book_photo
		FROM authors_books ab
		JOIN authors a ON ab.author_id = a.id
		JOIN books b ON ab.book_id = b.id
		WHERE a.id = \?`).
		WithArgs(1).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/authors_books/1", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedResponse := struct {
		AuthorFirstname string       `json:"author_firstname"`
		AuthorLastname  string       `json:"author_lastname"`
		AuthorPhoto     string       `json:"author_photo"`
		Books           []AuthorBook `json:"books"`
	}{
		AuthorFirstname: "Ion",
		AuthorLastname:  "Creangă",
		AuthorPhoto:     "ion.jpg",
		Books: []AuthorBook{
			{BookTitle: "Amintiri din copilărie", BookPhoto: "amintiri.jpg"},
			{BookTitle: "Povești", BookPhoto: "povesti.jpg"},
		},
	}

	var actualResponse struct {
		AuthorFirstname string       `json:"author_firstname"`
		AuthorLastname  string       `json:"author_lastname"`
		AuthorPhoto     string       `json:"author_photo"`
		Books           []AuthorBook `json:"books"`
	}
	err = json.NewDecoder(rr.Body).Decode(&actualResponse)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if !reflect.DeepEqual(expectedResponse, actualResponse) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedResponse, actualResponse)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for GetBookByID handler
func TestGetBookByID(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"book_title", "author_id", "book_photo", "is_borrowed", "book_id", "book_details", "author_lastname", "author_firstname"}
	rows := sqlmock.NewRows(columns).
		AddRow("Amintiri din copilărie", 1, "amintiri.jpg", false, 1, "A novel by Ion Creangă", "Creangă", "Ion")

	dbService.Mock.ExpectQuery(`
		SELECT 
			books.title AS book_title, 
			books.author_id AS author_id, 
			books.photo AS book_photo, 
			books.is_borrowed AS is_borrowed, 
			books.id AS book_id,
			books.details AS book_details,
			authors.Lastname AS author_lastname, 
			authors.Firstname AS author_firstname
		FROM books
		JOIN authors ON books.author_id = authors.id
		WHERE books.id = \?`).
		WithArgs(1).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/books/1", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBook := BookAuthorInfo{
		BookTitle:       "Amintiri din copilărie",
		AuthorID:        1,
		BookPhoto:       "amintiri.jpg",
		IsBorrowed:      false,
		BookID:          1,
		BookDetails:     "A novel by Ion Creangă",
		AuthorLastname:  "Creangă",
		AuthorFirstname: "Ion",
	}

	var actualBook BookAuthorInfo
	err = json.NewDecoder(rr.Body).Decode(&actualBook)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if !reflect.DeepEqual(expectedBook, actualBook) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedBook, actualBook)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for GetSubscribersByBookID handler
func TestGetSubscribersByBookID(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"Lastname", "Firstname", "Email"}
	rows := sqlmock.NewRows(columns).
		AddRow("Doe", "John", "john.doe@example.com").
		AddRow("Smith", "Jane", "jane.smith@example.com")

	dbService.Mock.ExpectQuery(`
		SELECT s.Lastname, s.Firstname, s.Email
		FROM subscribers s
		JOIN borrowed_books bb ON s.id = bb.subscriber_id
		WHERE bb.book_id = \?`).
		WithArgs("1").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/books/1/subscribers", nil)
	rr := httptest.NewRecorder()
	setupTestRouter(dbService).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	t.Logf("Response body: %s", rr.Body.String())

	expectedSubscribers := []Subscriber{
		{Lastname: "Doe", Firstname: "John", Email: "john.doe@example.com"},
		{Lastname: "Smith", Firstname: "Jane", Email: "jane.smith@example.com"},
	}

	var actualSubscribers []Subscriber
	err = json.NewDecoder(rr.Body).Decode(&actualSubscribers)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if !reflect.DeepEqual(expectedSubscribers, actualSubscribers) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedSubscribers, actualSubscribers)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for GetAllSubscribers handler
func TestGetAllSubscribers(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating test DB service: %v", err)
	}
	defer dbService.DB.Close()

	columns := []string{"lastname", "firstname", "email"}
	rows := sqlmock.NewRows(columns).
		AddRow("Doe", "John", "john.doe@example.com").
		AddRow("Smith", "Jane", "jane.smith@example.com")

	dbService.Mock.ExpectQuery("SELECT lastname, firstname, email FROM subscribers").
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/subscribers", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/subscribers", GetAllSubscribers(dbService.DB)).Methods("GET")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedSubscribers := []Subscriber{
		{Lastname: "Doe", Firstname: "John", Email: "john.doe@example.com"},
		{Lastname: "Smith", Firstname: "Jane", Email: "jane.smith@example.com"},
	}

	var actualSubscribers []Subscriber
	err = json.NewDecoder(rr.Body).Decode(&actualSubscribers)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if !reflect.DeepEqual(expectedSubscribers, actualSubscribers) {
		t.Errorf("Response body does not match. Expected %+v, got %+v", expectedSubscribers, actualSubscribers)
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for AddAuthorPhoto handler
func TestAddAuthorPhoto(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating TestDBService: %v", err)
	}
	defer dbService.DB.Close()

	router := setupTestRouter(dbService)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	file, err := writer.CreateFormFile("file", "testfile.jpg")
	if err != nil {
		t.Fatalf("Error creating form file: %v", err)
	}

	fileContent := []byte("test file content")
	_, err = file.Write(fileContent)
	if err != nil {
		t.Fatalf("Error writing file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Error closing multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/authors/1/photo", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	expectedPhotoPath := "./upload/1/fullsize.jpg" // Ensure this matches your handler's path
	dbService.Mock.ExpectExec(`^UPDATE authors SET photo = \? WHERE id = \?$`).
		WithArgs(expectedPhotoPath, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", rec.Code)
	}

	expected := "File uploaded successfully: " + expectedPhotoPath + "\n"
	if rec.Body.String() != expected {
		t.Fatalf("Expected response body '%s', got '%s'", expected, rec.Body.String())
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}

	os.RemoveAll("./upload/1")
}

// Test for AddAuthor handler
func TestAddAuthor(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating TestDBService: %v", err)
	}
	defer dbService.DB.Close()

	handler := AddAuthor(dbService.DB)

	req := httptest.NewRequest(http.MethodGet, "/authors", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Logf("Response body: %s", rec.Body.String())
		t.Fatalf("Expected status code 405, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/authors", strings.NewReader("{invalid json"))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Logf("Response body: %s", rec.Body.String())
		t.Fatalf("Expected status code 400, got %d", rec.Code)
	}

	reqBody := `{"firstname": ""}`
	req = httptest.NewRequest(http.MethodPost, "/authors", strings.NewReader(reqBody))
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Logf("Response body: %s", rec.Body.String())
		t.Fatalf("Expected status code 400, got %d", rec.Code)
	}

	reqBody = `{"firstname": "John", "lastname": "Doe"}`
	req = httptest.NewRequest(http.MethodPost, "/authors", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	dbService.Mock.ExpectExec(`^INSERT INTO authors \(lastname, firstname, photo\) VALUES \(\?, \?, \?\)$`).
		WithArgs("Doe", "John", "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Logf("Response body: %s", rec.Body.String())
		t.Fatalf("Expected status code 201, got %d", rec.Code)
	}

	expectedResponse := `{"id":1}`
	if strings.TrimSpace(rec.Body.String()) != expectedResponse {
		t.Fatalf("Expected response body '%s', got '%s'", expectedResponse, rec.Body.String())
	}

	if err := dbService.Mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unmet expectations: %v", err)
	}
}

// Test for AddBookPhoto handler
func TestAddBookPhoto(t *testing.T) {
	dbService, err := NewTestDBService()
	if err != nil {
		t.Fatalf("Error creating TestDBService: %v", err)
	}
	defer dbService.DB.Close()

	router := mux.NewRouter()
	router.HandleFunc("/books/{id}/photo", AddBookPhoto(dbService.DB)).Methods(http.MethodPost)

	t.Run("Invalid Book ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/books/invalid_id/photo", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Logf("Response body: %s", rec.Body.String())
			t.Fatalf("Expected status code 400, got %d", rec.Code)
		}
	})

	t.Run("Valid Book ID and File Upload", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.jpg")
		if err != nil {
			t.Fatalf("Error creating form file: %v", err)
		}

		_, err = part.Write([]byte("fake image data"))
		if err != nil {
			t.Fatalf("Error writing to form file: %v", err)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/books/1/photo", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		expectedDir := "./upload/books/1"
		expectedFilePath := expectedDir + "/fullsize.jpg"

		t.Logf("Expected photo path: %s", expectedFilePath)

		dbService.Mock.ExpectExec(`^UPDATE books SET photo = \? WHERE id = \?$`).
			WithArgs(expectedFilePath, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		router.ServeHTTP(rec, req)

		t.Logf("Response code: %d", rec.Code)
		t.Logf("Response body: %s", rec.Body.String())

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", rec.Code)
		}

		expectedResponse := fmt.Sprintf("File uploaded successfully: %s\n", expectedFilePath)
		if rec.Body.String() != expectedResponse {
			t.Fatalf("Expected response body '%s', got '%s'", expectedResponse, rec.Body.String())
		}

		if err := dbService.Mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unmet expectations: %v", err)
		}

		os.RemoveAll(expectedDir)
	})
}

// TestAddBook tests the AddBook handler function
func TestAddBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	book := Book{
		Title:      "Test Book",
		AuthorID:   1,
		Photo:      "testphoto.jpg", // Adjust if needed for testing
		IsBorrowed: false,
		Details:    "Details about test book",
	}

	bookJSON, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal book: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/books", bytes.NewBuffer(bookJSON))
	if err != nil {
		t.Fatalf("Failed to create a new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Adjust ExpectExec to match the handler's SQL query
	mock.ExpectExec(`INSERT INTO books \(title, photo, details, author_id, is_borrowed\)`).
		WithArgs(book.Title, "", book.Details, book.AuthorID, book.IsBorrowed).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Mocking the insertion with ID = 1

	rr := httptest.NewRecorder()
	handler := AddBook(db)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}

	expected := `{"id":1}` // Expect ID returned to be 1
	actual := strings.TrimSpace(rr.Body.String())

	t.Logf("Expected response: '%s'", expected)
	t.Logf("Actual response:   '%s'", actual)

	if actual != expected {
		t.Errorf("Expected response body %s, got %s", expected, actual)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unmet expectations: %v", err)
	}
}

// TestAddSubscriber tests the AddSubscriber handler function
func TestAddSubscriber(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	subscriber := Subscriber{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.doe@example.com",
	}

	subscriberJSON, err := json.Marshal(subscriber)
	if err != nil {
		t.Fatalf("Failed to marshal subscriber: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/subscribers", bytes.NewBuffer(subscriberJSON))
	if err != nil {
		t.Fatalf("Failed to create a new request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	mock.ExpectExec("INSERT INTO subscribers").
		WithArgs(subscriber.Lastname, subscriber.Firstname, subscriber.Email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := httptest.NewRecorder()

	handler := AddSubscriber(db)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	expected := `{"id":1}`
	actual := strings.TrimSpace(rr.Body.String()) // Trim any leading/trailing whitespace or newline characters

	t.Logf("Expected response: '%s'", expected)
	t.Logf("Actual response:   '%s'", actual)

	if actual != expected {
		t.Errorf("Expected response body %s, got %s", expected, actual)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unmet expectations: %v", err)
	}
}
