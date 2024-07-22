package handlers

import (
	"encoding/json"
	"fmt"
	"go_final_project/repository"
	"net/http"
)

func GetTasksHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	search := r.URL.Query().Get("search")
	tasks, err := repo.GetTasks(search)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error querying tasks: %v", err))
		return
	}

	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
