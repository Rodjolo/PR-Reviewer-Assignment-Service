package handlers

import (
	"net/http"
)

// GetStats godoc
// @Summary Получить статистику
// @Description Возвращает статистику по пользователям, командам и PR'ам
// @Tags Stats
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 500 {object} map[string]string
// @Router /stats [get]
func (h *Handlers) GetStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsService.GetStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, stats)
}
