
---

![received_1097514964147756.(jpg?1652351843](https://culturenl.co.uk/wp-content/uploads/2022/04/Cumbernauld-Library-Banner.jpg)

# Library Management System API

This is a Go-based REST API for managing a library, with features including CRUD operations for books, authors, and subscribers, book borrowing/returning functionality, and user authentication.

## Features

- **CRUD operations**: Manage books, authors, and subscribers.
- **Book borrowing and returning**: Keep track of borrowed books.
- **Search functionality**: Search for books and authors.
- **User authentication**: Signup and login for users.
- **Image uploads**: Support for uploading author photos.

## Prerequisites

Ensure the following are installed:

- **Go** (version 1.15 or later)
- **MySQL database**
- **Git** (optional, for cloning the repository)
- **Docker and Docker Compose** (for containerized deployment)

## Environment Setup

### Environment Variables

Create a `.env` file in the project root with these variables:

#### Database configuration

- `DB_HOSTNAME`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password

#### MySQL-specific variables

- `MYSQL_ROOT_PASSWORD`: MySQL root password
- `MYSQL_DATABASE`: Name of the MySQL database
- `MYSQL_USER`: MySQL user
- `MYSQL_PASSWORD`: MySQL user password

#### API configuration

- `API_URL`: Base URL for the API

These environment variables should replace any hardcoded values within the code, ensuring flexible configuration across development and production environments.

## Installation & Setup

### Dependency Management

This project uses both standard library and external dependencies. Key external libraries include:

- **MySQL driver**: `github.com/go-sql-driver/mysql`
- **Router**: `github.com/gorilla/mux`
- **Environment loader**: `github.com/joho/godotenv`

#### Installing Dependencies

To install external dependencies, run the following command:

```bash
go get github.com/go-sql-driver/mysql
go get github.com/gorilla/mux
go get github.com/joho/godotenv
```

Iniziate Go modules:

```bash
go mod init <module-name>
go mod tidy
```

This will create the necessary `go.mod` and `go.sum` files to manage your dependencies.

## API Endpoints

The API exposes the following endpoints:

### Books

- `GET /books`: Retrieve all books.
- `GET /books/{id}`: Retrieve a book by ID.
- `POST /books/new`: Add a new book.
- `PUT /books/{id}`: Update a book by ID.
- `DELETE /books/{id}`: Delete a book by ID.

### Authors

- `GET /authors`: Retrieve all authors.
- `GET /authors/{id}`: Retrieve books by a specific author.
- `POST /authors/new`: Add a new author.
- `PUT /authors/{id}`: Update an author by ID.
- `DELETE /authors/{id}`: Delete an author by ID.
- `POST /author/photo/{id}`: Upload a photo for an author.

### Subscribers

- `GET /subscribers`: Retrieve all subscribers.
- `GET /subscribers/{id}`: Retrieve a subscriber by book ID.
- `POST /subscribers/new`: Add a new subscriber.
- `PUT /subscribers/{id}`: Update a subscriber by ID.
- `DELETE /subscribers/{id}`: Delete a subscriber by ID.

### Book Borrowing & Returning

- `POST /book/borrow`: Borrow a book.
- `POST /book/return`: Return a borrowed book.

### Search

- `GET /search_books`: Search for books.
- `GET /search_authors`: Search for authors.

### User Authentication

- `POST /signup`: User signup.
- `POST /login`: User login.

## Running the Application with Docker Compose

 Command/Instruction | Description |
---------------------|-------------|
 `git clone https://github.com/yourusername/library-management-system.git`<br>`cd library-management-system` | Clone the repository and  into the project directory. |
 `docker-compose up-all --build` | Build and start the Docker containers. |
 `docker-compose down` | Stop the running containers. |
 `docker-compose up-all --build` | Rebuild and start the containers after making changes. |
 Open in browser: `http://localhost:8081` | Access the API in your browser. |
 `docker-compose logs` | View logs of the running containers. |
 `docker-compose down -v` | Stop and remove containers, volumes, and networks. |

**Note**:
Ensure ou have in root project `.env` file for connecting to the database and other configurations
---
