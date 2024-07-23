package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"net/http"

	// Внутренние библиотеки
	"go_final_project/repository"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	id := r.URL.Query().Get("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	err := repo.DeleteTask(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при удалении задачи")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"result": "success"}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
