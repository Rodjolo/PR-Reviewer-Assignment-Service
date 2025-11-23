// Package main provides database migration utilities.
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		up   = flag.Bool("up", false, "Run migrations up")
		down = flag.Bool("down", false, "Run migrations down")
	)
	flag.Parse()

	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if *up {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	} else if *down {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		fmt.Println("Migrations rolled back successfully")
	} else {
		flag.Usage()
		os.Exit(1)
	}
}
