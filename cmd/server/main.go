package main

import (
	"log"
	"net/http"
	"os"
	"pr-reviewer-service/internal/database"
	"pr-reviewer-service/internal/handlers"
	"pr-reviewer-service/internal/repository"
	"pr-reviewer-service/internal/router"
	"pr-reviewer-service/internal/service"

	_ "pr-reviewer-service/docs" // Swagger docs
)

func main() {
	// Получаем строку подключения к БД
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable"
	}

	// Подключаемся к БД
	db, err := database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db.DB)
	teamRepo := repository.NewTeamRepository(db.DB)
	prRepo := repository.NewPRRepository(db.DB)

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo)
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

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

