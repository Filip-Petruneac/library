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

	logger := log.New(io.Discard, "", log.LstdFlags) // A logger that discards all output

	return &App{
		DB:     db,
		Logger: logger,
	}, mock
}

// Test for getEnv function using Dependency Injection
func TestGetEnv(t *testing.T) {
	// Backup original environment variable
	originalValue, isSet := os.LookupEnv("TEST_KEY")
	defer func() {
		if isSet {
			os.Setenv("TEST_KEY", originalValue)
		} else {
			os.Unsetenv("TEST_KEY")
		}
	}()

	// Case 1: Environment variable is set
	os.Setenv("TEST_KEY", "expected_value")
	actual := getEnv("TEST_KEY", "default_value")
	assert.Equal(t, "expected_value", actual)

	// Case 2: Environment variable is not set, should return default value
	os.Unsetenv("TEST_KEY")
	actual = getEnv("TEST_KEY", "default_value")
	assert.Equal(t, "default_value", actual)
}

func TestInitDB(t *testing.T) {
	// Create a mock DB and sqlmock with monitoring pings enabled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err, "Error should be nil when creating sqlmock")
	defer db.Close()

	// Expected DSN (Data Source Name)
	dsn := "user:password@tcp(localhost:3306)/testdb"

	// Override sqlOpen with a mock
	originalSQLOpen := sqlOpen  // Save the original function
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		if dataSourceName == dsn {
			return db, nil
		}
		return nil, fmt.Errorf("unexpected DSN: %s", dataSourceName)
	}

	// Restore the original sql.Open after the test
	defer func() { sqlOpen = originalSQLOpen }()

	t.Run("Successful Database Initialization", func(t *testing.T) {
		// Set expectation for Ping method
		mock.ExpectPing()

		// Call the function to test
		_, err = initDB("user", "password", "localhost", "3306", "testdb")
		assert.NoError(t, err, "Error should be nil when initializing the DB")

		// Ensure all expectations are met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "There should be no unmet expectations")
	})

	t.Run("Failed to Open Database Connection", func(t *testing.T) {
		// Override sqlOpen to simulate a connection error
		sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
			return nil, fmt.Errorf("failed to connect to the database")
		}

		_, err = initDB("invalid_user", "invalid_password", "localhost", "3306", "testdb")
		assert.Error(t, err, "Expected an error when failing to connect to the database")
		assert.Contains(t, err.Error(), "failed to connect to the database")

		// Restore sqlOpen to the original mock
		sqlOpen = originalSQLOpen
	})

	t.Run("Failed to Ping Database", func(t *testing.T) {
		// Set expectation for Ping method to fail
		mock.ExpectPing().WillReturnError(fmt.Errorf("failed to ping"))

		// Override sqlOpen to use mock db that has ping failure
		sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
			if dataSourceName == dsn {
				return db, nil
			}
			return nil, fmt.Errorf("unexpected DSN: %s", dataSourceName)
		}

		_, err = initDB("user", "password", "localhost", "3306", "testdb")
		assert.Error(t, err, "Expected an error when pinging the database")
		assert.Contains(t, err.Error(), "failed to ping the database")

		// Ensure all expectations are met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "There should be no unmet expectations")
	})
}

// TestHome tests the Home handler
func TestHome(t *testing.T) {
	// Create a test app instance using the existing createTestApp function
	app, _ := createTestApp(t) // We don't need the mock in this case
	defer app.DB.Close()

	// Create a new HTTP request for the Home handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the Home handler directly, passing the ResponseRecorder and the request
	handler := http.HandlerFunc(app.Home)
	handler.ServeHTTP(rr, req)

	// Check that the status code is what we expect
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check that the response body is what we expect
	expectedBody := "Homepage"
	assert.Equal(t, expectedBody, rr.Body.String(), "Expected response body 'Homepage'")
}

// TestInfo tests the Info handler
func TestInfo(t *testing.T) {
	// Create a test app instance using the existing createTestApp function
	app, _ := createTestApp(t) // We don't need the mock in this case
	defer app.DB.Close()

	// Create a new HTTP request for the Info handler
	req, err := http.NewRequest("GET", "/info", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the Info handler directly, passing the ResponseRecorder and the request
	handler := http.HandlerFunc(app.Info)
	handler.ServeHTTP(rr, req)

	// Check that the status code is what we expect
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

	// Check that the response body is what we expect
	expectedBody := "Info page"
	assert.Equal(t, expectedBody, rr.Body.String(), "Expected response body 'Info page'")
}

// TestSetupRouter verifies that all routes are correctly set up in the router
func TestSetupRouter(t *testing.T) {
	// Create a test app instance using the existing createTestApp function
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Initialize the router
	router := app.setupRouter()

	// Define a list of test cases to check each route
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
				// Mock database call for GetAllSubscribers
				rows := sqlmock.NewRows([]string{"lastname", "firstname", "email"}).
					AddRow("Doe", "John", "john.doe@example.com").
					AddRow("Smith", "Jane", "jane.smith@example.com")

				// Expect the correct SQL query that matches the handler's query
				mock.ExpectQuery(`SELECT lastname, firstname, email FROM subscribers`).WillReturnRows(rows)
			},
		},
		{
			name:           "Get all books",
			method:         "GET",
			path:           "/books",
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				// Mock database call for GetAllBooks
				rows := sqlmock.NewRows([]string{
					"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
				}).
					AddRow(1, "Sample Book", 1, "book.jpg", false, "A sample book", "Doe", "John").
					AddRow(2, "Another Book", 2, "another.jpg", true, "Another sample book", "Smith", "Jane")

				// Expect the correct SQL query that matches the handler's query
				mock.ExpectQuery(`SELECT books.id AS book_id, books.title AS book_title, books.author_id AS author_id, books.photo AS book_photo, books.is_borrowed AS is_borrowed, books.details AS book_details, authors.Lastname AS author_lastname, authors.Firstname AS author_firstname FROM books JOIN authors ON books.author_id = authors.id`).WillReturnRows(rows)
			},
		},
	}

	// Iterate through each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock expectations for this specific test case
			tt.mockSetup()

			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			rr := httptest.NewRecorder()

			// Serve the request using the router
			router.ServeHTTP(rr, req)

			// Assert that the response status matches the expected status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Ensure all mock expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Not all expectations were met: %v", err)
			}
		})
	}
}

// TestRespondWithJSON tests the RespondWithJSON function
func TestRespondWithJSON(t *testing.T) {
    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Create a sample payload
    payload := map[string]string{"message": "success"}

    // Call the RespondWithJSON function
    RespondWithJSON(rr, http.StatusOK, payload)

    // Check the Content-Type header
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type should be application/json")

    // Check the status code
    assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200")

    // Check the response body
    expectedBody, _ := json.Marshal(payload)
    assert.JSONEq(t, string(expectedBody), rr.Body.String(), "Response body should match the payload")
}

func TestRespondWithJSON_ErrorEncoding(t *testing.T) {
    rr := httptest.NewRecorder()
    payload := make(chan int) // A channel cannot be encoded into JSON, will cause an error

    RespondWithJSON(rr, http.StatusOK, payload)

    // Check that the Content-Type header was originally application/json
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type should initially be application/json")

    // Check that the status code is now 500 due to JSON encoding error
    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500 for encoding error")

    // Check that the response body contains the error message
    assert.Equal(t, "Error encoding response\n", rr.Body.String(), "Response body should contain the error message")
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
    payload := make(chan int) // Provocăm o eroare

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
    // Case 1: Valid ID
    req := httptest.NewRequest("GET", "/authors/1", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    id, err := GetIDFromRequest(req, "id")
    assert.NoError(t, err, "Expected no error for a valid ID")
    assert.Equal(t, 1, id, "Expected ID to be 1")

    // Case 2: Invalid ID
    req = httptest.NewRequest("GET", "/authors/abc", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    _, err = GetIDFromRequest(req, "id")
    assert.Error(t, err, "Expected an error for an invalid ID")
    assert.Contains(t, err.Error(), "invalid id", "Error message should mention 'invalid id'")
}

func TestGetIDFromRequest_ParamNotFound(t *testing.T) {
    req := httptest.NewRequest("GET", "/authors", nil) 
    req = mux.SetURLVars(req, map[string]string{}) 

    _, err := GetIDFromRequest(req, "id")
    assert.Error(t, err, "Expected an error when parameter is not found")
    assert.Contains(t, err.Error(), "invalid id", "Error message should mention 'invalid id'")
}

func TestGetIDFromRequest_Success(t *testing.T) {
    req := httptest.NewRequest("GET", "/authors/1", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "1"})

    id, err := GetIDFromRequest(req, "id")
    assert.NoError(t, err, "Expected no error when ID is valid")
    assert.Equal(t, 1, id, "Expected ID to be 1")
}

func TestGetIDFromRequest_Error(t *testing.T) {
    req := httptest.NewRequest("GET", "/authors/abc", nil)
    req = mux.SetURLVars(req, map[string]string{"id": "abc"})

    _, err := GetIDFromRequest(req, "id")
    assert.Error(t, err, "Expected an error when ID is not a number")
    assert.Contains(t, err.Error(), "invalid id", "Error message should mention 'invalid id'")
}

// TestScanAuthors tests the ScanAuthors function
func TestScanAuthors(t *testing.T) {
    // Create a mock DB and sqlmock
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    // Create mock rows to be returned
    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow(1, "Doe", "John", "photo.jpg").
        AddRow(2, "Smith", "Jane", "photo2.jpg")

    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    // Execute the query
    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Query execution should not return an error")

    // Call the ScanAuthors function
    authors, err := ScanAuthors(result)
    assert.NoError(t, err, "Expected no error while scanning authors")

    // Check the number of authors returned
    assert.Equal(t, 2, len(authors), "Expected 2 authors")

    // Check details of the first author
    assert.Equal(t, "John", authors[0].Firstname, "Expected Firstname to be John")
    assert.Equal(t, "Doe", authors[0].Lastname, "Expected Lastname to be Doe")
}

func TestScanAuthors_ErrorAfterIteration(t *testing.T) {
    // Create a mock database and sqlmock
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    // Simulate a set of rows and set an error after the first row
    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow(1, "Doe", "John", "photo.jpg").
        RowError(0, fmt.Errorf("iteration error")) // Set the error after reading the first row

    // Expect the SQL query and return the mock rows with an error
    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    // Execute the query
    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Query execution should not return an error")

    // Call the ScanAuthors function
    authors, err := ScanAuthors(result)

    // Check that an error is returned after the iteration
    assert.Error(t, err, "Expected an error after iteration")
    assert.Nil(t, authors, "Authors should be nil on error")
}

func TestScanAuthors_ErrorDuringScan(t *testing.T) {
    // Create a mock database and sqlmock
    db, mock, err := sqlmock.New()
    assert.NoError(t, err, "Error should be nil when creating sqlmock")
    defer db.Close()

    // Simulate a set of rows that will trigger a scan error
    rows := sqlmock.NewRows([]string{"id", "lastname", "firstname", "photo"}).
        AddRow("invalid_id", "Doe", "John", "photo.jpg") // Intentionally use an invalid ID format to cause a scan error

    // Expect the SQL query and return the mock rows
    mock.ExpectQuery(`SELECT id, lastname, firstname, photo FROM authors`).WillReturnRows(rows)

    // Execute the query
    result, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
    assert.NoError(t, err, "Query execution should not return an error")

    // Call the ScanAuthors function
    authors, err := ScanAuthors(result)

    // Check that an error is returned during the scan
    assert.Error(t, err, "Expected an error during scan")
    assert.Nil(t, authors, "Authors should be nil on error")
}

// TestValidateAuthorData tests the ValidateAuthorData function
func TestValidateAuthorData(t *testing.T) {
    // Case 1: Valid data
    author := Author{Firstname: "John", Lastname: "Doe"}
    err := ValidateAuthorData(author)
    assert.NoError(t, err, "Expected no error for valid author data")

    // Case 2: Missing Firstname
    author = Author{Firstname: "", Lastname: "Doe"}
    err = ValidateAuthorData(author)
    assert.Error(t, err, "Expected an error for missing Firstname")
    assert.Contains(t, err.Error(), "firstname and lastname are required fields", "Error message should mention missing fields")

    // Case 3: Missing Lastname
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

    // Simulate an error in query execution
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

    // Simulate successful query execution but with an error during scanning
    mock.ExpectQuery(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE \? OR Lastname LIKE \?`).
        WithArgs("%John%", "%John%").
        WillReturnRows(sqlmock.NewRows([]string{"id", "Firstname", "Lastname", "photo"}).
            AddRow("invalid_id", "John", "Doe", "photo.jpg")) // Invalid data type to simulate scan error

    handler := http.HandlerFunc(app.SearchAuthors)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status code 500")
    assert.Contains(t, rr.Body.String(), "Error scanning authors", "Expected error message for scan error")

    err = mock.ExpectationsWereMet()
    assert.NoError(t, err, "There should be no unmet expectations")
}




// TestSearchBooks tests the SearchBooks handler
func TestSearchBooks(t *testing.T) {
	app, mock := createTestApp(t)
	defer app.DB.Close()

	// Setting up SQL mock expectations
	queryParam := "Harry"
	rows := sqlmock.NewRows([]string{
		"book_id", "book_title", "author_id", "book_photo", "is_borrowed", "book_details", "author_lastname", "author_firstname",
	}).AddRow(1, "Harry Potter", 1, "harry.jpg", false, "Magic book", "Rowling", "J.K.").
		AddRow(2, "Harry Potter and the Chamber of Secrets", 1, "chamber.jpg", true, "Second book in the series", "Rowling", "J.K.")

	mock.ExpectQuery("SELECT books.id AS book_id,").
		WithArgs("%"+queryParam+"%", "%"+queryParam+"%", "%"+queryParam+"%").
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/search_books?query=Harry", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Capturing the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.SearchBooks)
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
	assert.Equal(t, "Harry Potter", books[0].BookTitle)
	assert.Equal(t, "Rowling", books[0].AuthorLastname)
	assert.Equal(t, "Harry Potter and the Chamber of Secrets", books[1].BookTitle)

	// Ensuring all mock expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Not all expectations were met: %v", err)
	}
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



