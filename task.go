package main

import (
	"encoding/json"
	"fmt"
	"os"
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
