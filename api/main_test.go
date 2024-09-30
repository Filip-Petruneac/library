package main

import (
	"bytes"
	"mime/multipart"
	"encoding/json"
	"io"
	"log"
	"os"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
	"database/sql"
	
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// createTestApp creates a test instance of the application with mocked dependencies.
func createTestApp(t *testing.T) (*App, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating sqlmock: %v", err)
	}

	logger := log.New(io.Discard, "", log.LstdFlags) 

	return &App{
		DB:     db,
		Logger: logger,
	}, mock
}

// Test for getEnv function using Dependency Injection
func TestGetEnv(t *testing.T) {
	originalValue, isSet := os.LookupEnv("TEST_KEY")
	defer func() {
		if isSet {
			os.Setenv("TEST_KEY", originalValue)
		} else {
			os.Unsetenv("TEST_KEY")
		}
	}()

	os.Setenv("TEST_KEY", "expected_value")
	actual := getEnv("TEST_KEY", "default_value")
	assert.Equal(t, "expected_value", actual)

	os.Unsetenv("TEST_KEY")
	actual = getEnv("TEST_KEY", "default_value")
	assert.Equal(t, "default_value", actual)
}

func TestInitDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err, "Error should be nil when creating sqlmock")
	defer db.Close()

	dsn := "user:password@tcp(localhost:3306)/testdb"

	originalSQLOpen := sqlOpen  
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		if dataSourceName == dsn {
			return db, nil
		}
		return nil, fmt.Errorf("unexpected DSN: %s", dataSourceName)
	}

	defer func() { sqlOpen = originalSQLOpen }()

	t.Run("Successful Database Initialization", func(t *testing.T) {
		mock.ExpectPing()

		_, err = initDB("user", "password", "localhost", "3306", "testdb")
		assert.NoError(t, err, "Error should be nil when initializing the DB")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "There should be no unmet expectations")
	})

	t.Run("Failed to Open Database Connection", func(t *testing.T) {
		sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
			return nil, fmt.Errorf("failed to connect to the database")
		}

		_, err = initDB("invalid_user", "invalid_password", "localhost", "3306", "testdb")
		assert.Error(t, err, "Expected an error when failing to connect to the database")
		assert.Contains(t, err.Error(), "failed to connect to the database")

		sqlOpen = originalSQLOpen
	})

	t.Run("Failed to Ping Database", func(t *testing.T) {
		mock.ExpectPing().WillReturnError(fmt.Errorf("failed to ping"))

		sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
			if dataSourceName == dsn {
				return db, nil
			}
			return nil, fmt.Errorf("unexpected DSN: %s", dataSourceName)
		}

		_, err = initDB("user", "password", "localhost", "3306", "testdb")
		assert.Error(t, err, "Expected an error when pinging the database")
		assert.Contains(t, err.Error(), "failed to ping the database")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "There should be no unmet expectations")
	})
}

// TestHome tests the Home handler
func TestHome(t *testing.T) {
	app, _ := createTestApp(t) 
	defer app.DB.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.Home)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	expectedBody := "Homepage"
	assert.Equal(t, expectedBody, rr.Body.String(), "Expected response body 'Homepage'")
}

// TestInfo tests the Info handler
func TestInfo(t *testing.T) {
	app, _ := createTestApp(t) 
	defer app.DB.Close()

	req, err := http.NewRequest("GET", "/info", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.Info)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	expectedBody := "Info page"
	assert.Equal(t, expectedBody, rr.Body.String(), "Expected response body 'Info page'")
}

// TestSetupRouter verifies that all routes are correctly set up in the router
func TestSetupRouter(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	router := app.setupRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		mockSetup      func()
	}{
		{
			name:           "Get all subscribers",
			method:         "GET",
			path:           "/subscribers",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
					AddRow("Doe", "John", "john.doe@example.com").
					AddRow("Smith", "Jane", "jane.smith@example.com")
				mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).WillReturnRows(rows)
			},
		},
		{
			name:           "Get all books",
			method:         "GET",
			path:           "/books",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{
					"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
				}).
					AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
					AddRow(2, "Another Book", 2, "another.jpg", true, "Another sample book", "Smith", "Jane")

				mock.ExpectQuery(`SELECT books.id AS book_id, books.title AS book_title, books.author_id AS author_id, books.photo AS book_photo, books.is_borrowed AS is_borrowed, books.details AS book_details, authors.Lastname AS author_lastname, authors.Firstname AS author_firstname FROM books JOIN authors ON books.author_id = authors.id`).WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			tt.mockSetup()

			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Not all expectations were met: %v", err)
			}
		})
	}
}

// TestRespondWithJSON tests the RespondWithJSON function
func TestRespondWithJSON(t *testing.T) {
    rr := httptest.NewRecorder()

    payload := map[string]string{"message": "success"}

    RespondWithJSON(rr, http.StatusOK, payload)
  
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type should be application/json")
    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    expectedBody, _ := json.Marshal(payload)
    assert.JSONEq(t, string(expectedBody), rr.Body.String(), "Response body should match the payload")
}

func TestRespondWithJSON_Success(t *testing.T) {
    rr := httptest.NewRecorder()
    payload := map[string]string{"message": "test"}

    RespondWithJSON(rr, http.StatusOK, payload)

    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Expected Content-Type application/json")
    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")
    assert.JSONEq(t, `{"message": "test"}`, rr.Body.String(), "Expected JSON response")
}

func TestRespondWithJSON_Error(t *testing.T) {
    rr := httptest.NewRecorder()
    payload := make(chan int)

    RespondWithJSON(rr, http.StatusOK, payload)

    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Expected Content-Type application/json")
    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500 for encoding error")
    assert.Equal(t, "Error encoding response\n", rr.Body.String(), "Expected error message in response")
}

// TestHandleError tests the HandleError function
func TestHandleError(t *testing.T) {
    rr := httptest.NewRecorder()
    logger := log.New(io.Discard, "", log.LstdFlags) // Logger care nu afiseaza nimic
    message := "test error"
    err := fmt.Errorf("an example error")

    HandleError(rr, logger, message, err, http.StatusInternalServerError)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Equal(t, "test error\n", rr.Body.String(), "Expected error message in response")
}

// TestGetIDFromRequest tests the GetIDFromRequest function
func TestGetIDFromRequest(t *testing.T) {
    req := httptest.NewRequest("GET", "/authors/1", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    id, err := GetIDFromRequest(req, "id")
    assert.NoError(t, err, "Expected no error for a valid ID")
    assert.Equal(t, 1, id, "Expected ID to be 1")

    req = httptest.NewRequest("GET", "/authors/abc", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    _, err = GetIDFromRequest(req, "id")
    assert.Error(t, err, "Expected an error for an invalid ID")
    assert.Contains(t, err.Error(), "invalid id", "Error message should mention 'invalid id'")
}

func TestValidateBookData(t *testing.T) {
    book := Book{Title: "Valid Book Title", AuthorID: 1}
    err := ValidateBookData(book)
    assert.NoError(t, err, "Expected no error for valid book data")

    book = Book{Title: "", AuthorID: 1}
    err = ValidateBookData(book)
    assert.Error(t, err, "Expected an error for missing title")
    assert.Contains(t, err.Error(), "title and authorID are required fields", "Error message should mention missing title")

    book = Book{Title: "Valid Book Title", AuthorID: 0}
    err = ValidateBookData(book)
    assert.Error(t, err, "Expected an error for missing author ID")
    assert.Contains(t, err.Error(), "title and authorID are required fields", "Error message should mention missing author ID")

    book = Book{Title: "", AuthorID: 0}
    err = ValidateBookData(book)
    assert.Error(t, err, "Expected an error for missing title and author ID")
    assert.Contains(t, err.Error(), "title and authorID are required fields", "Error message should mention missing fields")
}

// TestScanAuthors tests the ScanAuthors function
func TestScanAuthors(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow(1, "Doe", "John", "photo.jpg").
        AddRow(2, "Smith", "Jane", "photo2.jpg")

    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Query execution should not return an error")
    authors, err := ScanAuthors(result)
    assert.NoError(t, err, "Expected no error while scanning authors")
    assert.Equal(t, 2, len(authors), "Expected 2 authors")
    assert.Equal(t, "John", authors[0].Firstname, "Expected Firstname to be John")
    assert.Equal(t, "Doe", authors[0].Lastname, "Expected Lastname to be Doe")
}

func TestScanAuthors_ErrorAfterIteration(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Eroarea ar trebui să fie nil la crearea sqlmock")
    defer db.Close()

    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow(1, "Doe", "John", "photo.jpg").
        AddRow(2, "Smith", "Jane", "photo2.jpg").
        RowError(1, fmt.Errorf("iteration error")) 

    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Execuția interogării nu ar trebui să returneze o eroare")

    authors, err := ScanAuthors(result)

    assert.Error(t, err, "Era de așteptat o eroare după iterație")
    assert.Nil(t, authors, "Lista de autori ar trebui să fie nil la eroare")
}


func TestScanAuthors_ErrorDuringScan(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow("invalid_id", "Doe", "John", "photo.jpg") 

    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Query execution should not return an error")

    authors, err := ScanAuthors(result)

    assert.Error(t, err, "Expected an error during scan")
    assert.Nil(t, authors, "Authors should be nil on error")
}

// TestValidateAuthorData tests the ValidateAuthorData function
func TestValidateAuthorData(t *testing.T) {
    author := Author{Firstname: "John", Lastname: "Doe"}
    err := ValidateAuthorData(author)
    assert.NoError(t, err, "Expected no error for valid author data")

    author = Author{Firstname: "", Lastname: "Doe"}
    err = ValidateAuthorData(author)
    assert.Error(t, err, "Expected an error for missing Firstname")
    assert.Contains(t, err.Error(), "firstname and lastname are required fields", "Error message should mention missing fields")

    author = Author{Firstname: "John", Lastname: ""}
    err = ValidateAuthorData(author)
    assert.Error(t, err, "Expected an error for missing Lastname")
    assert.Contains(t, err.Error(), "firstname and lastname are required fields", "Error message should mention missing fields")
}

// TestSearchAuthors_ErrorExecutingQuery tests the case where there is an error executing the SQL query
func TestSearchAuthors_ErrorExecutingQuery(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors?query=John", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \? OR Lastname LIKE \?`).
        WithArgs("%John%", "%John%").
        WillReturnError(fmt.Errorf("query execution error"))

    handler := http.HandlerFunc(app.SearchAuthors)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error executing query", "Expected error message for query execution error")

    err = mock.ExpectationsWereMet()
    assert.NoError(t, err, "There should be no unmet expectations")
}

// TestSearchAuthors_ErrorScanningAuthors tests the case where there is an error scanning the rows
func TestSearchAuthors_ErrorScanningAuthors(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors?query=John", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \? OR Lastname LIKE \?`).
        WithArgs("%John%", "%John%").
        WillReturnRows(sqlmock.NewRows([]string{"id", "Firstname", "Lastname", "photo"}).
            AddRow("invalid_id", "John", "Doe", "photo.jpg")) 

    handler := http.HandlerFunc(app.SearchAuthors)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error scanning authors", "Expected error message for scan error")

    err = mock.ExpectationsWereMet()
    assert.NoError(t, err, "There should be no unmet expectations")
}

func TestSearchAuthors_MissingQueryParameter(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/search_authors", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.SearchAuthors)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400 for missing query parameter")
    assert.Contains(t, rr.Body.String(), "Query parameter is required", "Expected error message for missing query parameter")
}

func TestSearchAuthors_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors?query=John", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow(1, "Doe", "John", "photo.jpg").
        AddRow(2, "Smith", "Jane", "photo2.jpg")

    mock.ExpectQuery(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \? OR Lastname LIKE \?`).
        WithArgs("%John%", "%John%").
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.SearchAuthors)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    expected := []map[string]interface{}{
        {"id": float64(1), "firstname": "John", "lastname": "Doe", "photo": "photo.jpg"},
        {"id": float64(2), "firstname": "Jane", "lastname": "Smith", "photo": "photo2.jpg"},
    }
    var actual []map[string]interface{}
    err = json.Unmarshal(rr.Body.Bytes(), &actual)
    assert.NoError(t, err, "Expected no error while unmarshaling JSON response")

    assert.Equal(t, expected, actual, "Expected JSON response")
}

func TestScanBooks(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    rows := sqlmock.NewRows([]string{"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname"}).
        AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
        AddRow(2, "Another Book", 2, "another.jpg", true, "Another sample book", "Smith", "Jane")

    mock.ExpectQuery(`SELECT (.+) FROM books`).WillReturnRows(rows)

    result, err := db.Query("SELECT book_id, book_title, author_id, book_photo, is_borrowed, book_details, author_lastname, author_firstname FROM books")
    assert.NoError(t, err, "Query execution should not return an error")

    books, err := ScanBooks(result)
    assert.NoError(t, err, "Expected no error while scanning books")
    assert.Equal(t, 2, len(books), "Expected 2 books")
    assert.Equal(t, "Sample Book", books[0].BookTitle, "Expected BookTitle to be 'Sample Book'")
    assert.Equal(t, "Doe", books[0].AuthorLastname, "Expected AuthorLastname to be 'Doe'")
}

func TestSearchBooks_MissingQuery(t *testing.T) {
    
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.SearchBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400")
    assert.Contains(t, rr.Body.String(), "Query parameter is required", "Expected error message for missing query parameter")
}

func TestSearchBooks_ErrorExecutingQuery(t *testing.T) {

    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books?query=Sample", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT (.+) FROM books`).
        WithArgs("%Sample%", "%Sample%", "%Sample%").
        WillReturnError(fmt.Errorf("query execution error"))


    handler := http.HandlerFunc(app.SearchBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error executing query", "Expected error message for query execution error")
}

func TestSearchBooks_ErrorScanningRows(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()
    req, err := http.NewRequest("GET", "/books?query=Sample", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT (.+) FROM books`).
        WithArgs("%Sample%", "%Sample%", "%Sample%").
        WillReturnRows(sqlmock.NewRows([]string{
            "book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
        }).AddRow("invalid_id", "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John")) // Valoare invalidă pentru a provoca o eroare

    handler := http.HandlerFunc(app.SearchBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error scanning books", "Expected error message for row scan error")
}

func TestSearchBooks_Success(t *testing.T) {

    app, mock := createTestApp(t)
    defer app.DB.Close()
    req, err := http.NewRequest("GET", "/books?query=Sample", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()
    rows := sqlmock.NewRows([]string{
        "book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
        AddRow(2, "Another Book", 2, "another.jpg", true, "Another sample book", "Smith", "Jane")

    mock.ExpectQuery(`SELECT (.+) FROM books`).
        WithArgs("%Sample%", "%Sample%", "%Sample%").
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.SearchBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    var books []BookAuthorInfo
    err = json.NewDecoder(rr.Body).Decode(&books)
    assert.NoError(t, err, "Expected no error decoding JSON response")

    assert.Equal(t, 2, len(books), "Expected 2 books")
    assert.Equal(t, "Sample Book", books[0].BookTitle, "Expected BookTitle to be 'Sample Book'")
    assert.Equal(t, "Doe", books[0].AuthorLastname, "Expected AuthorLastname to be 'Doe'")
    assert.Equal(t, "John", books[0].AuthorFirstname, "Expected AuthorFirstname to be 'John'")
}

func TestScanBooks_ErrorAfterIteration(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()


    rows := sqlmock.NewRows([]string{
        "book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
        RowError(0, fmt.Errorf("iteration error")) 

    mock.ExpectQuery(`SELECT (.+) FROM books`).WillReturnRows(rows)

    result, err := app.DB.Query("SELECT (.+) FROM books")
    assert.NoError(t, err, "Expected no error when executing query")

    books, err := ScanBooks(result)

    assert.Error(t, err, "Expected an error after iteration")
    assert.Nil(t, books, "Books should be nil on error")

    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Not all expectations were met: %v", err)
    }
}

// TestGetAuthors tests the GetAuthors handler with Dependency Injection
func TestGetAuthors(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	req, err := http.NewRequest("GET", "/authors", nil)
	assert.NoError(t, err, "Error should be nil when creating a new request")

	rr := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
		AddRow(1, "Doe", "John", "photo.jpg").
		AddRow(2, "Smith", "Jane", "photo2.jpg")

	mock.ExpectQuery(`SELECT id, Lastname, Firstname, photo FROM authors ORDER BY Lastname, Firstname`).
		WillReturnRows(rows)

	handler := http.HandlerFunc(app.GetAuthors)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	expected := []map[string]interface{}{
		{"id": float64(1), "lastname": "Doe", "firstname": "John", "photo": "photo.jpg"},
		{"id": float64(2), "lastname": "Smith", "firstname": "Jane", "photo": "photo2.jpg"},
	}
	var actual []map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &actual)
	assert.NoError(t, err, "Expected no error while unmarshaling JSON response")

	assert.Equal(t, expected, actual, "Expected JSON response")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAuthors_ErrorExecutingQuery tests the case where there is an error executing the SQL query in GetAuthors handler
func TestGetAuthors_ErrorExecutingQuery(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	req, err := http.NewRequest("GET", "/authors", nil)
	assert.NoError(t, err, "Error should be nil when creating a new request")

	rr := httptest.NewRecorder()

	mock.ExpectQuery(`SELECT id, Lastname, Firstname, photo FROM authors ORDER BY Lastname, Firstname`).
		WillReturnError(fmt.Errorf("query execution error"))

	handler := http.HandlerFunc(app.GetAuthors)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
	assert.Contains(t, rr.Body.String(), "Error executing query", "Expected error message for query execution error")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAuthors_ErrorScanningRows tests the case where there is an error scanning the rows in GetAuthors handler
func TestGetAuthors_ErrorScanningRows(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	req, err := http.NewRequest("GET", "/authors", nil)
	assert.NoError(t, err, "Error should be nil when creating a new request")

	rr := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
		AddRow("invalid_id", "Doe", "John", "photo.jpg") 

	mock.ExpectQuery(`SELECT id, Lastname, Firstname, photo FROM authors ORDER BY Lastname, Firstname`).
		WillReturnRows(rows)

	handler := http.HandlerFunc(app.GetAuthors)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
	assert.Contains(t, rr.Body.String(), "Error scanning authors", "Expected error message for row scan error")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
}

// TestGetAllBooks tests the GetAllBooks handler
func TestGetAllBooks_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{
        "book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
        AddRow(2, "Another Book", 2, "another.jpg", true, "Another sample book", "Smith", "Jane")

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id`).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAllBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    var books []BookAuthorInfo
    err = json.NewDecoder(rr.Body).Decode(&books)
    assert.NoError(t, err)
    assert.Equal(t, 2, len(books))
    assert.Equal(t, "Sample Book", books[0].BookTitle)
    assert.Equal(t, "Doe", books[0].AuthorLastname)
    assert.Equal(t, "John", books[0].AuthorFirstname)
}

func TestGetAllBooks_ErrorQuery(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id`).
        WillReturnError(fmt.Errorf("database query error"))

    handler := http.HandlerFunc(app.GetAllBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error executing query", "Expected 'Error executing query' in response")
}

func TestGetAllBooks_ErrorScan(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{
        "book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow("invalid_id", "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John")

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id`).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAllBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error scanning books")
}

// GetAuthorsAndBooks handler tests
func TestGetAuthorsAndBooks_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authorsbooks", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "book_title", "book_photo"}).
        AddRow("John", "Doe", "Book 1", "book1.jpg").
        AddRow("Jane", "Smith", "Book 2", "book2.jpg")

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id`).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAuthorsAndBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    expected := []map[string]interface{}{
        {"author_firstname": "John", "author_lastname": "Doe", "book_title": "Book 1", "book_photo": "book1.jpg"},
        {"author_firstname": "Jane", "author_lastname": "Smith", "book_title": "Book 2", "book_photo": "book2.jpg"},
    }
    var actual []map[string]interface{}
    err = json.Unmarshal(rr.Body.Bytes(), &actual)
    assert.NoError(t, err, "Expected no error while unmarshaling JSON response")
    assert.Equal(t, expected, actual, "Expected JSON response")
}

func TestGetAuthorsAndBooks_ErrorExecutingQuery(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authorsbooks", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id`).
        WillReturnError(fmt.Errorf("query execution error"))

    handler := http.HandlerFunc(app.GetAuthorsAndBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error executing query", "Expected error message for query execution error")
}

func TestGetAuthorsAndBooks_ErrorScanningRows(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authorsbooks", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "book_title", "book_photo"}).
        AddRow("invalid_data", "Doe", "Book 1", "book1.jpg").
        RowError(0, fmt.Errorf("scan error")) 

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id`).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAuthorsAndBooks)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error scanning authors and books", "Expected error message for scan error")
}

// GetAuthorsAndBooksByID tests
func TestGetAuthorBooksByID_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors/1", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "book_title", "book_photo"}).
        AddRow("John", "Doe", "Book 1", "book1.jpg").
        AddRow("John", "Doe", "Book 2", "book2.jpg")

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id WHERE a.id = ?`).
        WithArgs(1).WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAuthorBooksByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    expected := []map[string]interface{}{
        {"author_firstname": "John", "author_lastname": "Doe", "book_title": "Book 1", "book_photo": "book1.jpg"},
        {"author_firstname": "John", "author_lastname": "Doe", "book_title": "Book 2", "book_photo": "book2.jpg"},
    }
    var actual []map[string]interface{}
    err = json.Unmarshal(rr.Body.Bytes(), &actual)
    assert.NoError(t, err, "Expected no error while unmarshaling JSON response")
    assert.Equal(t, expected, actual, "Expected JSON response")
}

func TestGetAuthorBooksByID_ErrorInvalidID(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors/abc", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.GetAuthorBooksByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status code 400")
    assert.Contains(t, rr.Body.String(), "Invalid author ID", "Expected 'Invalid author ID' in response")
}

func TestGetAuthorBooksByID_ErrorExecutingQuery(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors/1", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id WHERE a.id = ?`).
        WithArgs(1).WillReturnError(fmt.Errorf("query execution error"))

    handler := http.HandlerFunc(app.GetAuthorBooksByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error executing query", "Expected error message for query execution error")
}

func TestGetAuthorBooksByID_ErrorScanningRows(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/authors/1", nil)
    assert.NoError(t, err, "Error should be nil when creating a new request")
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"author_firstname", "author_lastname", "book_title", "book_photo"}).
        AddRow("invalid_data", "Doe", "Book 1", "book1.jpg").
        RowError(0, fmt.Errorf("scan error")) 

    mock.ExpectQuery(`SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo FROM authors_books ab JOIN authors a ON ab.author_id = a.id JOIN books b ON ab.book_id = b.id WHERE a.id = ?`).
        WithArgs(1).WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAuthorBooksByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error scanning results", "Expected error message for scan error")
}

// TestGetBookByID tests the GetBookByID handler
func TestGetBookByID_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{
        "book_title", "author_id", "book_photo", "is_borrowed", "book_id", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow("Sample Book", 1, "book.jpg", false, 1, "A sample book", "Doe", "John")

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id WHERE books.id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetBookByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    var book BookAuthorInfo
    err = json.NewDecoder(rr.Body).Decode(&book)
    assert.NoError(t, err)
    assert.Equal(t, "Sample Book", book.BookTitle)
    assert.Equal(t, "Doe", book.AuthorLastname)
    assert.Equal(t, "John", book.AuthorFirstname)
}

func TestGetBookByID_InvalidID(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/abc", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.GetBookByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid book ID")
}

func TestGetBookByID_BookNotFound(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{
        "book_title", "author_id", "book_photo", "is_borrowed", "book_id", "book_details", "author_lastname", "author_firstname",
    })

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id WHERE books.id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetBookByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusNotFound, rr.Code)
    assert.Contains(t, rr.Body.String(), "Book not found")
}

func TestGetBookByID_ErrorExecutingQuery(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id WHERE books.id = ?`).
        WithArgs(1).
        WillReturnError(fmt.Errorf("query execution error"))

    handler := http.HandlerFunc(app.GetBookByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error executing query")
}

func TestGetBookByID_ErrorScanningRows(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{
        "book_title", "author_id", "book_photo", "is_borrowed", "book_id", "book_details", "author_lastname", "author_firstname",
    }).
        AddRow("Sample Book", "invalid_author_id", "book.jpg", false, 1, "A sample book", "Doe", "John")

    mock.ExpectQuery(`SELECT (.+) FROM books JOIN authors ON books.author_id = authors.id WHERE books.id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetBookByID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error scanning book")
}

// TestGetSubscribersByBookID tests the GetSubscribersByBookID handler
func TestGetSubscribersByBookID_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
        AddRow("Doe", "John", "john.doe@example.com").
        AddRow("Smith", "Jane", "jane.smith@example.com")

    mock.ExpectQuery(`SELECT s.Lastname, s.Firstname, s.Email FROM subscribers s JOIN borrowed_books bb ON s.id = bb.subscriber_id WHERE bb.book_id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetSubscribersByBookID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    expected := []Subscriber{
        {Lastname: "Doe", Firstname: "John", Email: "john.doe@example.com"},
        {Lastname: "Smith", Firstname: "Jane", Email: "jane.smith@example.com"},
    }
    var actual []Subscriber
    err = json.NewDecoder(rr.Body).Decode(&actual)
    assert.NoError(t, err)
    assert.Equal(t, expected, actual)
}

func TestGetSubscribersByBookID_InvalidID(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/invalid", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "invalid"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.GetSubscribersByBookID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid book ID")
}

func TestGetSubscribersByBookID_NoSubscribers(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"})

    mock.ExpectQuery(`SELECT s.Lastname, s.Firstname, s.Email FROM subscribers s JOIN borrowed_books bb ON s.id = bb.subscriber_id WHERE bb.book_id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetSubscribersByBookID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusNotFound, rr.Code)
    assert.Contains(t, rr.Body.String(), "No subscribers found")
}

func TestGetSubscribersByBookID_QueryError(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT s.Lastname, s.Firstname, s.Email FROM subscribers s JOIN borrowed_books bb ON s.id = bb.subscriber_id WHERE bb.book_id = ?`).
        WithArgs(1).
        WillReturnError(fmt.Errorf("query error"))

    handler := http.HandlerFunc(app.GetSubscribersByBookID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error querying the database")
}

func TestGetSubscribersByBookID_ScanError(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/books/1", nil)
    assert.NoError(t, err)

    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
        AddRow("Doe", "John", nil) 

    mock.ExpectQuery(`SELECT s.Lastname, s.Firstname, s.Email FROM subscribers s JOIN borrowed_books bb ON s.id = bb.subscriber_id WHERE bb.book_id = ?`).
        WithArgs(1).
        WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetSubscribersByBookID)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error scanning subscribers")
}

// TestGetAllSubscribers tests the GetAllSubscribers handler
func TestGetAllSubscribers_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/subscribers", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
        AddRow("Doe", "John", "john.doe@example.com").
        AddRow("Smith", "Jane", "jane.smith@example.com")

    mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAllSubscribers)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    expected := []Subscriber{
        {Lastname: "Doe", Firstname: "John", Email: "john.doe@example.com"},
        {Lastname: "Smith", Firstname: "Jane", Email: "jane.smith@example.com"},
    }
    var actual []Subscriber
    err = json.NewDecoder(rr.Body).Decode(&actual)
    assert.NoError(t, err)
    assert.Equal(t, expected, actual)
}

func TestGetAllSubscribers_QueryError(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/subscribers", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).
        WillReturnError(fmt.Errorf("query error"))

    handler := http.HandlerFunc(app.GetAllSubscribers)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error querying the database")
}

func TestGetAllSubscribers_ScanError(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/subscribers", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
        AddRow("Doe", "John", nil) 

    mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAllSubscribers)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error scanning subscribers")
}

func TestGetAllSubscribers_NoSubscribers(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    req, err := http.NewRequest("GET", "/subscribers", nil)
    assert.NoError(t, err)

    rr := httptest.NewRecorder()

    rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"})

    mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).WillReturnRows(rows)

    handler := http.HandlerFunc(app.GetAllSubscribers)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    var actual []Subscriber
    err = json.NewDecoder(rr.Body).Decode(&actual)
    assert.NoError(t, err)
    assert.Len(t, actual, 0) 
}

// Tests for AddAuthorPhoto handler
func TestAddAuthorPhoto_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)

    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)

    writer.Close()

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectExec(`UPDATE authors SET photo = \? WHERE id = \?`).
        WithArgs("./upload/1/fullsize.jpg", 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "File uploaded successfully")
}

func TestAddAuthorPhoto_InvalidAuthorID(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("POST", "/author/photo/abc", nil)  
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})  

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid author ID")
}

func TestAddAuthorPhoto_ErrorSavingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    originalCopy := ioCopy
    ioCopy = func(dst io.Writer, src io.Reader) (int64, error) {
        return 0, fmt.Errorf("Error saving file") 
    }
    defer func() { ioCopy = originalCopy }() 

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error saving file")
}


func TestAddAuthorPhoto_InvalidMethod(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("GET", "/author/photo/1", nil) 
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)  
    assert.Contains(t, rr.Body.String(), "Only POST method is supported")
}

func TestAddAuthorPhoto_ErrorGettingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    writer.Close()  

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error getting file from request")
}

func TestAddAuthorPhoto_ErrorCreatingDirectories(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    originalMkdirAll := mkdirAll
    mkdirAll = func(path string, perm os.FileMode) error {
        return fmt.Errorf("Error creating directories")
    }
    defer func() { mkdirAll = originalMkdirAll }()

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error creating directories")
}

func TestAddAuthorPhoto_ErrorCreatingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    originalCreate := osCreate
    osCreate = func(name string) (*os.File, error) {
        return nil, fmt.Errorf("Error creating file on disk")
    }
    defer func() { osCreate = originalCreate }() 

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error creating file on disk")
}

func TestAddAuthorPhoto_ErrorUpdatingAuthorPhotoInDB(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/author/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectExec(`UPDATE authors SET photo = ? WHERE id = ?`).
        WithArgs("./upload/1/fullsize.jpg", 1).
        WillReturnError(fmt.Errorf("Failed to update author photo"))

    handler := http.HandlerFunc(app.AddAuthorPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to update author photo")
}

// TestAddAuthor tests the AddAuthor handler
func TestAddAuthor_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    author := Author{Firstname: "John", Lastname: "Doe"}
    body, err := json.Marshal(author)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    mock.ExpectExec(`INSERT INTO authors \(lastname, firstname, photo\) VALUES \(\?, \?, \?\)`).
        WithArgs("Doe", "John", "").
        WillReturnResult(sqlmock.NewResult(1, 1))

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusCreated, rr.Code)

    var response map[string]int
    err = json.NewDecoder(rr.Body).Decode(&response)
    assert.NoError(t, err)

    assert.Equal(t, 1, response["id"], "Expected inserted author ID to be 1")
}

func TestAddAuthor_InvalidJSON(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer([]byte("invalid json")))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid JSON data")
}

func TestAddAuthor_FailedToInsert(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    author := Author{Firstname: "John", Lastname: "Doe"}
    body, err := json.Marshal(author)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    mock.ExpectExec(`INSERT INTO authors \(lastname, firstname, photo\) VALUES \(\?, \?, \?\)`).
        WillReturnError(fmt.Errorf("failed to insert author"))

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to insert author")
}

func TestAddAuthor_FailedToGetLastInsertID(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    author := Author{Firstname: "John", Lastname: "Doe"}
    body, err := json.Marshal(author)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    mock.ExpectExec(`INSERT INTO authors \(lastname, firstname, photo\) VALUES \(\?, \?, \?\)`).
        WithArgs("Doe", "John", "").
        WillReturnResult(sqlmock.NewResult(0, 1)) 

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to get last insert ID")
}

func TestAddAuthor_MethodNotAllowed(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("GET", "/authors/new", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
    assert.Contains(t, rr.Body.String(), "Only POST method is supported")
}

type errorWriter struct{}

func (e *errorWriter) Header() http.Header {
    return http.Header{}
}

func (e *errorWriter) Write([]byte) (int, error) {
    return 0, fmt.Errorf("error encoding json")
}

func (e *errorWriter) WriteHeader(statusCode int) {}

func TestAddAuthor_FailedToEncodeJSON(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    author := Author{Firstname: "John", Lastname: "Doe"}
    body, err := json.Marshal(author)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := &errorWriter{}

    mock.ExpectExec(`INSERT INTO authors \(lastname, firstname, photo\) VALUES \(\?, \?, \?\)`).
        WithArgs("Doe", "John", "").
        WillReturnResult(sqlmock.NewResult(1, 1))  

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

}

func TestAddAuthor_InvalidAuthorData(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    author := Author{Firstname: "", Lastname: ""}
    body, err := json.Marshal(author)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/authors/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddAuthor)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "firstname and lastname are required fields")
}

// TestAddBook tests the AddBook handler
func TestAddBookPhoto_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)

    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)

    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectExec(`UPDATE books SET photo = \? WHERE id = \?`).
        WithArgs("./upload/books/1/fullsize.jpg", 1).
        WillReturnResult(sqlmock.NewResult(1, 1))

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Contains(t, rr.Body.String(), "File uploaded successfully")
}

func TestAddBookPhoto_InvalidBookID(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("POST", "/book/photo/abc", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid book ID")
}

func TestAddBookPhoto_ErrorSavingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    originalIoCopy := ioCopy
    ioCopy = func(dst io.Writer, src io.Reader) (int64, error) {
        return 0, fmt.Errorf("mocked error saving file")
    }
    defer func() { ioCopy = originalIoCopy }() 

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error saving file")
}

func TestAddBookPhoto_InvalidMethod(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("GET", "/book/photo/1", nil)
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
    assert.Contains(t, rr.Body.String(), "Only POST method is supported")
}

func TestAddBookPhoto_ErrorGettingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error getting file from request")
}

func TestAddBookPhoto_ErrorCreatingDirectories(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    originalMkdirAll := osMkdirAll
    osMkdirAll = func(path string, perm os.FileMode) error {
        return fmt.Errorf("mocked error creating directories")
    }
    defer func() { osMkdirAll = originalMkdirAll }() 

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error creating directories")
}

func TestAddBookPhoto_ErrorCreatingFile(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    originalOsCreate := osCreate
    osCreate = func(name string) (*os.File, error) {
        return nil, fmt.Errorf("mocked error creating file")
    }
    defer func() { osCreate = originalOsCreate }() 

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Error creating file on disk")
}

func TestAddBookPhoto_ErrorUpdatingBookPhotoInDB(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.jpg")
    assert.NoError(t, err)
    _, err = part.Write([]byte("test data"))
    assert.NoError(t, err)
    writer.Close()

    req := httptest.NewRequest("POST", "/book/photo/1", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    rr := httptest.NewRecorder()

    mock.ExpectExec(`UPDATE books SET photo = ? WHERE id = ?`).
        WithArgs("./upload/books/1/fullsize.jpg", 1).
        WillReturnError(fmt.Errorf("Failed to update book photo"))

    handler := http.HandlerFunc(app.AddBookPhoto)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to update book photo")
}

// Tests for AddBook handler
func TestAddBook_Success(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    book := Book{
        Title:      "Test Book",
        Details:    "Test Details",
        AuthorID:   1,
        IsBorrowed: false,
    }
    body, err := json.Marshal(book)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    
    mock.ExpectExec(`INSERT INTO books \(title, photo, details, author_id, is_borrowed\) VALUES \(\?, \?, \?, \?, \?\)`).
        WithArgs("Test Book", "", "Test Details", 1, false).
        WillReturnResult(sqlmock.NewResult(1, 1))

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusCreated, rr.Code)

    var response map[string]int
    err = json.NewDecoder(rr.Body).Decode(&response)
    assert.NoError(t, err)
    assert.Equal(t, 1, response["id"])
}

func TestAddBook_InvalidJSON(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer([]byte("invalid json")))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.Contains(t, rr.Body.String(), "Invalid JSON data")
}

func TestAddBook_InvalidBookData(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    book := Book{Title: "", AuthorID: 0}
    body, err := json.Marshal(book)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    
    assert.Contains(t, rr.Body.String(), "title and authorID are required fields")
}

func TestAddBook_FailedToInsertBook(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    book := Book{
        Title:      "Test Book",
        Details:    "Test Details",
        AuthorID:   1,
        IsBorrowed: false,
    }
    body, err := json.Marshal(book)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    mock.ExpectExec(`INSERT INTO books \(title, photo, details, author_id, is_borrowed\) VALUES \(\?, \?, \?, \?, \?\)`).
        WithArgs("Test Book", "", "Test Details", 1, false).
        WillReturnError(fmt.Errorf("SQL error"))

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to insert book")
}

func TestAddBook_FailedToGetLastInsertID(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    book := Book{
        Title:      "Test Book",
        Details:    "Test Details",
        AuthorID:   1,
        IsBorrowed: false,
    }
    body, err := json.Marshal(book)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()

    mock.ExpectExec(`INSERT INTO books \(title, photo, details, author_id, is_borrowed\) VALUES \(\?, \?, \?, \?, \?\)`).
        WithArgs("Test Book", "", "Test Details", 1, false).
        WillReturnResult(sqlmock.NewResult(0, 1)) 

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
    assert.Contains(t, rr.Body.String(), "Failed to get last insert ID")
}


func TestAddBook_JSONEncodingError(t *testing.T) {
    app, mock := createTestApp(t)
    defer app.DB.Close()

    book := Book{
        Title:      "Test Book",
        Details:    "Test Details",
        AuthorID:   1,
        IsBorrowed: false,
    }
    body, err := json.Marshal(book)
    assert.NoError(t, err)

    req := httptest.NewRequest("POST", "/books/new", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := &errorWriter{}

    mock.ExpectExec(`INSERT INTO books \(title, photo, details, author_id, is_borrowed\) VALUES \(\?, \?, \?, \?, \?\)`).
        WithArgs("Test Book", "", "Test Details", 1, false).
        WillReturnResult(sqlmock.NewResult(1, 1))

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    app.Logger.Printf("JSON encoding error")
}

func TestAddBook_InvalidMethod(t *testing.T) {
    app, _ := createTestApp(t)
    defer app.DB.Close()

    req := httptest.NewRequest("GET", "/books/new", nil) 
    rr := httptest.NewRecorder()

    handler := http.HandlerFunc(app.AddBook)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
    assert.Contains(t, rr.Body.String(), "Only POST method is supported")
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



