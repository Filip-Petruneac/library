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
- Docker and Docker Compose (for containerized deployment)

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

## Docker Compose Commands

To run the application using Docker Compose, follow these steps:

1. Make sure you have Docker and Docker Compose installed on your system.

2. Clone the repository (if you haven't already):
git clone https://github.com/yourusername/library-management-system.git
cd library-management-system

3. Build and start the containers:
docker-compose up --build

4. To stop the containers, use:
docker-compose down

5. To rebuild the application after making changes:
docker-compose up --build

6. To view logs of the running containers:
docker-compose logs

7. To access the application, open a web browser and go to:
http://localhost:8081

8. To stop the containers and remove the volumes:
docker-compose down -v

Note: Make sure to update your Go application to use the environment variables for database connection (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME) instead of hardcoded values.
