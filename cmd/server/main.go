// Package main is the entry point for the PR reviewer service HTTP server.
package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Rodjolo/pr-reviewer-service/internal/database"
	"github.com/Rodjolo/pr-reviewer-service/internal/handlers"
	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/internal/router"
	"github.com/Rodjolo/pr-reviewer-service/internal/service"

	_ "github.com/Rodjolo/pr-reviewer-service/docs" // Swagger docs
)

func main() {
	// Получаем строку подключения к БД
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Подключаемся к БД
	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db.DB)
	teamRepo := repository.NewTeamRepository(db.DB)
	prRepo := repository.NewPRRepository(db.DB)

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo, prRepo, teamRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatsService(prRepo)

	// Инициализируем handlers
	h := handlers.NewHandlers(prService, userService, teamService, statsService)

	// Настраиваем роутер
	r := router.NewRouter(h)

	// Получаем порт из переменной окружения или используем 8080 по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Получаем внешний порт из переменной окружения (для Docker)
	externalPort := os.Getenv("EXTERNAL_PORT")
	if externalPort == "" {
		externalPort = port
	}

	log.Printf("Server starting on port %s (external: %s)", port, externalPort)

	// Используем net.Listen для явного указания IPv4
	listener, err := net.Listen("tcp4", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Server listening on %s (accessible externally on port %s)", listener.Addr().String(), externalPort)
	if err := http.Serve(listener, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
