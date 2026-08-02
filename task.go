// Package main implements a CLI task tracker with JSON file storage.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

// Status represents the state of a task.
type Status string

const (
	// StatusTodo is the initial state of a newly created task.
	StatusTodo Status = "todo"
	// StatusInProgress indicates a task is being worked on.
	StatusInProgress Status = "in-progress"
	// StatusDone indicates a completed task.
	StatusDone Status = "done"
)

func validStatus(s Status) bool {
	return s == StatusTodo || s == StatusInProgress || s == StatusDone
}

// Task represents a single task with metadata.
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Status      Status `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// LoadTasks reads and unmarshals tasks from the JSON file.
func LoadTasks(filePath string) ([]Task, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return nil, fmt.Errorf("locking %s: %w", filePath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	if len(data) == 0 {
		return []Task{}, nil
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("%s is corrupted", filePath)
	}
	return tasks, nil
}

// SaveTasks marshals and writes tasks to the JSON file.
func SaveTasks(filePath string, tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tasks: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("locking %s: %w", filePath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncating %s: %w", filePath, err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("seeking %s: %w", filePath, err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing %s: %w", filePath, err)
	}
	return nil
}

// rwTasks opens the file, acquires an exclusive flock, reads tasks, calls fn,
// and writes the result back — all under a single lock.
func rwTasks(filePath string, fn func([]Task) ([]Task, error)) ([]Task, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return nil, fmt.Errorf("locking %s: %w", filePath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	var tasks []Task
	if len(data) > 0 {
		if err := json.Unmarshal(data, &tasks); err != nil {
			return nil, fmt.Errorf("%s is corrupted", filePath)
		}
	}

	result, err := fn(tasks)
	if err != nil {
		return nil, err
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling tasks: %w", err)
	}

	if err := f.Truncate(0); err != nil {
		return nil, fmt.Errorf("truncating %s: %w", filePath, err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seeking %s: %w", filePath, err)
	}
	if _, err := f.Write(out); err != nil {
		return nil, fmt.Errorf("writing %s: %w", filePath, err)
	}
	return result, nil
}

// AddTask creates a new task with auto-incremented ID and status todo.
func AddTask(filePath, description string) (Task, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var added Task

	result, err := rwTasks(filePath, func(tasks []Task) ([]Task, error) {
		id := 1
		if len(tasks) > 0 {
			id = tasks[len(tasks)-1].ID + 1
		}
		added = Task{
			ID:          id,
			Description: description,
			Status:      StatusTodo,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		return append(tasks, added), nil
	})
	if err != nil {
		return Task{}, err
	}
	_ = result
	return added, nil
}

// ListTasks returns all tasks, optionally filtered by status.
func ListTasks(filePath string, status Status) ([]Task, error) {
	tasks, err := LoadTasks(filePath)
	if err != nil {
		return nil, err
	}

	if status == "" {
		return tasks, nil
	}

	var filtered []Task
	for _, t := range tasks {
		if t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// modifyTask locks, loads, mutates a task by ID, and saves.
func modifyTask(filePath string, id int, mutate func(*Task)) (Task, error) {
	var modified Task

	_, err := rwTasks(filePath, func(tasks []Task) ([]Task, error) {
		for i, t := range tasks {
			if t.ID == id {
				mutate(&tasks[i])
				tasks[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
				modified = tasks[i]
				return tasks, nil
			}
		}
		return nil, fmt.Errorf("task with ID %d not found", id)
	})
	if err != nil {
		return Task{}, err
	}
	return modified, nil
}

// UpdateTask updates the description of an existing task by ID.
func UpdateTask(filePath string, id int, description string) (Task, error) {
	return modifyTask(filePath, id, func(t *Task) { t.Description = description })
}

// DeleteTask removes a task by ID.
func DeleteTask(filePath string, id int) error {
	_, err := rwTasks(filePath, func(tasks []Task) ([]Task, error) {
		for i, t := range tasks {
			if t.ID == id {
				return append(tasks[:i], tasks[i+1:]...), nil
			}
		}
		return nil, fmt.Errorf("task with ID %d not found", id)
	})
	return err
}

// MarkTask updates the status of an existing task by ID.
func MarkTask(filePath string, id int, status Status) (Task, error) {
	if !validStatus(status) {
		return Task{}, fmt.Errorf("invalid status %q, must be one of: todo, in-progress, done", status)
	}
	return modifyTask(filePath, id, func(t *Task) { t.Status = status })
}
