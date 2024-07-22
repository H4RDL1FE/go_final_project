package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/model"
	"go_final_project/repository"
	"net/http"
	"time"
)

func EditTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var task model.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding JSON: %v", err))
		return
	}

	if task.ID == "" {
		respondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	if task.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	} else {
		_, err := time.Parse("20060102", task.Date)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
	}

	if task.Date < now.Format("20060102") {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error processing repeat rule: %v", err))
				return
			}
			task.Date = nextDate
		}
	}

	db, err := sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error opening database: %v", err))
		return
	}
	defer db.Close()

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task: %v", err))
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving rows affected: %v", err))
		return
	}

	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}
