// Package main implements a CLI task tracker with JSON file storage.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
)

var fileLocks = make(map[string]*sync.Mutex)
var fileLocksMu sync.Mutex

func getFileLock(path string) *sync.Mutex {
	fileLocksMu.Lock()
	defer fileLocksMu.Unlock()
	if fileLocks[path] == nil {
		fileLocks[path] = &sync.Mutex{}
	}
	return fileLocks[path]
}

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

// TaskFilePath returns the path to the JSON task file, defaulting to tasks.json.
func TaskFilePath() string {
	if v := os.Getenv("TASK_FILE"); v != "" {
		return v
	}
	return "tasks.json"
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

// AddTask creates a new task with auto-incremented ID and status todo.
func AddTask(filePath, description string) (Task, error) {
	getFileLock(filePath).Lock()
	defer getFileLock(filePath).Unlock()

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
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tasks = append(tasks, task)
	if err := SaveTasks(filePath, tasks); err != nil {
		return Task{}, err
	}
	return task, nil
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

// UpdateTask updates the description of an existing task by ID.
func UpdateTask(filePath string, id int, description string) (Task, error) {
	getFileLock(filePath).Lock()
	defer getFileLock(filePath).Unlock()

	tasks, err := LoadTasks(filePath)
	if err != nil {
		return Task{}, err
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Description = description
			tasks[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			if err := SaveTasks(filePath, tasks); err != nil {
				return Task{}, err
			}
			return tasks[i], nil
		}
	}
	return Task{}, fmt.Errorf("task with ID %d not found", id)
}

// DeleteTask removes a task by ID.
func DeleteTask(filePath string, id int) error {
	getFileLock(filePath).Lock()
	defer getFileLock(filePath).Unlock()

	tasks, err := LoadTasks(filePath)
	if err != nil {
		return err
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return SaveTasks(filePath, tasks)
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// MarkTask updates the status of an existing task by ID.
func MarkTask(filePath string, id int, status Status) (Task, error) {
	if !validStatus(status) {
		return Task{}, fmt.Errorf("invalid status %q, must be one of: todo, in-progress, done", status)
	}

	getFileLock(filePath).Lock()
	defer getFileLock(filePath).Unlock()

	tasks, err := LoadTasks(filePath)
	if err != nil {
		return Task{}, err
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Status = status
			tasks[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			if err := SaveTasks(filePath, tasks); err != nil {
				return Task{}, err
			}
			return tasks[i], nil
		}
	}
	return Task{}, fmt.Errorf("task with ID %d not found", id)
}
