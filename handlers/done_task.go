package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"net/http"
	"time"

	// Внутренние библиотеки
	"go_final_project/repository"
)

func DoneTaskHandler(w http.ResponseWriter, r *http.Request, repo *repository.Repository) {
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

	if task.Repeat == "" {
		err := repo.DeleteTask(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка при удалении задачи")
			return
		}
	} else {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Ошибка при вычислении следующей даты")
			return
		}

		err = repo.UpdateTaskDate(id, nextDate)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении даты задачи")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"result": "success"}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
