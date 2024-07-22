package main

import (
	"go_final_project/handlers"
	"go_final_project/middleware"
	"go_final_project/repository"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dbFile := filepath.Join(appPath, "scheduler.db")
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		dbFile = envDBFile
	}

	repo, err := repository.NewRepository(dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	webDir := "./web"

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.APINextDateHandler)
	http.HandleFunc("/api/signin", handlers.SigninHandler)
	http.HandleFunc("/api/task", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTaskHandler(w, r, repo)
		case http.MethodPost:
			handlers.AddTaskHandler(w, r, repo)
		case http.MethodPut:
			handlers.EditTaskHandler(w, r, repo)
		case http.MethodDelete:
			handlers.DeleteTaskHandler(w, r, repo)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/tasks", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetTasksHandler(w, r, repo)
	}))
	http.HandleFunc("/api/task/done", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.DoneTaskHandler(w, r, repo)
	}))

	log.Printf("Запуск сервера на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
