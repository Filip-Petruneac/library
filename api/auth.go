package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string
	Password string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type Response struct {
	ExistingUserID int    `json:"existingUserID"`
	Message        string `json:"message"`
}

var (
	pattern       = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	domainPattern = regexp.MustCompile(`@(gmail\.com|yahoo\.com|outlook\.com)$`)
)

func signupUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		matches := pattern.MatchString(u.Email)
		domainMatches := domainPattern.MatchString(u.Email)
		if !matches || !domainMatches || u.Password == "" || u.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
			return
		}

		var existingUserID int
		query := `SELECT id FROM users WHERE email = ?`
		err := db.QueryRow(query, u.Email).Scan(&existingUserID)
		if err != nil && err != sql.ErrNoRows {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Database error"})
			return
		}

		if existingUserID != 0 {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Email already in use"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Error hashing password"})
			return
		}

		query = `INSERT INTO users (email, password) VALUES(?, ?)`
		_, err = db.Exec(query, u.Email, hashedPassword)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: fmt.Sprintf("Failed to insert user: %v", err)})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
	}
}

func loginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := User{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		matches := pattern.MatchString(u.Email)
		domainMatches := domainPattern.MatchString(u.Email)

		if !matches || !domainMatches || u.Password == "" || u.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
			return
		}

		var existingUserID int
		var hashedPassword []byte
		query := `SELECT id, password FROM users WHERE email = ?`
		err := db.QueryRow(query, u.Email).Scan(&existingUserID, &hashedPassword)

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "User doesn't exist!"})
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Database error"})
			return
		}

		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(u.Password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid email or password"})
			return
		}

		resp := Response{
			ExistingUserID: existingUserID,
			Message:        "User registered successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
