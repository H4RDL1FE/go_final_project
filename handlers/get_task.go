package handlers

import (
	"encoding/json"
	"fmt"
	"go_final_project/repository"
	"net/http"
)

func GetTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	id := r.URL.Query().Get("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	task, err := repo.GetTask(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving task: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
