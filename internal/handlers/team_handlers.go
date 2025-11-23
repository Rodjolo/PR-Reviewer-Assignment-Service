package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/gorilla/mux"
)

// CreateTeam godoc
// @Summary Создать команду
// @Description Создает новую команду в системе
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body dto.CreateTeamRequest true "Данные команды"
// @Success 201 {object} models.Team
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Команда уже существует"
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams [post]
func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeamRequest
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
func (h *Handlers) ListTeams(w http.ResponseWriter, _ *http.Request) {
	teams, err := h.teamService.GetAllTeams()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, teams)
}

// AddTeamMember godoc
// @Summary Добавить участника в команду
// @Description Добавляет пользователя в команду
// @Tags Teams
// @Accept json
// @Produce json
// @Param name path string true "Имя команды"
// @Param request body dto.AddMemberRequest true "ID пользователя"
// @Success 200 {object} models.Team
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /teams/{name}/members [post]
func (h *Handlers) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamName := vars["name"]

	var req dto.AddMemberRequest
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

	h.respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "member removed"})
}
