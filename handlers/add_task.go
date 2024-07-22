package handlers

import (
	"encoding/json"
	"fmt"
	"go_final_project/model"
	"go_final_project/repository"
	"net/http"
	"time"
)

func AddTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var task model.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error decoding JSON: %v", err))
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

	id, err := repo.AddTask(task)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error inserting task into database: %v", err))
		return
	}

	response := map[string]interface{}{"id": fmt.Sprintf("%d", id)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
