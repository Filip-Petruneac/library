package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Sample data structure to store dummy data
type Author struct {
	ID        int    `json:"id"`
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
	Books     int    `json:"books"`
}

// type Book struct {
//     ID            int    `json:"id"`
//     Photo         string `json:"photo"`
//     Title         string `json:"title"`
//     Author        int    `json:"author"`
//     Description   string `json:"description"`
//     Subscriber    int    `json:"subscriber"`
//     BorrowedBooks int    `json:"borrowedbooks"`
//     IsBorrowed    bool   `json:"isborrowed"`
// }


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

// GetBooks handles requests to retrieve all items from the database
func GetAuthors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM authors")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer rows.Close()

		var books []Author
		for rows.Next() {
			var author Author
			if err := rows.Scan(&author.ID, &author.Lastname, &author.Firstname, &author.Books); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			books = append(books, author)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(books)
	}
}

// AddItem handles requests to add a new item to the database
// func AddItem(db http.ResponseWriter, r *http.Request) {
// 	var newItem Book
// 	err := json.NewDecoder(r.Body).Decode(&newItem)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// _, err = db.Exec("INSERT INTO authors (Firstname, Lastname) VALUES (Vasile, Grigore)", newItem.Firstname, newItem.Lastname)
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }
// json.NewEncoder(w).Encode(newItem)
// }

// UpdateItem handles requests to update an existing item in the database
// func UpdateItem(w http.ResponseWriter, r *http.Request) {
// 	var updatedItem Item
// 	err := json.NewDecoder(r.Body).Decode(&updatedItem)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// _, err = db.Exec("UPDATE items SET name = ? WHERE id = ?", updatedItem.Lastname, updatedItem.ID)
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }

// json.NewEncoder(w).Encode(updatedItem)
// }

// DeleteItem handles requests to delete an existing item from the database
// func DeleteItem(w http.ResponseWriter, r *http.Request) {
// 	var deleteItem Item
// 	err := json.NewDecoder(r.Body).Decode(&deleteItem)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// _, err = db.Exec("DELETE FROM items WHERE id = 1", deleteItem.ID)
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }

// 	fmt.Fprintf(w, "Item deleted successfully")
// }
