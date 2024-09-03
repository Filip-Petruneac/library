package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Regular expressions for email validation
var (
	pattern       = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	domainPattern = regexp.MustCompile(`@(gmail\.com|yahoo\.com|outlook\.com)$`)
)

// User represents a user entity in the application
type User struct {
	Email    string
	Password string
}

// ErrorResponse represents an error response format
type ErrorResponse struct {
	Message string `json:"message"`
}

// Response represents a successful response format
type Response struct {
	ExistingUserID int    `json:"existingUserID"`
	Message        string `json:"message"`
}

// SignupUser handles user registration
func (app *App) SignupUser(w http.ResponseWriter, r *http.Request) {
    var u User

    // Decode JSON body to User struct
    err := json.NewDecoder(r.Body).Decode(&u)
    if err != nil {
        app.Logger.Printf("Error decoding JSON: %v", err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid JSON format"})
        return
    }

    // Log the input values for debugging
    app.Logger.Printf("Received signup request: Email=%s, Password=%s", u.Email, u.Password)

    // Validate the email and password
    matches := pattern.MatchString(u.Email)
    domainMatches := domainPattern.MatchString(u.Email)
    if !matches || !domainMatches || u.Password == "" || u.Email == "" {
        app.Logger.Println("Validation failed: Invalid email or password")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
        return
    }

    // Check if the user already exists
    var existingUserID int
    query := `SELECT id FROM users WHERE email = ?`
    err = app.DB.QueryRow(query, u.Email).Scan(&existingUserID)  // Use `=` instead of `:=`
    if err != nil && err != sql.ErrNoRows {
        app.Logger.Printf("Database error: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Database error"})
        return
    }

    if existingUserID != 0 {
        w.WriteHeader(http.StatusConflict)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Email already in use"})
        return
    }

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
    if err != nil {
        app.Logger.Printf("Error hashing password: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Error hashing password"})
        return
    }

    // Insert the new user
    query = `INSERT INTO users (email, password) VALUES(?, ?)`
    _, err = app.DB.Exec(query, u.Email, hashedPassword)
    if err != nil {
        app.Logger.Printf("Failed to insert user: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: fmt.Sprintf("Failed to insert user: %v", err)})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginUser handles user authentication
func (app *App) LoginUser(w http.ResponseWriter, r *http.Request) {
    var u User

    // Decode JSON body to User struct
    err := json.NewDecoder(r.Body).Decode(&u)
    if err != nil {
        app.Logger.Printf("Error decoding JSON: %v", err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid JSON format"})
        return
    }

    // Log the input values for debugging
    app.Logger.Printf("Received login request: Email=%s, Password=%s", u.Email, u.Password)

    // Validate the email and password
    matches := pattern.MatchString(u.Email)
    domainMatches := domainPattern.MatchString(u.Email)
    if !matches || !domainMatches || u.Password == "" || u.Email == "" {
        app.Logger.Println("Validation failed: Invalid email or password")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
        return
    }

    // Check if the user exists
    var existingUserID int
    var hashedPassword []byte
    query := `SELECT id, password FROM users WHERE email = ?`
    err = app.DB.QueryRow(query, u.Email).Scan(&existingUserID, &hashedPassword)
    if err == sql.ErrNoRows {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "User doesn't exist!"})
        return
    }

    if err != nil {
        app.Logger.Printf("Database error: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Database error"})
        return
    }

    // Compare the hashed password
    err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(u.Password))
    if err == bcrypt.ErrMismatchedHashAndPassword {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
        return
    }

    resp := Response{
        ExistingUserID: existingUserID,
        Message:        "User logged in successfully",
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(resp)
}
