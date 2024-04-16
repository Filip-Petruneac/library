package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

)

// Sample data structure to store dummy data
type Author struct {
	ID           int    `json:"id"`
	Lastname     string `json:"lastname"`
	Firstname    string `json:"firstname"`
	Photo        string `json:"photo"`
}


type AuthorBook struct {
	AuthorFirstname string `json:"author_firstname"`
    AuthorLastname  string `json:"author_lastname"`
    BookTitle string `json:"book_title"`
    BookPhoto string `json:"book_photo"`

}

type BookAuthorInfo struct {
    BookTitle       string `json:"book_title"`
    AuthorID        int    `json:"author_id"`
    BookPhoto       string `json:"book_photo"`
    IsBorrowed      bool   `json:"is_borrowed"`
    BookID          int    `json:"book_id"`
    BookDetails     string `json:"book_details"`
    AuthorLastname  string `json:"author_lastname"`
    AuthorFirstname string `json:"author_firstname"`
}

type Subscriber struct {
	ID        int    `json:"id"`
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
	Email     string `json:"email"`
}


func initDB(username, password, hostname, port, dbname string) (*sql.DB, error) {
	var err error

	// Constructing the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)

	// Open a connection to the database
	var db *sql.DB
	db, err = sql.Open("mysql", dsn)
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

func main() {
	port := flag.String("port", "8080", "Server Port")
	dbUsername := flag.String("db-user", "root", "Database Username")
	dbPassword := flag.String("db-password", "password", "Database Password")
	dbHostname := flag.String("db-hostname", "localhost", "Database hostname")
	dbPort := flag.String("db-port", "4450", "Database port")
	dbName := flag.String("db-name", "library", "Database name")

	db, err := initDB(*dbUsername, *dbPassword, *dbHostname, *dbPort, *dbName)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	log.Println("Starting our server.")

	router := mux.NewRouter()

	router.HandleFunc("/", Home)
	router.HandleFunc("/info", Info)
	router.HandleFunc("/authors", GetAuthors(db)).Methods("GET")
	router.HandleFunc("/authorsbooks", GetAuthorsAndBooks(db)).Methods("GET")
	router.HandleFunc("/authors/{id}", GetAuthorBooksByID(db)).Methods("GET")
	router.HandleFunc("/books", GetBookByID(db)).Methods("GET")
	router.HandleFunc("/book/borrow", BorrowBook(db)).Methods("POST")
	router.HandleFunc("/book/return", ReturnBorrowedBook(db)).Methods("POST")
	router.HandleFunc("/subscribers_by_book", GetSubscribersByBookId(db)).Methods("GET")
	router.HandleFunc("/authors/new", AddAuthor(db)).Methods("POST")
	router.HandleFunc("/books/new", AddBook(db)).Methods("POST")

	http.Handle("/", router)


	log.Println("Started on port", *port)
	fmt.Println("To close connection CTRL+C :-)")

	// Spinning up the server.
	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal(err)
	}
}


// Handler functions...

// Home handles requests to the homepage
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}

// Info handles requests to the info page
func Info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Info page")
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
            AuthorFirstname string        `json:"author_firstname"`
            AuthorLastname  string        `json:"author_lastname"`
            AuthorPhoto     string        `json:"author_photo"`
            Books           []AuthorBook `json:"books"`
        }{
            AuthorFirstname: authorFirstname,
            AuthorLastname:  authorLastname,
            AuthorPhoto:     authorPhoto,
            Books:           books,
        }

        json.NewEncoder(w).Encode(authorAndBooks)
    }
}


// GetBookById retrieves information about a specific book based on its ID
func GetBookByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookID := r.URL.Query().Get("book_id")
		intBookID, err := strconv.Atoi(bookID)
        if err != nil {
            http.Error(w, "Invalid book ID", http.StatusBadRequest)
            return
        }
		query :=`
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

// GetSubscribersByBookId retrieves subscribers who have borrowed a specific book
func GetSubscribersByBookId(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	
		bookID := r.URL.Query().Get("book_id")
		if bookID == "" {
			http.Error(w, "Missing book ID parameter", http.StatusBadRequest)
			return
		}

		query := `
			SELECT s.id, s.Lastname, s.Firstname, s.Email
			FROM subscribers s
			JOIN borrowed_books bb ON s.id = bb.subscriber_id
			WHERE bb.book_id = ?
		`

		rows, err := db.Query(query, bookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var subscribers []Subscriber
		
		// Iterate over the query result set and populate the subscribers slice
		for rows.Next() {
			var subscriber Subscriber
			if err := rows.Scan(&subscriber.ID, &subscriber.Lastname, &subscriber.Firstname, &subscriber.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			subscribers = append(subscribers, subscriber)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(subscribers)
	}
}

// AddAuthor add a new author to the database
func AddAuthor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verificăm metoda HTTP
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

		// Query to add author
		query := `
			INSERT INTO authors (lastname, firstname, photo) 
			VALUES (?, ?, ?)
		`

		// We run the query
		result, err := db.Exec(query, author.Lastname, author.Firstname, author.Photo)
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

		// We return the response with the author ID inserted
		response := map[string]int{"id": int(id)}
		json.NewEncoder(w).Encode(response)
	}
}

// AddBook adds a new book to the database
func AddBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the HTTP method
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		// Parse the JSON data received from the request
		var book struct {
			Title       string `json:"title"`
			AuthorID    int    `json:"author_id"`
			Photo       string `json:"photo"`
			Details     string `json:"details"`
		}
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

		// Query to add the book
		query := `
			INSERT INTO books (title, author_id, photo, details) 
			VALUES (?, ?, ?, ?)
		`

		// Execute the query
		result, err := db.Exec(query, book.Title, book.AuthorID, book.Photo, book.Details)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert book: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the ID of the inserted book
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get last insert ID", http.StatusInternalServerError)
			return
		}

		// Return the response with the ID of the inserted book
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

		// Check if the book is already borrowed
		var isBorrowed bool
		err = db.QueryRow("SELECT is_borrowed FROM books WHERE id = ?", requestBody.BookID).Scan(&isBorrowed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if isBorrowed {
			http.Error(w, "Book is already borrowed", http.StatusConflict)
			return
		}

		// Insert a new record in the borrowed_books table
		_, err = db.Exec("INSERT INTO borrowed_books (subscriber_id, book_id, date_of_borrow) VALUES (?, ?, NOW())", requestBody.SubscriberID, requestBody.BookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the is_borrowed status of the book
		_, err = db.Exec("UPDATE books SET is_borrowed = TRUE WHERE id = ?", requestBody.BookID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Book borrowed successfully")
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

