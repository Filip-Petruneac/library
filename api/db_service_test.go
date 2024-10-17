package main

import (
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
)

type TestDBService struct {
	DB   *sql.DB
	Mock sqlmock.Sqlmock
}

func NewTestDBService() (*TestDBService, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, fmt.Errorf("an error occurred when opening a stub database connection: %w", err)
	}

	return &TestDBService{
		DB:   db,
		Mock: mock,
	}, nil
}
