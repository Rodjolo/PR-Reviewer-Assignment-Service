package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
	"strconv"

	"github.com/gorilla/mux"
)

type Handlers struct {
	prService    *service.PRService
	userService  *service.UserService
	teamService  *service.TeamService
	statsService *service.StatsService
}

func NewHandlers(prService *service.PRService, userService *service.UserService, teamService *service.TeamService, statsService *service.StatsService) *Handlers {
	return &Handlers{
		prService:    prService,
		userService:  userService,
		teamService:  teamService,
		statsService: statsService,
	}
}

func (h *Handlers) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

// PR Handlers

type CreatePRRequest struct {
	Title    string `json:"title" example:"Add new feature"`
	AuthorID int    `json:"author_id" example:"1"`
}

// CreatePR godoc
// @Summary Создать Pull Request
// @Description Создает новый PR и автоматически назначает до 2 ревьюверов из команды автора
// @Tags PR
// @Accept json
// @Produce json
// @Param request body CreatePRRequest true "Данные PR"
// @Success 201 {object} models.PR
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /prs [post]
func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pr, err := h.prService.CreatePR(req.Title, req.AuthorID)
	if err != nil {
		if err.Error() == "author not found" || err.Error() == "author is not in any team" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, pr)
}

// GetPR godoc
// @Summary Получить PR по ID
// @Description Возвращает информацию о PR по его идентификатору
// @Tags PR
// @Produce json
// @Param id path int true "ID PR"
// @Success 200 {object} models.PR
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /prs/{id} [get]
func (h *Handlers) GetPR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid PR ID")
		return
	}

	pr, err := h.prService.GetPR(id)
	if err != nil {
		if err.Error() == "PR not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, pr)
}

// ListPRs godoc
// @Summary Получить список PR'ов
// @Description Возвращает список всех PR'ов или PR'ов конкретного пользователя
// @Tags PR
// @Produce json
// @Param user_id query int false "ID пользователя (фильтр)"
// @Success 200 {array} models.PR
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /prs [get]
func (h *Handlers) ListPRs(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		prs, err := h.prService.GetPRsByUserID(userID)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		h.respondJSON(w, http.StatusOK, prs)
		return
	}

	prs, err := h.prService.GetAllPRs()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, prs)
}

type ReassignRequest struct {
	OldReviewerID int `json:"old_reviewer_id" example:"2"`
}

// ReassignReviewer godoc
// @Summary Переназначить ревьювера
// @Description Заменяет одного ревьювера на случайного активного участника из команды заменяемого ревьювера
// @Tags PR
// @Accept json
// @Produce json
// @Param id path int true "ID PR"
// @Param request body ReassignRequest true "Данные для переназначения"
// @Success 200 {object} models.PR
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string "PR уже мержен"
// @Failure 500 {object} map[string]string
// @Router /prs/{id}/reassign [patch]
func (h *Handlers) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prID, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid PR ID")
		return
	}

	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pr, err := h.prService.ReassignReviewer(prID, req.OldReviewerID)
	if err != nil {
		if err.Error() == "PR not found" || err.Error() == "old reviewer is not assigned to this PR" || err.Error() == "reviewer is not in any team" || err.Error() == "no available reviewers in the team" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "cannot reassign reviewer: PR is already merged" {
			h.respondError(w, http.StatusConflict, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, pr)
}

// MergePR godoc
// @Summary Мержить PR
// @Description Изменяет статус PR на MERGED. После мержа изменения ревьюверов запрещены
// @Tags PR
// @Produce json
// @Param id path int true "ID PR"
// @Success 200 {object} models.PR
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string "PR уже мержен"
// @Failure 500 {object} map[string]string
// @Router /prs/{id}/merge [post]
func (h *Handlers) MergePR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid PR ID")
		return
	}

	pr, err := h.prService.MergePR(id)
	if err != nil {
		if err.Error() == "PR not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, pr)
}

// User Handlers

type CreateUserRequest struct {
	Name     string `json:"name" example:"Alice"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}

// CreateUser godoc
// @Summary Создать пользователя
// @Description Создает нового пользователя в системе
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user, err := h.userService.CreateUser(req.Name, isActive)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}

// GetUser godoc
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по его идентификатору
// @Tags Users
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// ListUsers godoc
// @Summary Получить список пользователей
// @Description Возвращает список всех пользователей в системе
// @Tags Users
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, users)
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" example:"Alice Updated"`
	IsActive *bool   `json:"is_active,omitempty" example:"false"`
}

// UpdateUser godoc
// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [patch]
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(id, req.Name, req.IsActive)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// Team Handlers

type CreateTeamRequest struct {
	Name string `json:"name" example:"backend"`
}

// CreateTeam godoc
// @Summary Создать команду
// @Description Создает новую команду в системе
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body CreateTeamRequest true "Данные команды"
// @Success 201 {object} models.Team
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string "Команда уже существует"
// @Failure 500 {object} map[string]string
// @Router /teams [post]
func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	team, err := h.teamService.CreateTeam(req.Name)
	if err != nil {
		if err.Error() == "team already exists" {
			h.respondError(w, http.StatusConflict, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, team)
}

// GetTeam godoc
// @Summary Получить команду по имени
// @Description Возвращает информацию о команде по её имени
// @Tags Teams
// @Produce json
// @Param name path string true "Имя команды"
// @Success 200 {object} models.Team
// @Failure 404 {object} map[string]string
// @Router /teams/{name} [get]
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	team, err := h.teamService.GetTeam(name)
	if err != nil {
		if err.Error() == "team not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, team)
}

// ListTeams godoc
// @Summary Получить список команд
// @Description Возвращает список всех команд в системе
// @Tags Teams
// @Produce json
// @Success 200 {array} models.Team
// @Failure 500 {object} map[string]string
// @Router /teams [get]
func (h *Handlers) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teamService.GetAllTeams()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, teams)
}

type AddMemberRequest struct {
	UserID int `json:"user_id" example:"1"`
}

// AddTeamMember godoc
// @Summary Добавить участника в команду
// @Description Добавляет пользователя в команду
// @Tags Teams
// @Accept json
// @Produce json
// @Param name path string true "Имя команды"
// @Param request body AddMemberRequest true "ID пользователя"
// @Success 200 {object} models.Team
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /teams/{name}/members [post]
func (h *Handlers) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["name"]

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	team, err := h.teamService.AddMember(teamName, req.UserID)
	if err != nil {
		if err.Error() == "team not found" || err.Error() == "user not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, team)
}

// RemoveTeamMember godoc
// @Summary Удалить участника из команды
// @Description Удаляет пользователя из команды
// @Tags Teams
// @Produce json
// @Param name path string true "Имя команды"
// @Param user_id query int true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /teams/{name}/members [delete]
func (h *Handlers) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["name"]

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := h.teamService.RemoveMember(teamName, userID); err != nil {
		if err.Error() == "team not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

// Stats Handler

// GetStats godoc
// @Summary Получить статистику
// @Description Возвращает статистику по пользователям, командам и PR'ам
// @Tags Stats
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 500 {object} map[string]string
// @Router /stats [get]
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, stats)
}

