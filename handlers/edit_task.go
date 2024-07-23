package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	// Внутренние библиотеки
	"go_final_project/model"
	"go_final_project/repository"
)

func EditTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
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

	err = repo.EditTask(task)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"result": "success"}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
