package repository

import (
	"fmt"
	"go_final_project/model"
	"time"
)

func (r *Repository) AddTask(task model.Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *Repository) GetTask(id string) (model.Task, error) {
	var task model.Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err := r.DB.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, err
	}
	return task, nil
}

func (r *Repository) GetTasks(search string) ([]model.Task, error) {
	tasks := []model.Task{}
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

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		task.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	return tasks, nil
}

func (r *Repository) EditTask(task model.Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := r.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	return err
}

func (r *Repository) UpdateTaskDate(id string, date string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	_, err := r.DB.Exec(query, date, id)
	return err
}
