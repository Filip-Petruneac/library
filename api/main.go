package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"io"
	
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// App struct holds all the dependencies for the application
type App struct {
	DB     *sql.DB
	Logger *log.Logger
}

// Data structures for handling information
type AuthorInfo struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Photo     string `json:"photo,omitempty"`
}

type Author struct {
	ID        int    `json:"id"`
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
	Photo     string `json:"photo"`
}

type AuthorBook struct {
	AuthorFirstname string `json:"author_firstname"`
	AuthorLastname  string `json:"author_lastname"`
	BookTitle       string `json:"book_title"`
	BookPhoto       string `json:"book_photo"`
}

type BookAuthorInfo struct {
	BookID          int    `json:"book_id"`
	BookTitle       string `json:"book_title"`
	AuthorID        int    `json:"author_id"`
	BookPhoto       string `json:"book_photo"`
	IsBorrowed      bool   `json:"is_borrowed"`
	BookDetails     string `json:"book_details"`
	AuthorLastname  string `json:"author_lastname"`
	AuthorFirstname string `json:"author_firstname"`
}

type Subscriber struct {
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
	Email     string `json:"email"`
}

type Book struct {
	Title      string `json:"title"`
	AuthorID   int    `json:"author_id"`
	Photo      string `json:"photo"`
	IsBorrowed bool   `json:"is_borrowed"`
	Details    string `json:"details"`
}

// setupRouter configures the application's routes
func (app *App) setupRouter() *mux.Router {
	r := mux.NewRouter()

	// Attach the App's methods to each route
	r.HandleFunc("/", app.Home).Methods("GET")
	r.HandleFunc("/info", app.Info).Methods("GET")
	r.HandleFunc("/books", app.GetAllBooks).Methods("GET")
	r.HandleFunc("/authors", app.GetAuthors).Methods("GET")
	r.HandleFunc("/authorsbooks", app.GetAuthorsAndBooks).Methods("GET")
	r.HandleFunc("/authors/{id}", app.GetAuthorBooksByID).Methods("GET")
	r.HandleFunc("/books/{id}", app.GetBookByID).Methods("GET")
	r.HandleFunc("/subscribers/{id}", app.GetSubscribersByBookID).Methods("GET")
	r.HandleFunc("/subscribers", app.GetAllSubscribers).Methods("GET")

	// Routes for creating resources
	r.HandleFunc("/authors/new", app.AddAuthor).Methods("POST")
	r.HandleFunc("/books/new", app.AddBook).Methods("POST")
	r.HandleFunc("/subscribers/new", app.AddSubscriber).Methods("POST")
	r.HandleFunc("/book/borrow", app.BorrowBook).Methods("POST")
	r.HandleFunc("/book/return", app.ReturnBorrowedBook).Methods("POST")
	r.HandleFunc("/author/photo/{id}", app.AddAuthorPhoto).Methods("POST")
	r.HandleFunc("/books/photo/{id}", app.AddBookPhoto).Methods("POST")

	// Routes for updating resources
	r.HandleFunc("/authors/{id}", app.UpdateAuthor).Methods("PUT", "POST")
	r.HandleFunc("/books/{id}", app.UpdateBook).Methods("PUT", "POST")
	r.HandleFunc("/subscribers/{id}", app.UpdateSubscriber).Methods("PUT", "POST")

	// Routes for deleting resources
	r.HandleFunc("/authors/{id}", app.DeleteAuthor).Methods("DELETE")
	r.HandleFunc("/books/{id}", app.DeleteBook).Methods("DELETE")
	r.HandleFunc("/subscribers/{id}", app.DeleteSubscriber).Methods("DELETE")

	// Routes for searching
	r.HandleFunc("/search_books", app.SearchBooks).Methods("GET")
	r.HandleFunc("/search_authors", app.SearchAuthors).Methods("GET")

	// Routes for login
	r.HandleFunc("/signup", app.SignupUser).Methods("POST") 
    r.HandleFunc("/login", app.LoginUser).Methods("POST")
	return r
}

func main() {
	port := flag.String("port", "8080", "Server Port")
	flag.Parse()

	// Load environment variables from the .env file
	err := godotenv.Load("../.env.local")
	if err != nil {
		log.Println("No .env file found, continuing with environment variables or defaults")
	}

	// Get the necessary environment variables for the database
	dbUsername := getEnv("MYSQL_USER", "root")
	dbPassword := getEnv("MYSQL_PASSWORD", "password")
	dbHostname := getEnv("DB_HOSTNAME", "db")
	dbPort := getEnv("MYSQL_PORT", "3306")
	dbName := getEnv("MYSQL_DATABASE", "db")

	// Initialize the database connection
	db, err := initDB(dbUsername, dbPassword, dbHostname, dbPort, dbName)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Create a logger for the application
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Create an instance of App with the necessary dependencies
	app := &App{
		DB:     db,
		Logger: logger,
	}

	// Configure the router
	r := app.setupRouter()

	// Start the HTTP server
	log.Println("Starting server on port", *port)
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		app.Logger.Fatal(err)
	}
}

// getEnv returns the value of an environment variable or a default value if it is not set
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Use a variable to allow overriding in tests
var sqlOpen = sql.Open

// initDB initializes the connection to the MySQL database
func initDB(username, password, hostname, port, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)

	// Open a connection to the database
	db, err := sqlOpen("mysql", dsn)  // Use the sqlOpen variable here
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Check if the connection is successful
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}
	log.Println("Successfully connected to the MySQL database!")
	return db, nil
}

// Home handles requests to the homepage
func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	app.Logger.Println("Homepage handler called")
	fmt.Fprintf(w, "Homepage")
}

// Info handles requests to the info page
func (app *App) Info(w http.ResponseWriter, r *http.Request) {
	app.Logger.Println("Info handler called")
	fmt.Fprintf(w, "Info page")
}

// HandleError is a method of App that handles errors by logging them and sending an appropriate HTTP response
func (app *App) HandleError(w http.ResponseWriter, message string, err error, status int) {
    app.Logger.Printf("%s: %v", message, err)
    http.Error(w, message, status)
}

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
    // Set Content-Type header to application/json
    w.Header().Set("Content-Type", "application/json")
    // Attempt to encode the payload
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        log.Printf("Error encoding response: %v", err)
        // Set Content-Type header again in case of error
        w.Header().Set("Content-Type", "application/json")
        // Set status to 500
        w.WriteHeader(http.StatusInternalServerError)
        // Write the error message, including a newline
        w.Write([]byte("Error encoding response\n"))
        return
    }
    // Set the status code
    w.WriteHeader(status)
}

// HandleError handles errors by logging them and sending an appropriate HTTP response
func HandleError(w http.ResponseWriter, logger *log.Logger, message string, err error, status int) {
    // Log the error message with additional context
    logger.Printf("%s: %v", message, err)
    // Send an HTTP error response with the given status code
    http.Error(w, message, status)
}

// GetIDFromRequest extracts and validates an ID parameter from the URL
func GetIDFromRequest(r *http.Request, paramName string) (int, error) {
    // Retrieve the parameter from the URL
    vars := mux.Vars(r)
    idStr := vars[paramName]
    // Convert the string parameter to an integer
    id, err := strconv.Atoi(idStr)
    if err != nil {
        // Return an error if conversion fails, using a lowercase message
        return 0, fmt.Errorf("invalid %s: %v", paramName, err) 
    }
    return id, nil
}

// ScanAuthors processes rows from the SQL query and returns a list of authors
func ScanAuthors(rows *sql.Rows) ([]Author, error) {
    defer rows.Close()

    var authors []Author

    for rows.Next() {
        var author Author
        if err := rows.Scan(&author.ID, &author.Lastname, &author.Firstname, &author.Photo); err != nil {
            return nil, err
        }
        authors = append(authors, author)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return authors, nil
}

// ValidateAuthorData checks if the required fields for an author are present
func ValidateAuthorData(author Author) error {
    if author.Firstname == "" || author.Lastname == "" {
        return fmt.Errorf("firstname and lastname are required fields") // Lowercase error message
    }
    return nil
}

// ValidateBookData checks if the required fields for a book are present
func ValidateBookData(book Book) error {
    // Ensure Title and AuthorID are not empty or zero
    if book.Title == "" || book.AuthorID == 0 {
        return fmt.Errorf("title and authorID are required fields") // Lowercase error message
    }
    return nil
}

// SearchAuthors searches for authors based on a query parameter
func (app *App) SearchAuthors(w http.ResponseWriter, r *http.Request) {
    // Log the entry to the handler for debugging purposes
    app.Logger.Println("SearchAuthors handler called")

    // Get the "query" parameter from the URL
    query := r.URL.Query().Get("query")
    if query == "" {
        // If the query parameter is missing, return a 400 Bad Request error
        HandleError(w, app.Logger, "Query parameter is required", nil, http.StatusBadRequest)
        return
    }

    // Prepare the SQL query with an ORDER BY clause to ensure consistent result order
    sqlQuery := `
        SELECT id, Firstname, Lastname, photo 
        FROM authors 
        WHERE Firstname LIKE ? OR Lastname LIKE ?
        ORDER BY Lastname, Firstname
    `

    // Execute the SQL query to fetch authors based on the query parameter
    rows, err := app.DB.Query(sqlQuery, "%"+query+"%", "%"+query+"%")
    if err != nil {
        // Log error executing the query and return a 500 Internal Server Error
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close() // Always close rows after use to release resources

    // Use the utility function to scan and process the SQL rows
    authors, err := ScanAuthors(rows)
    if err != nil {
        // Log error scanning the rows and return a 500 Internal Server Error
        HandleError(w, app.Logger, "Error scanning authors", err, http.StatusInternalServerError)
        return
    }

    // Send the JSON response using the utility function
    RespondWithJSON(w, http.StatusOK, authors)
}

// ScanBooks processes rows from the SQL query and returns a list of books with author information
func ScanBooks(rows *sql.Rows) ([]BookAuthorInfo, error) {
    defer rows.Close()

    var books []BookAuthorInfo

    for rows.Next() {
        var book BookAuthorInfo
        if err := rows.Scan(&book.BookID, &book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
            return nil, err
        }
        books = append(books, book)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return books, nil
}

// SearchBooks searches for books based on a query parameter
func (app *App) SearchBooks(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("SearchBooks handler called")

    query := r.URL.Query().Get("query")
    if query == "" {
        HandleError(w, app.Logger, "Query parameter is required", nil, http.StatusBadRequest)
        return
    }

    sqlQuery := `
        SELECT 
            books.id AS book_id,
            books.title AS book_title, 
            books.author_id AS author_id, 
            books.photo AS book_photo, 
            books.is_borrowed AS is_borrowed, 
            books.details AS book_details,
            authors.Lastname AS author_lastname, 
            authors.Firstname AS author_firstname
        FROM books
        JOIN authors ON books.author_id = authors.id
        WHERE books.title LIKE ? OR authors.Firstname LIKE ? OR authors.Lastname LIKE ?
        ORDER BY books.title, authors.Lastname, authors.Firstname
    `

    rows, err := app.DB.Query(sqlQuery, "%"+query+"%", "%"+query+"%", "%"+query+"%")
    if err != nil {
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close() 

    books, err := ScanBooks(rows)
    if err != nil {
        HandleError(w, app.Logger, "Error scanning books", err, http.StatusInternalServerError)
        return
    }

    RespondWithJSON(w, http.StatusOK, books)
}


// GetAuthors retrieves all authors from the database
func (app *App) GetAuthors(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("GetAuthors handler called")

    sqlQuery := `
        SELECT id, Lastname, Firstname, photo 
        FROM authors
        ORDER BY Lastname, Firstname
    `
    rows, err := app.DB.Query(sqlQuery)
    if err != nil {
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    authors, err := ScanAuthors(rows)
    if err != nil {
        HandleError(w, app.Logger, "Error scanning authors", err, http.StatusInternalServerError)
        return
    }

    RespondWithJSON(w, http.StatusOK, authors)
}

// GetAllBooks retrieves all books from the database along with the author's first and last name
func (app *App) GetAllBooks(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("GetAllBooks handler called")

    query := `
        SELECT 
            books.id AS book_id,
            books.title AS book_title, 
            books.author_id AS author_id, 
            books.photo AS book_photo, 
            books.is_borrowed AS is_borrowed, 
            books.details AS book_details,
            authors.Lastname AS author_lastname, 
            authors.Firstname AS author_firstname
        FROM books
        JOIN authors ON books.author_id = authors.id
        ORDER BY books.title, authors.Lastname, authors.Firstname
    `

    rows, err := app.DB.Query(query)
    if err != nil {
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }

    books, err := ScanBooks(rows)
    if err != nil {
        HandleError(w, app.Logger, "Error scanning books", err, http.StatusInternalServerError)
        return
    }

    RespondWithJSON(w, http.StatusOK, books)
}

// GetAuthorsAndBooks retrieves information about authors and their books
func (app *App) GetAuthorsAndBooks(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("GetAuthorsAndBooks handler called")

    query := `
        SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo
        FROM authors_books ab
        JOIN authors a ON ab.author_id = a.id
        JOIN books b ON ab.book_id = b.id
    `

    rows, err := app.DB.Query(query)
    if err != nil {
        app.Logger.Printf("Query error: %v", err)
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    authorsAndBooks, err := ScanAuthorAndBooks(rows)
    if err != nil {
        app.Logger.Printf("Scan error: %v", err)
        HandleError(w, app.Logger, "Error scanning authors and books", err, http.StatusInternalServerError)
        return
    }

    RespondWithJSON(w, http.StatusOK, authorsAndBooks)
}

// GetAuthorBooksByID retrieves information about an author and their books by the author's ID
func (app *App) GetAuthorBooksByID(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("GetAuthorBooksByID handler called")

    authorID, err := GetIDFromRequest(r, "id")
    if err != nil {
        app.HandleError(w, "Invalid author ID", err, http.StatusBadRequest)
        return
    }

    query := `
        SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo
        FROM authors_books ab
        JOIN authors a ON ab.author_id = a.id
        JOIN books b ON ab.book_id = b.id
        WHERE a.id = ?
    `

    rows, err := app.DB.Query(query, authorID)
    if err != nil {
        app.HandleError(w, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    authorAndBooks, err := ScanAuthorAndBooks(rows)
    if err != nil {
        app.HandleError(w, "Error scanning results", err, http.StatusInternalServerError)
        return
    }

    RespondWithJSON(w, http.StatusOK, authorAndBooks)
}

// ScanAuthorAndBooks processes rows from the SQL query and returns a list of author and book information
func ScanAuthorAndBooks(rows *sql.Rows) ([]AuthorBook, error) {
    var authorsAndBooks []AuthorBook

    for rows.Next() {
        var authorBook AuthorBook
        if err := rows.Scan(&authorBook.AuthorFirstname, &authorBook.AuthorLastname, &authorBook.BookTitle, &authorBook.BookPhoto); err != nil {
            return nil, err 
        }
        authorsAndBooks = append(authorsAndBooks, authorBook)
    }

    if err := rows.Err(); err != nil {
        return nil, err 
    }

    return authorsAndBooks, nil
}

// GetBookByID retrieves information about a specific book based on its ID
func (app *App) GetBookByID(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("GetBookByID handler called")

    bookID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid book ID", err, http.StatusBadRequest)
        return
    }

    query := `
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
        WHERE books.id = ?
    `

    rows, err := app.DB.Query(query, bookID)
    if err != nil {
        HandleError(w, app.Logger, "Error executing query", err, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var book BookAuthorInfo
    if rows.Next() {
        if err := rows.Scan(&book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookID, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
            HandleError(w, app.Logger, "Error scanning book", err, http.StatusInternalServerError)
            return
        }
    } else {
        HandleError(w, app.Logger, "Book not found", nil, http.StatusNotFound)
        return
    }

    RespondWithJSON(w, http.StatusOK, book)
}


// GetSubscribersByBookID retrieves the list of subscribers who have borrowed a specific book based on the book's ID
func (app *App) GetSubscribersByBookID(w http.ResponseWriter, r *http.Request) {
	bookID, err := GetIDFromRequest(r, "id")
	if err != nil {
		HandleError(w, app.Logger, "Invalid book ID", err, http.StatusBadRequest)
		return
	}

	query := `
		SELECT s.Lastname, s.Firstname, s.Email
		FROM subscribers s
		JOIN borrowed_books bb ON s.id = bb.subscriber_id
		WHERE bb.book_id = ?
	`

	rows, err := app.DB.Query(query, bookID)
	if err != nil {
		HandleError(w, app.Logger, "Error querying the database", err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subscribers []Subscriber
	if err := ScanRows(rows, &subscribers); err != nil {
		HandleError(w, app.Logger, "Error scanning subscribers", err, http.StatusInternalServerError)
		return
	}

	if len(subscribers) == 0 {
		http.Error(w, "No subscribers found", http.StatusNotFound)
		return
	}

	RespondWithJSON(w, http.StatusOK, subscribers)
}

func ScanRows(rows *sql.Rows, subscribers *[]Subscriber) error {
    for rows.Next() {
        var subscriber Subscriber
        if err := rows.Scan(&subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
            return err
        }
        *subscribers = append(*subscribers, subscriber)
    }
    return rows.Err()
}

// GetAllSubscribers retrieves all subscribers from the database
func (app *App) GetAllSubscribers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT lastname, firstname, email FROM subscribers"
	
	rows, err := app.DB.Query(query)
	if err != nil {
		HandleError(w, app.Logger, "Error querying the database", err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subscribers []Subscriber
	if err := ScanRows(rows, &subscribers); err != nil {
		HandleError(w, app.Logger, "Error scanning subscribers", err, http.StatusInternalServerError)
		return
	}

	RespondWithJSON(w, http.StatusOK, subscribers)
}

var mkdirAll = os.MkdirAll
var osCreate = os.Create
var ioCopy = io.Copy 

// AddAuthorPhoto handles the upload of an author's photo and updates the database
func (app *App) AddAuthorPhoto(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("AddAuthorPhoto handler called")

    if r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only POST method is supported", nil, http.StatusMethodNotAllowed)
        return
    }

    authorID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid author ID", err, http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("file")
    if err != nil {
        HandleError(w, app.Logger, "Error getting file from request", err, http.StatusInternalServerError)
        return
    }
    defer file.Close()

    filename := "fullsize.jpg"
    ext := filepath.Ext(filename)

    photoDir := "./upload/" + strconv.Itoa(authorID)
    photoPath := photoDir + "/fullsize" + ext

    if err := mkdirAll(photoDir, 0777); err != nil {
        HandleError(w, app.Logger, "Error creating directories", err, http.StatusInternalServerError)
        return
    }

    out, err := osCreate(photoPath) 
    if err != nil {
        HandleError(w, app.Logger, "Error creating file on disk", err, http.StatusInternalServerError)
        return
    }
    defer out.Close()

    if _, err := ioCopy(out, file); err != nil {
        HandleError(w, app.Logger, "Error saving file", err, http.StatusInternalServerError)
        return
    }

    query := `UPDATE authors SET photo = ? WHERE id = ?`
    if _, err := app.DB.Exec(query, photoPath, authorID); err != nil {
        HandleError(w, app.Logger, "Failed to update author photo", err, http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
}

var jsonEncoder = json.NewEncoder

// AddAuthor adds a new author to the database
func (app *App) AddAuthor(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("AddAuthor handler called")

    if r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only POST method is supported", nil, http.StatusMethodNotAllowed)
        return
    }

    var author Author
    if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
        HandleError(w, app.Logger, "Invalid JSON data", err, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := ValidateAuthorData(author); err != nil {
        HandleError(w, app.Logger, err.Error(), nil, http.StatusBadRequest)
        return
    }

    query := `INSERT INTO authors (lastname, firstname, photo) VALUES (?, ?, ?)`
    result, err := app.DB.Exec(query, author.Lastname, author.Firstname, "")
    if err != nil {
        HandleError(w, app.Logger, "Failed to insert author", err, http.StatusInternalServerError)
        return
    }

    id, err := result.LastInsertId()
    if err != nil || id == 0 {  
        HandleError(w, app.Logger, "Failed to get last insert ID", err, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

    response := map[string]int{"id": int(id)}
    if err := jsonEncoder(w).Encode(response); err != nil {
        app.Logger.Printf("JSON encoding error: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}

// Define a global variable for os.MkdirAll to allow overriding in tests
var osMkdirAll = os.MkdirAll

// AddBookPhoto handles the upload of a book's photo and updates the database
func (app *App) AddBookPhoto(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("AddBookPhoto handler called")

    if r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only POST method is supported", nil, http.StatusMethodNotAllowed)
        return
    }

    bookID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid book ID", err, http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("file")
    if err != nil {
        HandleError(w, app.Logger, "Error getting file from request", err, http.StatusInternalServerError)
        return
    }
    defer file.Close()

    photoDir := "./upload/books/" + strconv.Itoa(bookID)
    photoPath := photoDir + "/fullsize.jpg"

    if err := osMkdirAll(photoDir, 0777); err != nil {
        HandleError(w, app.Logger, "Error creating directories", err, http.StatusInternalServerError)
        return
    }

    out, err := osCreate(photoPath)
    if err != nil {
        HandleError(w, app.Logger, "Error creating file on disk", err, http.StatusInternalServerError)
        return
    }
    defer out.Close()

    if _, err := ioCopy(out, file); err != nil {
        HandleError(w, app.Logger, "Error saving file", err, http.StatusInternalServerError)
        return
    }

    query := `UPDATE books SET photo = ? WHERE id = ?`
    if _, err := app.DB.Exec(query, photoPath, bookID); err != nil {
        HandleError(w, app.Logger, "Failed to update book photo", err, http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
}

// AddBook adds a new book to the database
func (app *App) AddBook(w http.ResponseWriter, r *http.Request) {
    app.Logger.Println("AddBook handler called")

    if r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only POST method is supported", nil, http.StatusMethodNotAllowed)
        return
    }

    var book Book
    if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
        HandleError(w, app.Logger, "Invalid JSON data", err, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := ValidateBookData(book); err != nil {
        HandleError(w, app.Logger, err.Error(), nil, http.StatusBadRequest)
        return
    }

    query := `INSERT INTO books (title, photo, details, author_id, is_borrowed) VALUES (?, ?, ?, ?, ?)`
    result, err := app.DB.Exec(query, book.Title, "", book.Details, book.AuthorID, book.IsBorrowed)
    if err != nil {
        HandleError(w, app.Logger, "Failed to insert book", err, http.StatusInternalServerError)
        return
    }

    id, err := result.LastInsertId()
    if err != nil || id == 0 {
        HandleError(w, app.Logger, "Failed to get last insert ID", err, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    response := map[string]int{"id": int(id)}
    if err := json.NewEncoder(w).Encode(response); err != nil {
        app.Logger.Printf("JSON encoding error: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}
// AddSubscriber adds a new subscriber to the database
func (app *App) AddSubscriber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var subscriber Subscriber
	if err := json.NewDecoder(r.Body).Decode(&subscriber); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if subscriber.Firstname == "" || subscriber.Lastname == "" || subscriber.Email == "" {
		http.Error(w, "Firstname, Lastname, and Email are required fields", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO subscribers (lastname, firstname, email) VALUES (?, ?, ?)`
	result, err := app.DB.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email)
	if err != nil {
		app.Logger.Printf("Failed to insert subscriber: %v", err)
		http.Error(w, "Failed to insert subscriber", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		app.Logger.Printf("Failed to get last insert ID: %v", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": int(id)})
}


// BorrowBook handles borrowing a book by a subscriber
func (app *App) BorrowBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		SubscriberID int `json:"subscriber_id"`
		BookID       int `json:"book_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.SubscriberID == 0 || requestBody.BookID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var isBorrowed bool
	err = app.DB.QueryRow("SELECT is_borrowed FROM books WHERE id = ?", requestBody.BookID).Scan(&isBorrowed)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if isBorrowed {
		http.Error(w, "Book is already borrowed", http.StatusConflict)
		return
	}

	_, err = app.DB.Exec("INSERT INTO borrowed_books (subscriber_id, book_id, date_of_borrow) VALUES (?, ?, NOW())", requestBody.SubscriberID, requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec("UPDATE books SET is_borrowed = TRUE WHERE id = ?", requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Book borrowed successfully"}`)
}

// ReturnBorrowedBook handles returning a borrowed book by a subscriber
func (app *App) ReturnBorrowedBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		SubscriberID int `json:"subscriber_id"`
		BookID       int `json:"book_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.SubscriberID == 0 || requestBody.BookID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var isBorrowed bool
	err = app.DB.QueryRow("SELECT is_borrowed FROM books WHERE id = ?", requestBody.BookID).Scan(&isBorrowed)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			app.Logger.Printf("Database error: %v", err)
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !isBorrowed {
		http.Error(w, "Book is not borrowed", http.StatusBadRequest)
		return
	}

	_, err = app.DB.Exec("UPDATE borrowed_books SET return_date = NOW() WHERE subscriber_id = ? AND book_id = ?", requestBody.SubscriberID, requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}


	_, err = app.DB.Exec("UPDATE books SET is_borrowed = FALSE WHERE id = ?", requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Book returned successfully"}`)
}

// UpdateAuthor updates an existing author in the database
func (app *App) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut && r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only PUT or POST methods are supported", nil, http.StatusMethodNotAllowed)
        return
    }

    authorID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid author ID", err, http.StatusBadRequest)
        return
    }

    var author Author
    if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
        HandleError(w, app.Logger, "Invalid JSON data", err, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if author.Firstname == "" || author.Lastname == "" {
        HandleError(w, app.Logger, "Firstname and Lastname are required fields", nil, http.StatusBadRequest)
        return
    }

    query := `
        UPDATE authors 
        SET lastname = ?, firstname = ?, photo = ? 
        WHERE id = ?
    `

    result, err := app.DB.Exec(query, author.Lastname, author.Firstname, author.Photo, authorID)
    if err != nil {
        HandleError(w, app.Logger, "Failed to update author", err, http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        HandleError(w, app.Logger, "Failed to retrieve affected rows", err, http.StatusInternalServerError)
        return
    }
    if rowsAffected == 0 {
        HandleError(w, app.Logger, "Author not found", nil, http.StatusNotFound)
        return
    }

    RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Author updated successfully"})
}

// UpdateBook handles the updating of an existing book in the database
func (app *App) UpdateBook(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut && r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only PUT or POST methods are supported", nil, http.StatusMethodNotAllowed)
        return
    }

    bookID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid book ID", err, http.StatusBadRequest)
        return
    }

    var book Book
    if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
        HandleError(w, app.Logger, "Invalid JSON data", err, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := ValidateBookData(book); err != nil {
        HandleError(w, app.Logger, err.Error(), nil, http.StatusBadRequest)
        return
    }

    query := `
        UPDATE books 
        SET title = ?, author_id = ?, photo = ?, details = ?, is_borrowed = ?
        WHERE id = ?
    `

    result, err := app.DB.Exec(query, book.Title, book.AuthorID, book.Photo, book.Details, book.IsBorrowed, bookID)
    if err != nil {
        HandleError(w, app.Logger, "Failed to update book", err, http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        HandleError(w, app.Logger, "Failed to retrieve affected rows", err, http.StatusInternalServerError)
        return
    }
    if rowsAffected == 0 {
        HandleError(w, app.Logger, "Book not found", nil, http.StatusNotFound)
        return
    }

    RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Book updated successfully"})
}


// UpdateSubscriber updates an existing subscriber in the database
func (app *App) UpdateSubscriber(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut && r.Method != http.MethodPost {
        HandleError(w, app.Logger, "Only PUT or POST methods are supported", nil, http.StatusMethodNotAllowed)
        return
    }

    subscriberID, err := GetIDFromRequest(r, "id")
    if err != nil {
        HandleError(w, app.Logger, "Invalid subscriber ID", err, http.StatusBadRequest)
        return
    }

    var subscriber Subscriber
    if err := json.NewDecoder(r.Body).Decode(&subscriber); err != nil {
        HandleError(w, app.Logger, "Invalid JSON data", err, http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := ValidateSubscriberData(subscriber); err != nil {
        HandleError(w, app.Logger, err.Error(), nil, http.StatusBadRequest)
        return
    }

    query := `
        UPDATE subscribers 
        SET lastname = ?, firstname = ?, email = ?
        WHERE id = ?
    `

    result, err := app.DB.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email, subscriberID)
    if err != nil {
        HandleError(w, app.Logger, "Failed to update subscriber", err, http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        HandleError(w, app.Logger, "Failed to retrieve affected rows", err, http.StatusInternalServerError)
        return
    }
    if rowsAffected == 0 {
        HandleError(w, app.Logger, "Subscriber not found", nil, http.StatusNotFound)
        return
    }

    RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Subscriber updated successfully"})
}

// ValidateSubscriberData checks if the required fields for a subscriber are present
func ValidateSubscriberData(subscriber Subscriber) error {
    if subscriber.Firstname == "" || subscriber.Lastname == "" || subscriber.Email == "" {
        return fmt.Errorf("firstname, lastname, and email are required fields")
    }
    return nil
}


// DeleteAuthor deletes an existing author from the database
func (app *App) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	authorID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	var numBooks int
	err = app.DB.QueryRow(`SELECT COUNT(*) FROM books WHERE author_id = ?`, authorID).Scan(&numBooks)
	if err != nil {
		http.Error(w, "Failed to check for books", http.StatusInternalServerError)
		return
	}

	if numBooks > 0 {
		http.Error(w, "Author has associated books, delete books first", http.StatusBadRequest)
		return
	}

	result, err := app.DB.Exec(`DELETE FROM authors WHERE id = ?`, authorID)
	if err != nil {
		http.Error(w, "Failed to delete author", http.StatusInternalServerError)
		return
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "Author deleted successfully")
}


// DeleteBook deletes an existing book from the database
func (app *App) DeleteBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract the book ID from the URL path
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Query to get the author ID of the book
	authorIDQuery := `
        SELECT author_id
        FROM books
        WHERE id = ?
    `

	// Execute the query
	var authorID int
	err = app.DB.QueryRow(authorIDQuery, bookID).Scan(&authorID)
	if err != nil {
		app.Logger.Printf("Failed to retrieve author ID: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve author ID: %v", err), http.StatusInternalServerError)
		return
	}

	// Query to check if the author has any other books
	otherBooksQuery := `
        SELECT COUNT(*)
        FROM books
        WHERE author_id = ? AND id != ?
    `

	// Execute the query
	var numOtherBooks int
	err = app.DB.QueryRow(otherBooksQuery, authorID, bookID).Scan(&numOtherBooks)
	if err != nil {
		app.Logger.Printf("Failed to check for other books: %v", err)
		http.Error(w, fmt.Sprintf("Failed to check for other books: %v", err), http.StatusInternalServerError)
		return
	}

	// Query to delete the book
	deleteBookQuery := `
        DELETE FROM books
        WHERE id = ?
    `

	// Execute the query to delete the book
	result, err := app.DB.Exec(deleteBookQuery, bookID)
	if err != nil {
		app.Logger.Printf("Failed to delete book: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete book: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any row was actually deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// If the author has no other books, delete the author as well
	if numOtherBooks == 0 {
		deleteAuthorQuery := `
            DELETE FROM authors
            WHERE id = ?
        `

		// Execute the query to delete the author
		_, err = app.DB.Exec(deleteAuthorQuery, authorID)
		if err != nil {
			app.Logger.Printf("Failed to delete author: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete author: %v", err), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "Book deleted successfully")
}

// DeleteSubscriber deletes an existing subscriber from the database
func (app *App) DeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract the subscriber ID from the URL path
	vars := mux.Vars(r)
	subscriberID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscriber ID", http.StatusBadRequest)
		return
	}

	// Query to delete the subscriber
	deleteQuery := `
        DELETE FROM subscribers
        WHERE id = ?
    `

	// Execute the query to delete the subscriber
	result, err := app.DB.Exec(deleteQuery, subscriberID)
	if err != nil {
		app.Logger.Printf("Failed to delete subscriber: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete subscriber: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any row was actually deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Subscriber not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Subscriber deleted successfully")
}
