# Library Management System API

This is a Go-based REST API for a Library Management System. It provides endpoints for managing books, authors, and subscribers in a library.

## Features

- CRUD operations for books, authors, and subscribers
- Book borrowing and returning functionality
- Search functionality for books and authors
- User authentication (signup and login)
- Image upload for author photos


## Prerequisites

- Go (version 1.15 or later)
- MySQL database
- Git (optional, for cloning the repository)

## Dependencies

This project uses the following external packages:

- github.com/go-sql-driver/mysql
- github.com/gorilla/mux
- golang.org/x/crypto/bcrypt


## API Endpoints

- `GET /books`: Get all books
- `GET /authors`: Get all authors
- `GET /authorsbooks`: Get all authors and their books
- `GET /authors/{id}`: Get books by author ID
- `GET /books/{id}`: Get book by ID
- `GET /subscribers/{id}`: Get subscribers by book ID
- `GET /subscribers`: Get all subscribers
- `POST /book/borrow`: Borrow a book
- `POST /book/return`: Return a borrowed book
- `POST /authors/new`: Add a new author
- `POST /author/photo/{id}`: Add author photo
- `POST /books/new`: Add a new book
- `POST /subscribers/new`: Add a new subscriber
- `PUT /authors/{id}`: Update an author
- `PUT /books/{id}`: Update a book
- `PUT /subscribers/{id}`: Update a subscriber
- `DELETE /authors/{id}`: Delete an author
- `DELETE /books/{id}`: Delete a book
- `DELETE /subscribers/{id}`: Delete a subscriber
- `GET /search_books`: Search books
- `GET /search_authors`: Search authors
- `POST /signup`: User signup
- `POST /login`: User login
