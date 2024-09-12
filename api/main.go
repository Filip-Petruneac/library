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
	port := flag.String("port", "8081", "Server Port")
	flag.Parse()

	// Load environment variables from the .env file
	err := godotenv.Load()
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

// SearchAuthors searches for authors based on a query parameter
func (app *App) SearchAuthors(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		app.Logger.Println("Query parameter is required") 
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	sqlQuery := `
		SELECT id, Firstname, Lastname, photo 
		FROM authors 
		WHERE Firstname LIKE ? OR Lastname LIKE ?
	`

	rows, err := app.DB.Query(sqlQuery, "%"+query+"%", "%"+query+"%")
	if err != nil {
		app.Logger.Printf("Error executing query: %v", err)
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var authors []AuthorInfo
	for rows.Next() {
		var author AuthorInfo
		if err := rows.Scan(&author.ID, &author.Firstname, &author.Lastname, &author.Photo); err != nil {
			app.Logger.Printf("Error scanning row: %v", err)
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, "Error fetching results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authors); err != nil {
		app.Logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// SearchBooks searches for books based on a query parameter
func (app *App) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is missing", http.StatusBadRequest)
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
	`

	rows, err := app.DB.Query(sqlQuery, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		app.Logger.Printf("Error executing query: %v", err)
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []BookAuthorInfo
	for rows.Next() {
		var book BookAuthorInfo
		if err := rows.Scan(&book.BookID, &book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
			app.Logger.Printf("Error scanning row: %v", err)
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, "Error fetching results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		app.Logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// GetAuthors retrieves all authors from the database
func (app *App) GetAuthors(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, lastname, firstname, photo FROM authors")
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var author Author
		if err := rows.Scan(&author.ID, &author.Lastname, &author.Firstname, &author.Photo); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authors); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetAllBooks retrieves all books from the database along with the author's first and last name
func (app *App) GetAllBooks(w http.ResponseWriter, r *http.Request) {
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
	`

	rows, err := app.DB.Query(query)
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []BookAuthorInfo
	for rows.Next() {
		var book BookAuthorInfo
		if err := rows.Scan(&book.BookID, &book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetAuthorsAndBooks retrieves information about authors and their books
func (app *App) GetAuthorsAndBooks(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo
		FROM authors_books ab
		JOIN authors a ON ab.author_id = a.id
		JOIN books b ON ab.book_id = b.id
	`
	rows, err := app.DB.Query(query)
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var authorsAndBooks []AuthorBook
	for rows.Next() {
		var authorBook AuthorBook
		if err := rows.Scan(&authorBook.AuthorFirstname, &authorBook.AuthorLastname, &authorBook.BookTitle, &authorBook.BookPhoto); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		authorsAndBooks = append(authorsAndBooks, authorBook)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authorsAndBooks); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetAuthorBooksByID retrieves information about an author and their books by the author's ID
func (app *App) GetAuthorBooksByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authorID := vars["id"]
	id, err := strconv.Atoi(authorID)
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, a.Photo AS author_photo, b.title AS book_title, b.photo AS book_photo
		FROM authors_books ab
		JOIN authors a ON ab.author_id = a.id
		JOIN books b ON ab.book_id = b.id
		WHERE a.id = ?
	`

	rows, err := app.DB.Query(query, id)
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var authorFirstname, authorLastname, authorPhoto string
	var books []AuthorBook

	for rows.Next() {
		var bookTitle, bookPhoto string
		if err := rows.Scan(&authorFirstname, &authorLastname, &authorPhoto, &bookTitle, &bookPhoto); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, AuthorBook{
			BookTitle: bookTitle,
			BookPhoto: bookPhoto,
		})
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authorAndBooks := struct {
		AuthorFirstname string       `json:"author_firstname"`
		AuthorLastname  string       `json:"author_lastname"`
		AuthorPhoto     string       `json:"author_photo"`
		Books           []AuthorBook `json:"books"`
	}{
		AuthorFirstname: authorFirstname,
		AuthorLastname:  authorLastname,
		AuthorPhoto:     authorPhoto,
		Books:           books,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authorAndBooks); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetBookByID retrieves information about a specific book based on its ID
func (app *App) GetBookByID(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]
	intBookID, err := strconv.Atoi(bookID)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
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

	rows, err := app.DB.Query(query, intBookID)
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var book BookAuthorInfo
	if rows.Next() {
		if err := rows.Scan(&book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookID, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(book); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetSubscribersByBookID retrieves the list of subscribers who have borrowed a specific book based on the book's ID
func (app *App) GetSubscribersByBookID(w http.ResponseWriter, r *http.Request) {
	bookID := mux.Vars(r)["id"]
	if bookID == "" {
		http.Error(w, "Missing book ID parameter", http.StatusBadRequest)
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
		app.Logger.Printf("Error querying the database: %v", err)
		http.Error(w, "Error querying the database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subscribers []Subscriber
	for rows.Next() {
		var subscriber Subscriber
		if err := rows.Scan(&subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		subscribers = append(subscribers, subscriber)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Row iteration error: %v", err)
		http.Error(w, "Row error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(subscribers) == 0 {
		http.Error(w, "No subscribers found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subscribers); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetAllSubscribers retrieves all subscribers from the database
func (app *App) GetAllSubscribers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT lastname, firstname, email FROM subscribers"
	rows, err := app.DB.Query(query)
	if err != nil {
		app.Logger.Printf("Query error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subscribers []Subscriber
	for rows.Next() {
		var subscriber Subscriber
		if err := rows.Scan(&subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
			app.Logger.Printf("Scan error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subscribers = append(subscribers, subscriber)
	}

	if err := rows.Err(); err != nil {
		app.Logger.Printf("Rows error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subscribers); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddAuthorPhoto handles the upload of an author's photo and updates the database
func (app *App) AddAuthorPhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		app.Logger.Printf("Error getting file from request: %v", err)
		http.Error(w, "Error getting file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := "fullsize.jpg" // Default or extracted filename
	ext := filepath.Ext(filename)

	photoDir := "./upload/" + strconv.Itoa(authorID)
	photoPath := photoDir + "/fullsize" + ext

	err = os.MkdirAll(photoDir, 0777)
	if err != nil {
		app.Logger.Printf("Error creating directories: %v", err)
		http.Error(w, "Unable to create the directories on disk", http.StatusInternalServerError)
		return
	}

	out, err := os.Create(photoPath)
	if err != nil {
		app.Logger.Printf("Error creating file on disk: %v", err)
		http.Error(w, "Unable to create the file on disk", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		app.Logger.Printf("Error saving file: %v", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	query := `UPDATE authors SET photo = ? WHERE id = ?`
	_, err = app.DB.Exec(query, photoPath, authorID)
	if err != nil {
		app.Logger.Printf("Failed to update author photo: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update author photo: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
}

// AddAuthor adds a new author to the database
func (app *App) AddAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON data received from the request
	var author Author
	err := json.NewDecoder(r.Body).Decode(&author)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check if all required fields are filled
	if author.Firstname == "" || author.Lastname == "" {
		http.Error(w, "Firstname and Lastname are required fields", http.StatusBadRequest)
		return
	}

	// Query to add author with photo path
	query := `INSERT INTO authors (lastname, firstname, photo) VALUES (?, ?, ?)`

	// Execute the query
	result, err := app.DB.Exec(query, author.Lastname, author.Firstname, "")
	if err != nil {
		app.Logger.Printf("Failed to insert author: %v", err)
		http.Error(w, fmt.Sprintf("Failed to insert author: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the inserted author ID
	id, err := result.LastInsertId()
	if err != nil {
		app.Logger.Printf("Failed to get last insert ID: %v", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the response with the inserted author ID
	response := map[string]int{"id": int(id)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// AddBookPhoto handles the upload of a book's photo and updates the database
func (app *App) AddBookPhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		app.Logger.Printf("Error getting file: %v", err)
		http.Error(w, "Error getting file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filename := header.Filename
	ext := filepath.Ext(filename)
	photoDir := "./upload/books/" + strconv.Itoa(bookID)
	photoPath := photoDir + "/fullsize" + ext

	err = os.MkdirAll(photoDir, 0777)
	if err != nil {
		app.Logger.Printf("Error creating directories: %v", err)
		http.Error(w, "Unable to create directories on disk", http.StatusInternalServerError)
		return
	}

	out, err := os.Create(photoPath)
	if err != nil {
		app.Logger.Printf("Error creating file on disk: %v", err)
		http.Error(w, "Unable to create file on disk", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		app.Logger.Printf("Error saving file: %v", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	query := `UPDATE books SET photo = ? WHERE id = ?`
	_, err = app.DB.Exec(query, photoPath, bookID)
	if err != nil {
		app.Logger.Printf("Failed to update book: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
}

// AddBook adds a new book to the database
func (app *App) AddBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON data received from the request
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check if all required fields are filled
	if book.Title == "" || book.AuthorID == 0 {
		http.Error(w, "Title and AuthorID are required fields", http.StatusBadRequest)
		return
	}

	// Query to add book with photo path
	query := `INSERT INTO books (title, photo, details, author_id, is_borrowed) VALUES (?, ?, ?, ?, ?)`

	// Execute the query
	result, err := app.DB.Exec(query, book.Title, "", book.Details, book.AuthorID, book.IsBorrowed)
	if err != nil {
		app.Logger.Printf("Failed to insert book: %v", err)
		http.Error(w, fmt.Sprintf("Failed to insert book: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the inserted book ID
	id, err := result.LastInsertId()
	if err != nil {
		app.Logger.Printf("Failed to get last insert ID: %v", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return the response with the inserted book ID
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

	// Parse the JSON data received from the request
	var subscriber Subscriber
	err := json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check if all required fields are filled
	if subscriber.Firstname == "" || subscriber.Lastname == "" || subscriber.Email == "" {
		http.Error(w, "Firstname, Lastname, and Email are required fields", http.StatusBadRequest)
		return
	}

	// Query to add subscriber
	query := `INSERT INTO subscribers (lastname, firstname, email) VALUES (?, ?, ?)`

	// Execute the query
	result, err := app.DB.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email)
	if err != nil {
		app.Logger.Printf("Failed to insert subscriber: %v", err)
		http.Error(w, fmt.Sprintf("Failed to insert subscriber: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the ID of the inserted subscriber
	id, err := result.LastInsertId()
	if err != nil {
		app.Logger.Printf("Failed to get last insert ID: %v", err)
		http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return the response with the subscriber ID inserted
	response := map[string]int{"id": int(id)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.Logger.Printf("JSON encoding error: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// BorrowBook handles borrowing a book by a subscriber
func (app *App) BorrowBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body to get subscriber ID and book ID
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

	// Check if the book is already borrowed
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

	// Insert a new record in the borrowed_books table
	_, err = app.DB.Exec("INSERT INTO borrowed_books (subscriber_id, book_id, date_of_borrow) VALUES (?, ?, NOW())", requestBody.SubscriberID, requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the is_borrowed status of the book
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

	// Parse the request body to get subscriber ID and book ID
	var requestBody struct {
		SubscriberID int `json:"subscriber_id"`
		BookID       int `json:"book_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the book is actually borrowed by the subscriber
	var isBorrowed bool
	err = app.DB.QueryRow("SELECT is_borrowed FROM books WHERE id = ?", requestBody.BookID).Scan(&isBorrowed)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			app.Logger.Printf("Database error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !isBorrowed {
		http.Error(w, "Book is not borrowed", http.StatusBadRequest)
		return
	}

	// Update borrowed_books table to mark book as returned
	_, err = app.DB.Exec("UPDATE borrowed_books SET return_date = NOW() WHERE subscriber_id = ? AND book_id = ?", requestBody.SubscriberID, requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update books table to mark book as not borrowed
	_, err = app.DB.Exec("UPDATE books SET is_borrowed = FALSE WHERE id = ?", requestBody.BookID)
	if err != nil {
		app.Logger.Printf("Database error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Book returned successfully")
}

// UpdateAuthor updates an existing author in the database
func (app *App) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Only PUT or POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	var author Author
	err = json.NewDecoder(r.Body).Decode(&author)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if author.Firstname == "" || author.Lastname == "" {
		http.Error(w, "Firstname and Lastname are required fields", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE authors 
		SET lastname = ?, firstname = ?, photo = ? 
		WHERE id = ?
	`

	result, err := app.DB.Exec(query, author.Lastname, author.Firstname, author.Photo, authorID)
	if err != nil {
		app.Logger.Printf("Failed to update author: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update author: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Author updated successfully")
}

// UpdateBook updates an existing book in the database
func (app *App) UpdateBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Only PUT or POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract the book ID from the URL path
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Parse the JSON data received from the request
	var book struct {
		Title      string `json:"title"`
		AuthorID   int    `json:"author_id"`
		Photo      string `json:"photo"`
		Details    string `json:"details"`
		IsBorrowed bool   `json:"is_borrowed"`
	}
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check if all required fields are filled
	if book.Title == "" || book.AuthorID == 0 {
		http.Error(w, "Title and AuthorID are required fields", http.StatusBadRequest)
		return
	}

	// Query to update the book
	query := `
		UPDATE books 
		SET title = ?, author_id = ?, photo = ?, details = ?, is_borrowed = ? 
		WHERE id = ?
	`

	// Execute the query
	result, err := app.DB.Exec(query, book.Title, book.AuthorID, book.Photo, book.Details, book.IsBorrowed, bookID)
	if err != nil {
		app.Logger.Printf("Failed to update book: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any row was actually updated
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Book updated successfully")
}

// UpdateSubscriber updates an existing subscriber in the database
func (app *App) UpdateSubscriber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Only PUT or POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract the subscriber ID from the URL path
	vars := mux.Vars(r)
	subscriberID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscriber ID", http.StatusBadRequest)
		return
	}

	// Parse the JSON data received from the request
	var subscriber Subscriber
	err = json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log the subscriber ID and received data for update
	app.Logger.Printf("Updating subscriber with ID: %d", subscriberID)
	app.Logger.Printf("Received data: %+v", subscriber)

	// Check if all required fields are filled
	if subscriber.Firstname == "" || subscriber.Lastname == "" || subscriber.Email == "" {
		http.Error(w, "Firstname, Lastname, and Email are required fields", http.StatusBadRequest)
		return
	}

	// Query to update the subscriber
	query := `
        UPDATE subscribers 
        SET lastname = ?, firstname = ?, email = ? 
        WHERE id = ?
    `

	// Execute the query
	result, err := app.DB.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email, subscriberID)
	if err != nil {
		app.Logger.Printf("Failed to update subscriber: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update subscriber: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any row was actually updated
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Subscriber not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Subscriber updated successfully")
}

// DeleteAuthor deletes an existing author from the database
func (app *App) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract the author ID from the URL path
	vars := mux.Vars(r)
	authorID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	// Query to check if the author has books
	booksQuery := `
        SELECT COUNT(*)
        FROM books
        WHERE author_id = ?
    `

	// Execute the query
	var numBooks int
	err = app.DB.QueryRow(booksQuery, authorID).Scan(&numBooks)
	if err != nil {
		app.Logger.Printf("Failed to check for books: %v", err)
		http.Error(w, fmt.Sprintf("Failed to check for books: %v", err), http.StatusInternalServerError)
		return
	}

	// If author has books, respond with a bad request
	if numBooks > 0 {
		http.Error(w, "Author has associated books, delete books first", http.StatusBadRequest)
		return
	}

	// Query to delete the author
	deleteQuery := `
        DELETE FROM authors
        WHERE id = ?
    `

	// Execute the query to delete the author
	result, err := app.DB.Exec(deleteQuery, authorID)
	if err != nil {
		app.Logger.Printf("Failed to delete author: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete author: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any row was actually deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Author deleted successfully")
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
