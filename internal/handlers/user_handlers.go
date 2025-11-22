package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
)

// CreateUser godoc
// @Summary Создать пользователя
// @Description Создает нового пользователя в системе
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "Данные пользователя"
// @Success 201 {object} models.User
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
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

// UpdateUser godoc
// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body dto.UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} models.User
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [patch]
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
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

// BulkDeactivateTeam godoc
// @Summary Массовая деактивация пользователей команды
// @Description Деактивирует всех пользователей команды и безопасно переназначает ревьюверов в открытых PR
// @Tags Users
// @Produce json
// @Param name path string true "Имя команды"
// @Success 200 {object} dto.BulkDeactivateTeamResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{name}/deactivate [post]
func (h *Handlers) BulkDeactivateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["name"]

	result, err := h.userService.BulkDeactivateTeam(teamName)
	if err != nil {
		if err.Error() == "team not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

