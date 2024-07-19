package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

//region NextDate

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("empty repeat rule")
	}

	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	// Разбор правила повторения
	repeatParts := strings.Split(repeat, " ")
	if len(repeatParts) < 1 || len(repeatParts) > 2 {
		return "", fmt.Errorf("invalid repeat rule: %s", repeat)
	}

	rule := repeatParts[0]
	var value string
	if len(repeatParts) == 2 {
		value = repeatParts[1]
	}

	switch rule {
	case "d":
		days, err := strconv.Atoi(value)
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("invalid day interval: %s", value)
		}
		taskDate = taskDate.AddDate(0, 0, days)
		for taskDate.Before(now) || taskDate.Equal(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
		return taskDate.Format("20060102"), nil
	case "y":
		if value != "" {
			return "", fmt.Errorf("invalid repeat rule: %s", repeat)
		}
		taskDate = taskDate.AddDate(1, 0, 0)
		for taskDate.Before(now) || taskDate.Equal(now) {
			taskDate = taskDate.AddDate(1, 0, 0)
		}
		return taskDate.Format("20060102"), nil
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", rule)
	}
}

//endregion NextDate

//region addTaskHandler

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var task Task
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

	db, err := sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error opening database: %v", err))
		return
	}
	defer db.Close()

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error inserting task into database: %v", err))
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving last insert id: %v", err))
		return
	}

	response := map[string]interface{}{"id": fmt.Sprintf("%d", id)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//endregion addTaskHandler

//region apiNextDateHandler

func apiNextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "invalid now date format", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, nextDate)
}

//endregion apiNextDateHandler

//region getTasksHandler

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "scheduler.db")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error opening database: %v", err))
		return
	}
	defer db.Close()

	search := r.URL.Query().Get("search")
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1"
	args := []interface{}{}

	if search != "" {
		if searchDate, err := time.Parse("02.01.2006", search); err == nil {
			query += " AND date = ?"
			args = append(args, searchDate.Format("20060102"))
		} else {
			query += " AND (title LIKE ? OR comment LIKE ?)"
			searchPattern := "%" + search + "%"
			args = append(args, searchPattern, searchPattern)
		}
	}

	query += " ORDER BY date LIMIT 50"

	rows, err := db.Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error querying tasks: %v", err))
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error scanning task: %v", err))
			return
		}
		task.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error with rows: %v", err))
		return
	}

	if tasks == nil {
		tasks = []Task{}
	}

	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//endregion getTasksHandler

//region getTaskHandler

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	var task Task
	err = db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Задача не найдена")
		} else {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving task: %v", err))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

//endregion getTaskHandler

//region editTaskHandler

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var task Task
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

//endregion editTaskHandler

//region doneTaskHandler

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	var task Task
	err = db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Задача не найдена")
		} else {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error retrieving task: %v", err))
		}
		return
	}

	log.Printf("Task before processing: %+v", task)

	if task.Repeat == "" {
		_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
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

		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating task: %v", err))
			return
		}

		log.Printf("Task after processing: ID=%s, Date=%s", task.ID, nextDate)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}

//endregion doneTaskHandler

//region deleteTaskHandler

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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

//endregion deleteTaskHandler

//region respondWithError

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("Error: %s", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

//endregion respondWithError

//region main

func main() {
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(appPath, "scheduler.db")
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		dbFile = envDBFile
	}

	log.Printf("Используется файл базы данных: %s", dbFile)

	_, err = os.Stat(dbFile)
	install := false
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if install {
		log.Println("Создание таблицы scheduler...")
		createTableSQL := `CREATE TABLE scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT
        );`
		_, err = db.Exec(createTableSQL)
		if err != nil {
			log.Fatal(err)
		}

		createIndexSQL := `CREATE INDEX idx_date ON scheduler(date);`
		_, err = db.Exec(createIndexSQL)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Таблица scheduler создана успешно")
	} else {
		log.Println("Таблица scheduler уже существует")
	}

	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", apiNextDateHandler)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTaskHandler(w, r)
		case http.MethodPost:
			addTaskHandler(w, r)
		case http.MethodPut:
			editTaskHandler(w, r)
		case http.MethodDelete:
			deleteTaskHandler(w, r)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/tasks", getTasksHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)

	fmt.Printf("Запуск сервера на порту %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

//endregion main
