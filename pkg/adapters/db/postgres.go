package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

type PostgresDB struct {
    *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
    )

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    log.Println("Successfully connected to PostgreSQL database")
    return &PostgresDB{db}, nil
}

func (p *PostgresDB) Close() error {
    return p.DB.Close()
}