package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rodjolo/pr-reviewer-service/internal/database"
	"github.com/Rodjolo/pr-reviewer-service/internal/handlers"
	"github.com/Rodjolo/pr-reviewer-service/internal/repository"
	"github.com/Rodjolo/pr-reviewer-service/internal/router"
	"github.com/Rodjolo/pr-reviewer-service/internal/service"
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	testServer *httptest.Server
	testDB     *database.DB
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	dbURL      string
)

// TestMain настраивает и очищает тестовое окружение
func TestMain(m *testing.M) {
	var err error

	// Создаем pool для работы с Docker
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// Запускаем PostgreSQL контейнер
	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbURL = fmt.Sprintf("postgres://testuser:secret@%s/testdb?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", dbURL)

	// Устанавливаем таймаут для подключения к БД
	resource.Expire(120)

	// Ждем готовности БД
	var db *sql.DB
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Применяем миграции
	if err := applyMigrations(db); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	db.Close()

	// Инициализируем тестовый сервер
	testDB, err = database.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)
	prRepo := repository.NewPRRepository(testDB.DB)

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo, prRepo, teamRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatsService(prRepo)

	// Инициализируем handlers
	h := handlers.NewHandlers(prService, userService, teamService, statsService)

	// Настраиваем роутер
	r := router.NewRouter(h)

	// Создаем тестовый сервер
	testServer = httptest.NewServer(r)

	// Запускаем тесты
	code := m.Run()

	// Очистка
	testServer.Close()
	testDB.Close()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// applyMigrations применяет миграции к БД
func applyMigrations(db *sql.DB) error {
	// Находим директорию с миграциями
	migrationsPath := "../../migrations"
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Используем iofs для загрузки миграций (работает на Windows)
	sourceDriver, err := iofs.New(os.DirFS(absPath), ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// cleanupTestData очищает тестовые данные из БД
func cleanupTestData(t *testing.T) {
	queries := []string{
		"DELETE FROM pr_reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM team_members",
		"DELETE FROM teams",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := testDB.Exec(query); err != nil {
			t.Logf("Warning: failed to cleanup: %v", err)
		}
	}
}

// makeRequest выполняет HTTP запрос к тестовому серверу
func makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, testServer.URL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	return client.Do(req)
}

// TestCreateUser создает пользователя через API
func TestCreateUser(t *testing.T) {
	cleanupTestData(t)

	req := dto.CreateUserRequest{
		Name:     "TestUser",
		IsActive: boolPtr(true),
	}

	resp, err := makeRequest("POST", "/users", req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Name != "TestUser" {
		t.Errorf("Expected name 'TestUser', got '%s'", user.Name)
	}
	if !user.IsActive {
		t.Error("Expected user to be active")
	}
	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
}

// TestGetUser получает пользователя по ID
func TestGetUser(t *testing.T) {
	cleanupTestData(t)

	// Создаем пользователя
	createReq := dto.CreateUserRequest{Name: "TestUser", IsActive: boolPtr(true)}
	resp, _ := makeRequest("POST", "/users", createReq)

	var createdUser models.User
	json.NewDecoder(resp.Body).Decode(&createdUser)
	resp.Body.Close()

	// Получаем пользователя
	resp, err := makeRequest("GET", fmt.Sprintf("/users/%d", createdUser.ID), nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, user.ID)
	}
}

// TestCreateTeam создает команду
func TestCreateTeam(t *testing.T) {
	cleanupTestData(t)

	req := dto.CreateTeamRequest{Name: "backend"}

	resp, err := makeRequest("POST", "/teams", req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var team models.Team
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if team.Name != "backend" {
		t.Errorf("Expected team name 'backend', got '%s'", team.Name)
	}
}

// TestAddTeamMember добавляет участника в команду
func TestAddTeamMember(t *testing.T) {
	cleanupTestData(t)

	// Создаем пользователя
	userReq := dto.CreateUserRequest{Name: "Alice", IsActive: boolPtr(true)}
	resp, _ := makeRequest("POST", "/users", userReq)
	var user models.User
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()

	// Создаем команду
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ = makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	// Добавляем участника
	memberReq := dto.AddMemberRequest{UserID: user.ID}
	resp, err := makeRequest("POST", "/teams/backend/members", memberReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var team models.Team
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(team.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(team.Members))
	}
	if team.Members[0].ID != user.ID {
		t.Errorf("Expected member ID %d, got %d", user.ID, team.Members[0].ID)
	}
}

// TestCreatePR создает PR с автоматическим назначением ревьюверов
func TestCreatePR(t *testing.T) {
	cleanupTestData(t)

	// Создаем команду с несколькими участниками
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ := makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	// Создаем пользователей
	users := []string{"Alice", "Bob", "Charlie"}
	var userIDs []int
	for _, name := range users {
		userReq := dto.CreateUserRequest{Name: name, IsActive: boolPtr(true)}
		resp, _ := makeRequest("POST", "/users", userReq)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()
		userIDs = append(userIDs, user.ID)

		// Добавляем в команду
		memberReq := dto.AddMemberRequest{UserID: user.ID}
		resp, _ = makeRequest("POST", "/teams/backend/members", memberReq)
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Создаем PR от первого пользователя
	prReq := dto.CreatePRRequest{
		Title:    "Test PR",
		AuthorID: userIDs[0],
	}

	resp, err := makeRequest("POST", "/prs", prReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var pr models.PR
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if pr.Title != "Test PR" {
		t.Errorf("Expected title 'Test PR', got '%s'", pr.Title)
	}
	if pr.AuthorID != userIDs[0] {
		t.Errorf("Expected author ID %d, got %d", userIDs[0], pr.AuthorID)
	}
	if len(pr.Reviewers) == 0 {
		t.Error("Expected at least one reviewer")
	}
	if len(pr.Reviewers) > 2 {
		t.Errorf("Expected at most 2 reviewers, got %d", len(pr.Reviewers))
	}
	// Проверяем, что автор не назначен ревьювером
	for _, reviewerID := range pr.Reviewers {
		if reviewerID == userIDs[0] {
			t.Error("Author should not be assigned as reviewer")
		}
	}
}

// TestReassignReviewer переназначает ревьювера
func TestReassignReviewer(t *testing.T) {
	cleanupTestData(t)

	// Настраиваем тестовые данные
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ := makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	users := []string{"Alice", "Bob", "Charlie", "David"}
	var userIDs []int
	for _, name := range users {
		userReq := dto.CreateUserRequest{Name: name, IsActive: boolPtr(true)}
		resp, _ := makeRequest("POST", "/users", userReq)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()
		userIDs = append(userIDs, user.ID)

		memberReq := dto.AddMemberRequest{UserID: user.ID}
		resp, _ = makeRequest("POST", "/teams/backend/members", memberReq)
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Создаем PR
	prReq := dto.CreatePRRequest{Title: "Test PR", AuthorID: userIDs[0]}
	prResp, _ := makeRequest("POST", "/prs", prReq)
	var pr models.PR
	json.NewDecoder(prResp.Body).Decode(&pr)
	prResp.Body.Close()

	if len(pr.Reviewers) == 0 {
		t.Fatal("PR should have at least one reviewer")
	}

	oldReviewerID := pr.Reviewers[0]

	// Переназначаем ревьювера
	reassignReq := dto.ReassignRequest{OldReviewerID: oldReviewerID}
	resp, err := makeRequest("PATCH", fmt.Sprintf("/prs/%d/reassign", pr.ID), reassignReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var updatedPR models.PR
	if err := json.NewDecoder(resp.Body).Decode(&updatedPR); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Проверяем, что старый ревьювер заменен
	found := false
	for _, reviewerID := range updatedPR.Reviewers {
		if reviewerID == oldReviewerID {
			found = true
			break
		}
	}
	if found {
		t.Error("Old reviewer should not be in the reviewers list")
	}
}

// TestMergePR мержит PR
func TestMergePR(t *testing.T) {
	cleanupTestData(t)

	// Настраиваем тестовые данные
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ := makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	// Создаем несколько пользователей для команды
	users := []string{"Alice", "Bob", "Charlie"}
	var userIDs []int
	for _, name := range users {
		userReq := dto.CreateUserRequest{Name: name, IsActive: boolPtr(true)}
		resp, _ := makeRequest("POST", "/users", userReq)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()
		userIDs = append(userIDs, user.ID)

		memberReq := dto.AddMemberRequest{UserID: user.ID}
		resp, _ = makeRequest("POST", "/teams/backend/members", memberReq)
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Создаем PR от первого пользователя
	prReq := dto.CreatePRRequest{Title: "Test PR", AuthorID: userIDs[0]}
	prResp, _ := makeRequest("POST", "/prs", prReq)
	var pr models.PR
	json.NewDecoder(prResp.Body).Decode(&pr)
	prResp.Body.Close()

	// Мержим PR
	resp, err := makeRequest("POST", fmt.Sprintf("/prs/%d/merge", pr.ID), nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var mergedPR models.PR
	if err := json.NewDecoder(resp.Body).Decode(&mergedPR); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if mergedPR.Status != models.PRStatusMerged {
		t.Errorf("Expected status MERGED, got '%s'", mergedPR.Status)
	}

	// Проверяем, что после мержа нельзя переназначить ревьювера (если есть ревьюверы)
	if len(mergedPR.Reviewers) > 0 {
		reassignReq := dto.ReassignRequest{OldReviewerID: mergedPR.Reviewers[0]}
		resp, _ = makeRequest("PATCH", fmt.Sprintf("/prs/%d/reassign", mergedPR.ID), reassignReq)
		resp.Body.Close()
		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 after merge, got %d", resp.StatusCode)
		}
	}
}

// TestBulkDeactivateTeam тестирует массовую деактивацию команды
func TestBulkDeactivateTeam(t *testing.T) {
	cleanupTestData(t)

	// Создаем команду
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ := makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	// Создаем пользователей
	users := []string{"Alice", "Bob", "Charlie"}
	var userIDs []int
	for _, name := range users {
		userReq := dto.CreateUserRequest{Name: name, IsActive: boolPtr(true)}
		resp, _ := makeRequest("POST", "/users", userReq)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()
		userIDs = append(userIDs, user.ID)

		memberReq := dto.AddMemberRequest{UserID: user.ID}
		resp, _ = makeRequest("POST", "/teams/backend/members", memberReq)
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Создаем PR с ревьюверами
	prReq := dto.CreatePRRequest{Title: "Test PR", AuthorID: userIDs[0]}
	var prResp *http.Response
	prResp, _ = makeRequest("POST", "/prs", prReq)
	var pr models.PR
	json.NewDecoder(prResp.Body).Decode(&pr)
	prResp.Body.Close()

	// Деактивируем команду
	resp, err := makeRequest("POST", "/teams/backend/deactivate", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var result dto.BulkDeactivateTeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.DeactivatedUsers != 3 {
		t.Errorf("Expected 3 deactivated users, got %d", result.DeactivatedUsers)
	}

	// Проверяем, что пользователи деактивированы
	for _, userID := range userIDs {
		resp, _ := makeRequest("GET", fmt.Sprintf("/users/%d", userID), nil)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)
		resp.Body.Close()

		if user.IsActive {
			t.Errorf("Expected user %d to be inactive", userID)
		}
	}
}

// TestGetStats получает статистику
func TestGetStats(t *testing.T) {
	cleanupTestData(t)

	// Создаем тестовые данные
	teamReq := dto.CreateTeamRequest{Name: "backend"}
	resp, _ := makeRequest("POST", "/teams", teamReq)
	if resp != nil {
		resp.Body.Close()
	}

	userReq := dto.CreateUserRequest{Name: "Alice", IsActive: boolPtr(true)}
	resp, _ = makeRequest("POST", "/users", userReq)
	var user models.User
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()

	memberReq := dto.AddMemberRequest{UserID: user.ID}
	resp, _ = makeRequest("POST", "/teams/backend/members", memberReq)
	if resp != nil {
		resp.Body.Close()
	}

	prReq := dto.CreatePRRequest{Title: "Test PR", AuthorID: user.ID}
	resp, _ = makeRequest("POST", "/prs", prReq)
	if resp != nil {
		resp.Body.Close()
	}

	// Получаем статистику
	resp, err := makeRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var stats map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats["total_users"] < 1 {
		t.Errorf("Expected at least 1 user, got %d", stats["total_users"])
	}
	if stats["total_teams"] < 1 {
		t.Errorf("Expected at least 1 team, got %d", stats["total_teams"])
	}
	if stats["total_prs"] < 1 {
		t.Errorf("Expected at least 1 PR, got %d", stats["total_prs"])
	}
}

// boolPtr возвращает указатель на bool
func boolPtr(b bool) *bool {
	return &b
}
