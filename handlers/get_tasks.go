package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"net/http"

	// Внутренние библиотеки
	"go_final_project/repository"
)

func GetTasksHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	search := r.URL.Query().Get("search")
	tasks, err := repo.GetTasks(search)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении задач")
		return
	}

	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
