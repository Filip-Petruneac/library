package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

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

func main() {
	port := flag.String("port", "8081", "Server Port")
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with environment variables or defaults")
	}

	dbUsername := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHostname := getEnv("DB_HOSTNAME", "db")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "db")

	db, err := initDB(dbUsername, dbPassword, dbHostname, dbPort, dbName)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	log.Println("Starting our server on port", *port)

	r := setupRouter(db)

	log.Println("Started on port", *port)
	fmt.Println("To close connection CTRL+C :-)")

	// Spinning up the server.
	err = http.ListenAndServe(":"+*port, r)
	if err != nil {
		log.Fatal(err)
	}
}

func setupRouter(db *sql.DB) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", Home)
	r.HandleFunc("/info", Info)
	r.HandleFunc("/books", GetAllBooks(db)).Methods("GET")
	r.HandleFunc("/authors", GetAuthors(db)).Methods("GET")
	r.HandleFunc("/authorsbooks", GetAuthorsAndBooks(db)).Methods("GET")
	r.HandleFunc("/authors/{id}", GetAuthorBooksByID(db)).Methods("GET")
	r.HandleFunc("/books/{id}", GetBookByID(db)).Methods("GET")
	r.HandleFunc("/subscribers/{id}", GetSubscribersByBookID(db)).Methods("GET")
	r.HandleFunc("/subscribers", GetAllSubscribers(db)).Methods("GET")
	r.HandleFunc("/book/borrow", BorrowBook(db)).Methods("POST")
	r.HandleFunc("/book/return", ReturnBorrowedBook(db)).Methods("POST")
	r.HandleFunc("/authors/new", AddAuthor(db)).Methods("POST")
	r.HandleFunc("/author/photo/{id}", AddAuthorPhoto(db)).Methods("POST")
	r.HandleFunc("/books/new", AddBook(db)).Methods("POST")
	r.HandleFunc("/books/photo/{id}", AddBookPhoto(db)).Methods("POST")
	r.HandleFunc("/subscribers/new", AddSubscriber(db)).Methods("POST")
	r.HandleFunc("/authors/{id}", UpdateAuthor(db)).Methods("PUT", "POST")
	r.HandleFunc("/books/{id}", UpdateBook(db)).Methods("PUT", "POST")
	r.HandleFunc("/subscribers/{id}", UpdateSubscriber(db)).Methods("PUT", "POST")
	r.HandleFunc("/authors/{id}", DeleteAuthor(db)).Methods("DELETE")
	r.HandleFunc("/books/{id}", DeleteBook(db)).Methods("DELETE")
	r.HandleFunc("/subscribers/{id}", DeleteSubscriber(db)).Methods("DELETE")
	r.HandleFunc("/search_books", SearchBooks(db)).Methods("GET")
	r.HandleFunc("/search_authors", SearchAuthors(db)).Methods("GET")
	r.HandleFunc("/singup", signupUser(db)).Methods("POST")
	r.HandleFunc("/login", loginUser(db)).Methods("POST")

	return r
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func initDB(username, password, hostname, port, dbname string) (*sql.DB, error) {
	var err error

	// Constructing the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Connected to the MySQL database!")
	return db, nil
}

// Home handles requests to the homepage
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}

// Info handles requests to the info page
func Info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Info page")
}

// GetAllBooks returns a handler that gets all the books in the database along with the author's first and last name.
func GetAllBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var books []BookAuthorInfo
		for rows.Next() {
			var book BookAuthorInfo
			if err := rows.Scan(&book.BookID, &book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			books = append(books, book)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(books)
	}
}

func SearchBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query(sqlQuery, "%"+query+"%", "%"+query+"%", "%"+query+"%")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var books []BookAuthorInfo
		for rows.Next() {
			var book BookAuthorInfo
			if err := rows.Scan(&book.BookID, &book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			books = append(books, book)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(books)
	}
}

func SearchAuthors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`SELECT id, Firstname, Lastname, photo FROM authors WHERE Firstname LIKE ? OR Lastname LIKE ?`, "%"+query+"%", "%"+query+"%")
		if err != nil {
			http.Error(w, "Error executing query", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		authors := []AuthorInfo{}
		for rows.Next() {
			var author AuthorInfo
			if err := rows.Scan(&author.ID, &author.Firstname, &author.Lastname, &author.Photo); err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			authors = append(authors, author)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(authors); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

func GetAuthors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, lastname, firstname, photo FROM authors")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var authors []Author
		for rows.Next() {
			var author Author
			if err := rows.Scan(&author.ID, &author.Lastname, &author.Firstname, &author.Photo); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			authors = append(authors, author)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(authors)
	}
}

// GetAuthorsAndBooks returns a handler function that retrieves information about authors and their books.
func GetAuthorsAndBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title, b.photo AS book_photo
			FROM authors_books ab
			JOIN authors a ON ab.author_id = a.id
			JOIN books b ON ab.book_id = b.id
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var authorsAndBooks []AuthorBook
		for rows.Next() {
			var authorFirstname, authorLastname, bookTitle, bookPhoto string
			if err := rows.Scan(&authorFirstname, &authorLastname, &bookTitle, &bookPhoto); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			authorBook := AuthorBook{
				AuthorFirstname: authorFirstname,
				AuthorLastname:  authorLastname,
				BookTitle:       bookTitle,
				BookPhoto:       bookPhoto,
			}

			authorsAndBooks = append(authorsAndBooks, authorBook)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(authorsAndBooks)
	}
}

// GetAuthorBooksByID returns a handler function that retrieves information about an author and their books by the author's ID.
func GetAuthorBooksByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		rows, err := db.Query(query, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var authorFirstname, authorLastname, authorPhoto, bookTitle, bookPhoto string
		var books []AuthorBook

		for rows.Next() {
			if err := rows.Scan(&authorFirstname, &authorLastname, &authorPhoto, &bookTitle, &bookPhoto); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			books = append(books, AuthorBook{
				BookTitle: bookTitle,
				BookPhoto: bookPhoto,
			})
		}

		if err := rows.Err(); err != nil {
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

		err = json.NewEncoder(w).Encode(authorAndBooks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// GetBookById retrieves information about a specific book based on its ID
func GetBookByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		rows, err := db.Query(query, intBookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var books []BookAuthorInfo
		for rows.Next() {
			var book BookAuthorInfo
			if err := rows.Scan(&book.BookTitle, &book.AuthorID, &book.BookPhoto, &book.IsBorrowed, &book.BookID, &book.BookDetails, &book.AuthorLastname, &book.AuthorFirstname); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			books = append(books, book)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(books) == 0 {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(books[0])
	}
}

func GetSubscribersByBookID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		rows, err := db.Query(query, bookID)
		if err != nil {
			http.Error(w, "Error querying the database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var subscribers []Subscriber
		for rows.Next() {
			var subscriber Subscriber
			if err := rows.Scan(&subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
				http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
				return
			}
			subscribers = append(subscribers, subscriber)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Row error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if len(subscribers) == 0 {
			http.Error(w, "No subscribers found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(subscribers)
	}
}

// GetAllSubscribers returns a handler that gets all the subscribers in the database.
func GetAllSubscribers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT lastname, firstname, email FROM subscribers"
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Query error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var subscribers []Subscriber
		for rows.Next() {
			var subscriber Subscriber
			if err := rows.Scan(&subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
				log.Printf("Scan error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			subscribers = append(subscribers, subscriber)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Rows error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(subscribers)
		if err != nil {
			log.Printf("JSON encoding error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func AddAuthorPhoto(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			fmt.Printf("Error getting file from request: %v\n", err)
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
			fmt.Printf("Error creating directories: %v\n", err)
			http.Error(w, "Unable to create the directories on disk", http.StatusInternalServerError)
			return
		}

		out, err := os.Create(photoPath)
		if err != nil {
			fmt.Printf("Error creating file on disk: %v\n", err)
			http.Error(w, "Unable to create the file on disk", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Printf("Error saving file: %v\n", err)
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		query := `UPDATE authors SET photo = ? WHERE id = ?`
		_, err = db.Exec(query, photoPath, authorID)
		if err != nil {
			fmt.Printf("Failed to update author photo: %v\n", err)
			http.Error(w, fmt.Sprintf("Failed to insert author: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
	}
}

// AddAuthor adds a new author to the database
func AddAuthor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		// We parse the JSON data received from the request
		var author Author
		err := json.NewDecoder(r.Body).Decode(&author)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// We check if all required fields are filled
		if author.Firstname == "" || author.Lastname == "" {
			http.Error(w, "Firstname and Lastname are required fields", http.StatusBadRequest)
			return
		}

		// Query to add author with photo path
		query := `INSERT INTO authors (lastname, firstname, photo) VALUES (?, ?, ?)`

		// We run the query
		result, err := db.Exec(query, author.Lastname, author.Firstname, "")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert author: %v", err), http.StatusInternalServerError)
			return
		}

		// We get the inserted author ID
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// We return the response with the author ID inserted
		response := map[string]int{"id": int(id)}
		json.NewEncoder(w).Encode(response)
	}
}

func AddBookPhoto(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Unable to create directories on disk", http.StatusInternalServerError)
			return
		}

		out, err := os.Create(photoPath)
		if err != nil {
			http.Error(w, "Unable to create file on disk", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		query := `UPDATE books SET photo = ? WHERE id = ?`
		_, err = db.Exec(query, photoPath, bookID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "File uploaded successfully: %s\n", photoPath)
	}
}

func AddBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		var book Book
		err := json.NewDecoder(r.Body).Decode(&book)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if book.Title == "" || book.AuthorID == 0 {
			http.Error(w, "Title and AuthorID are required fields", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO books (title, photo, details, author_id, is_borrowed) VALUES (?, ?, ?, ?, ?)`
		result, err := db.Exec(query, book.Title, "", book.Details, book.AuthorID, book.IsBorrowed)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert book: %v", err), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]int{"id": int(id)}
		json.NewEncoder(w).Encode(response)
	}
}

// AddSubscriber adds a new subscriber to the database
func AddSubscriber(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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
		query := `
			INSERT INTO subscribers (lastname, firstname, email) 
			VALUES (?, ?, ?)
		`

		// Execute the query
		result, err := db.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert subscriber: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the ID of the inserted subscriber
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
			return
		}

		// Return the response with the subscriber ID inserted
		response := map[string]int{"id": int(id)}
		json.NewEncoder(w).Encode(response)
	}
}

// BorrowBook handles borrowing a book by a subscriber
func BorrowBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// Check if the book is already borrowed
		var isBorrowed bool
		err = db.QueryRow("SELECT is_borrowed FROM books WHERE id = ?", requestBody.BookID).Scan(&isBorrowed)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if isBorrowed {
			http.Error(w, "Book is already borrowed", http.StatusConflict)
			return
		}

		// Insert a new record in the borrowed_books table
		_, err = db.Exec("INSERT INTO borrowed_books (subscriber_id, book_id, date_of_borrow) VALUES (?, ?, NOW())", requestBody.SubscriberID, requestBody.BookID)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the is_borrowed status of the book
		_, err = db.Exec("UPDATE books SET is_borrowed = TRUE WHERE id = ?", requestBody.BookID)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Book borrowed successfully"}`)
	}
}

// ReturnBorrowedBook handles returning a borrowed book by a subscriber
func ReturnBorrowedBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		err = db.QueryRow("SELECT is_borrowed FROM books WHERE id = ? AND is_borrowed = TRUE", requestBody.BookID).Scan(&isBorrowed)
		if err != nil {
			http.Error(w, "Book is not borrowed", http.StatusNotFound)
			return
		}

		// Update borrowed_books table to mark book as returned
		_, err = db.Exec("UPDATE borrowed_books SET return_date = NOW() WHERE subscriber_id = ? AND book_id = ?", requestBody.SubscriberID, requestBody.BookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update books table to mark book as not borrowed
		_, err = db.Exec("UPDATE books SET is_borrowed = FALSE WHERE id = ?", requestBody.BookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Book returned successfully")
	}
}

func UpdateAuthor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		result, err := db.Exec(query, author.Lastname, author.Firstname, author.Photo, authorID)
		if err != nil {
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
}

// UpdateBook updates an existing book in the database
func UpdateBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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

		// Log the book ID and received data for update
		log.Printf("Updating book with ID: %d", bookID)
		log.Printf("Received data: %+v", book)

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
		result, err := db.Exec(query, book.Title, book.AuthorID, book.Photo, book.Details, book.IsBorrowed, bookID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update book: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if any row was actually updated
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		// Return the success response
		fmt.Fprintf(w, "Book updated successfully")
	}
}

// UpdateSubscriber updates an existing subscriber in the database
func UpdateSubscriber(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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
		log.Printf("Updating subscriber with ID: %d", subscriberID)
		log.Printf("Received data: %+v", subscriber)

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
		result, err := db.Exec(query, subscriber.Lastname, subscriber.Firstname, subscriber.Email, subscriberID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update subscriber: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if any row was actually updated
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Subscriber not found", http.StatusNotFound)
			return
		}

		// Return the success response
		fmt.Fprintf(w, "Subscriber updated successfully")
	}
}

// DeleteAuthor deletes an existing author from the database
func DeleteAuthor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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
		err = db.QueryRow(booksQuery, authorID).Scan(&numBooks)
		if err != nil {
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
		result, err := db.Exec(deleteQuery, authorID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete author: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if any row was actually deleted
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Author not found", http.StatusNotFound)
			return
		}

		// Return the success response
		fmt.Fprintf(w, "Author deleted successfully")
	}
}

// DeleteBook deletes an existing book from the database
func DeleteBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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
		err = db.QueryRow(authorIDQuery, bookID).Scan(&authorID)
		if err != nil {
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
		err = db.QueryRow(otherBooksQuery, authorID, bookID).Scan(&numOtherBooks)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check for other books: %v", err), http.StatusInternalServerError)
			return
		}

		// Query to delete the book
		deleteBookQuery := `
            DELETE FROM books
            WHERE id = ?
        `

		// Execute the query to delete the book
		result, err := db.Exec(deleteBookQuery, bookID)
		if err != nil {
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
			_, err = db.Exec(deleteAuthorQuery, authorID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete author: %v", err), http.StatusInternalServerError)
				return
			}
		}

		fmt.Fprintf(w, "Book deleted successfully")
	}
}

// DeleteSubscriber deletes an existing subscriber from the database
func DeleteSubscriber(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
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
		result, err := db.Exec(deleteQuery, subscriberID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete subscriber: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if any row was actually deleted
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Subscriber not found", http.StatusNotFound)
			return
		}

		// Return the success response
		fmt.Fprintf(w, "Subscriber deleted successfully")
	}
}
