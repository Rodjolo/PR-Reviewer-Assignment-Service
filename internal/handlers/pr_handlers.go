package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Rodjolo/pr-reviewer-service/internal/service"
	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
	"github.com/Rodjolo/pr-reviewer-service/pkg/validator"
	"github.com/gorilla/mux"
)

// CreatePR godoc
// @Summary Создать Pull Request
// @Description Создает новый PR и автоматически назначает до 2 ревьюверов из команды автора
// @Tags PR
// @Accept json
// @Produce json
// @Param request body dto.CreatePRRequest true "Данные PR"
// @Success 201 {object} models.PR
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /prs [post]
func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, validator.FormatValidationErrors(err))
		return
	}

	pr, err := h.prService.CreatePR(req.Title, req.AuthorID)
	if err != nil {
		if errors.Is(err, service.ErrAuthorNotFound) || errors.Is(err, service.ErrAuthorNotInTeam) {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal server error")
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
		if errors.Is(err, service.ErrPRNotFound) {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal server error")
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
			h.respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		h.respondJSON(w, http.StatusOK, prs)
		return
	}

	prs, err := h.prService.GetAllPRs()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	h.respondJSON(w, http.StatusOK, prs)
}

// ReassignReviewer godoc
// @Summary Переназначить ревьювера
// @Description Заменяет одного ревьювера на случайного активного участника из команды заменяемого ревьювера
// @Tags PR
// @Accept json
// @Produce json
// @Param id path int true "ID PR"
// @Param request body dto.ReassignRequest true "Данные для переназначения"
// @Success 200 {object} models.PR
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "PR уже мержен"
// @Failure 500 {object} dto.ErrorResponse
// @Router /prs/{id}/reassign [patch]
func (h *Handlers) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prID, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid PR ID")
		return
	}

	var req dto.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, validator.FormatValidationErrors(err))
		return
	}

	pr, err := h.prService.ReassignReviewer(prID, req.OldReviewerID)
	if err != nil {
		if errors.Is(err, service.ErrPRNotFound) || errors.Is(err, service.ErrReviewerNotAssigned) || errors.Is(err, service.ErrReviewerNotInTeam) || errors.Is(err, service.ErrNoAvailableReviewers) {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, service.ErrPRAlreadyMerged) {
			h.respondError(w, http.StatusConflict, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal server error")
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
		if errors.Is(err, service.ErrPRNotFound) {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusOK, pr)
}
