package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func TaskFilePath() string {
	if v := os.Getenv("TASK_FILE"); v != "" {
		return v
	}
	return "tasks.json"
}

func LoadTasks(filePath string) ([]Task, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	if len(data) == 0 {
		return []Task{}, nil
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	return tasks, nil
}

func SaveTasks(filePath string, tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tasks: %w", err)
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", tmp, err)
	}

	if err := os.Rename(tmp, filePath); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmp, filePath, err)
	}
	return nil
}

func AddTask(filePath, description string) (Task, error) {
	tasks, err := LoadTasks(filePath)
	if err != nil {
		return Task{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := 1
	if len(tasks) > 0 {
		id = tasks[len(tasks)-1].ID + 1
	}

	task := Task{
		ID:          id,
		Description: description,
		Status:      "todo",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tasks = append(tasks, task)
	if err := SaveTasks(filePath, tasks); err != nil {
		return Task{}, err
	}
	return task, nil
}
