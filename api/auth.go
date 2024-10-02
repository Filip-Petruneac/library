package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Regular expressions for email validation
var (
	pattern       = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	domainPattern = regexp.MustCompile(`@(gmail\.com|yahoo\.com|outlook\.com)$`)
)

var sessionStore = SessionStore{store: make(map[string]Session)}

// Session holds session token and associated user ID
type Session struct {
	UserID    int
	ExpiresAt time.Time
}

// SessionStore to hold active sessions
type SessionStore struct {
	store map[string]Session
	mu    sync.Mutex
}

// Add a session to the session store
func (s *SessionStore) Add(token string, session Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[token] = session
}

// Get a session from the session store
func (s *SessionStore) Get(token string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.store[token]
	return session, exists
}

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
	err = app.DB.QueryRow(query, u.Email).Scan(&existingUserID) // Use `=` instead of `:=`
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

	// Generate a session token
	token, err := generateSessionToken()
	if err != nil {
		app.Logger.Printf("Error generating session token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error generating session token"})
		return
	}

	sessionExpiration := time.Now().Add(1440 * time.Minute)

	// Add the session to the session store
	sessionStore.Add(token, Session{
		UserID:    existingUserID,
		ExpiresAt: sessionExpiration,
	})

	// Respond to the client with the session token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":          token,
		"existingUserID": existingUserID,
		"message":        "User logged in successfully",
	})
}

func (app *App) VerifySessionToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("works")
		// Get the session token from the cookie
		cookie, err := r.Cookie("token")
		fmt.Println("Cookie name", cookie)
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Message: "Unauthorized access"})
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Bad request"})
			return
		}

		// Retrieve the session token from the cookie
		sessionToken := cookie.Value

		// Get the session from the store
		session, exists := sessionStore.Get(sessionToken)
		if !exists {
			// If the session token is not valid, return unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid session token"})
			return
		}

		// Check if the session has expired
		if session.ExpiresAt.Before(time.Now()) {
			// If the session is expired, return unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Session expired"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32) // Generate a random 32-byte token
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
