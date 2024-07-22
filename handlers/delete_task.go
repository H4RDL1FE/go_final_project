package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/repository"
	"net/http"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	id := r.URL.Query().Get("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	db, err := sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error opening database: %v", err))
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting task: %v", err))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting rows affected: %v", err))
		return
	}

	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}
