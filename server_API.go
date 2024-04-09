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
)

// Sample data structure to store dummy data
type Authors struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	
}


type AuthorBook struct {
	AuthorFirstname string `json:"author_firstname"`
	AuthorLastname  string `json:"author_lastname"`
	BookTitle       string `json:"book_title"`
}

type Books struct {
	Book_id		  	int			`json:"book_id"`
	Photo 	  		string		`json:"photo"`
	Title	  		string		`json:"title"`
	Author_id 		int	    	`json:"author_id"`
	Description		string		`json:"description"`
	Is_borrowed		bool		`json:"is_borrowed"`	
}


func initDB(username, password, hostname, port, dbname string) (*sql.DB, error) {
	var err error

	// Constructing the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)

	// Open a connection to the database
	var db *sql.DB
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}
	
	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
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

	db, _ := initDB(*dbUsername, *dbPassword, *dbHostname, *dbPort, *dbName)
	defer db.Close()

	log.Println("Starting our server.")

	http.HandleFunc("/", Home)
	http.HandleFunc("/info", Info)
	http.HandleFunc("/authors", GetAuthors(db))
	http.HandleFunc("/authorsbooks", GetAuthorsAndBooks(db))
	http.HandleFunc("/authors/", GetAuthorsAndBooksByID(db))
	http.HandleFunc("/books", GetBooksById(db))

	// http.HandleFunc("/books/add", AddItem)
	// http.HandleFunc("/books/update", UpdateItem)
	// http.HandleFunc("/books/delete", DeleteItem)

	log.Println("Started on port", *port)
	fmt.Println("To close connection CTRL+C :-)")

	// Spinning up the server.
	err := http.ListenAndServe(":"+*port, nil)
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

// GetAuthors handles requests to retrieve all items from the database
func GetAuthors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM authors")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer rows.Close()

		var authors []Authors
		for rows.Next() {
			var author Authors
			if err := rows.Scan(&author.ID,  &author.Firstname, &author.Lastname); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			authors = append(authors, author)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(authors)
	}
}





func GetAuthorsAndBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title
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
			var authorBook AuthorBook
			if err := rows.Scan(&authorBook.AuthorFirstname, &authorBook.AuthorLastname, &authorBook.BookTitle); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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

func GetAuthorsAndBooksByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorID := r.URL.Path[len("/authors/"):]
		id, err := strconv.Atoi(authorID)
		if err != nil {
			http.Error(w, "Invalid author ID", http.StatusBadRequest)
			return
		}

		query := `
			SELECT a.Firstname AS author_firstname, a.Lastname AS author_lastname, b.title AS book_title
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

		var authorFirstname, authorLastname, bookTitle string
		var books []string

		for rows.Next() {
			if err := rows.Scan(&authorFirstname, &authorLastname, &bookTitle); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			books = append(books, bookTitle)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authorAndBooks := struct {
			AuthorFirstname string   `json:"author_firstname"`
			AuthorLastname  string   `json:"author_lastname"`
			Books           []string `json:"books"`
		}{
			AuthorFirstname: authorFirstname,
			AuthorLastname:  authorLastname,
			Books:           books,
		}

		json.NewEncoder(w).Encode(authorAndBooks)
	}
}


// GetBooks handles requests to retrieve all items from the database
func GetBooksById(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bookID := r.URL.Query().Get("book_id")

		query := fmt.Sprintf("SELECT b.*, a.Firstname, a.Lastname FROM authors b JOIN authors a ON b.author_id = a.id WHERE b.id = %s", bookID)

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var books []Books
		for rows.Next() {
			var book Books
			if err := rows.Scan(&book.Book_id, &book.Photo, &book.Title, &book.Author_id, &book.Description, &book.Is_borrowed); err != nil {
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
