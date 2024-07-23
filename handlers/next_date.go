package handlers

import (
	// Стандартные библиотеки

	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("empty repeat rule")
	}

	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

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

func APINextDateHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, err := fmt.Fprint(w, nextDate); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}
