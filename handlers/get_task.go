package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"net/http"

	// Внутренние библиотеки
	"go_final_project/repository"
)

func GetTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	id := r.URL.Query().Get("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	task, err := repo.GetTask(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении задачи")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
