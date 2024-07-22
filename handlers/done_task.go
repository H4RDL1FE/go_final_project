package handlers

import (
	"encoding/json"
	"fmt"
	"go_final_project/repository"
	"log"
	"net/http"
	"time"
)

func DoneTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
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

	log.Printf("Task before processing: %+v", task)

	if task.Repeat == "" {
		err := repo.DeleteTask(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting task: %v", err))
			return
		}
	} else {
		taskDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid task date format: %v", err))
			return
		}

		nextDate, err := NextDate(taskDate, task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error processing repeat rule: %v", err))
			return
		}

		log.Printf("Next date calculated: %s", nextDate)

		err = repo.UpdateTaskDate(id, nextDate)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task: %v", err))
			return
		}

		log.Printf("Task after processing: ID=%s, Date=%s", task.ID, nextDate)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}
