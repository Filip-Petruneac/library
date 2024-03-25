package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    _ "github.com/go-sql-driver/mysql"
)

// Port we listen on.
const portNum string = ":8080"

// Sample data structure to store dummy data
type Item struct {
    ID   int `json:"id"`
    Lastname string `json:"lastname"`
	Firstname string `json:"firstname"`
	Books     int    `json:"books"`


}

var db *sql.DB

func initDB() {
    var err error
    // Database connection parameters
    username := "root"
    password := "password"
    hostname := "localhost"
    port := "4450" // Change this to the port where your MySQL is running
    dbname := "library"

    // Constructing the DSN (Data Source Name)
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)

    // Open a connection to the database
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Check if the connection is successful
    err = db.Ping()
    if err != nil {
        log.Fatal("Failed to ping database:", err)
    }
    log.Println("Connected to the MySQL database!")
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
func GetBooks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT * authors")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var books []Item
    for rows.Next() {
        var book Item
        if err := rows.Scan(&book.ID, &book.Lastname, &book.Firstname, &book.Books); err != nil {
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

// AddItem handles requests to add a new item to the database
func AddItem(w http.ResponseWriter, r *http.Request) {
    var newItem Item
    err := json.NewDecoder(r.Body).Decode(&newItem)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    _, err = db.Exec("INSERT INTO authors (Firstname, Lastname) VALUES (Vasile, Grigore)", newItem.Firstname, newItem.Lastname)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(newItem)
}

// UpdateItem handles requests to update an existing item in the database
func UpdateItem(w http.ResponseWriter, r *http.Request) {
    var updatedItem Item
    err := json.NewDecoder(r.Body).Decode(&updatedItem)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE items SET name = ? WHERE id = ?", updatedItem.Lastname, updatedItem.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(updatedItem)
}

// DeleteItem handles requests to delete an existing item from the database
func DeleteItem(w http.ResponseWriter, r *http.Request) {
    var deleteItem Item
    err := json.NewDecoder(r.Body).Decode(&deleteItem)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM items WHERE id = 1", deleteItem.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Item deleted successfully")
}

func main() {
    initDB()
    defer db.Close()

    log.Println("Starting our server.")

    http.HandleFunc("/", Home)
    http.HandleFunc("/info", Info)
    http.HandleFunc("/books", GetBooks)
    http.HandleFunc("/books/add", AddItem)
    http.HandleFunc("/books/update", UpdateItem)
    http.HandleFunc("/books/delete", DeleteItem)

    log.Println("Started on port", portNum)
    fmt.Println("To close connection CTRL+C :-)")

    // Spinning up the server.
    err := http.ListenAndServe(portNum, nil)
    if err != nil {
        log.Fatal(err)
    }
}
