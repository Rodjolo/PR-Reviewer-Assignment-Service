package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула соединений для высокой нагрузки
	db.SetMaxOpenConns(25)                 // Максимум открытых соединений
	db.SetMaxIdleConns(5)                  // Максимум неактивных соединений
	db.SetConnMaxLifetime(5 * 60 * 1000000000) // 5 минут
	db.SetConnMaxIdleTime(10 * 60 * 1000000000) // 10 минут

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

