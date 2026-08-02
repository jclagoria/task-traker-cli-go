package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	tasks := []Task{
		{ID: 1, Description: "Buy milk", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
		{ID: 2, Description: "Walk dog", Status: StatusDone, CreatedAt: "2026-01-02T00:00:00Z", UpdatedAt: "2026-01-02T00:00:00Z"},
	}

	if err := SaveTasks(path, tasks); err != nil {
		t.Fatalf("SaveTasks: %v", err)
	}

	loaded, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("got %d tasks, want 2", len(loaded))
	}
	if loaded[0].Description != "Buy milk" {
		t.Errorf("task 1 description = %q, want %q", loaded[0].Description, "Buy milk")
	}
	if loaded[1].Status != StatusDone {
		t.Errorf("task 2 status = %q, want %q", loaded[1].Status, StatusDone)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	tasks, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks on missing file: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(tasks))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks on empty file: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(tasks))
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "is corrupted") {
		t.Errorf("error %q should contain %q", err.Error(), "is corrupted")
	}
}

func TestTaskFilePathDefault(t *testing.T) {
	got := TaskFilePath()
	if got != "tasks.json" {
		t.Errorf("TaskFilePath() = %q, want %q", got, "tasks.json")
	}
}

func TestTaskFilePathEnvOverride(t *testing.T) {
	t.Setenv("TASK_FILE", "/custom/path.json")
	got := TaskFilePath()
	if got != "/custom/path.json" {
		t.Errorf("TaskFilePath() = %q, want %q", got, "/custom/path.json")
	}
}

func TestTaskFilePathEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env.json")
	t.Setenv("TASK_FILE", path)

	tasks := []Task{{ID: 1, Description: "Env task", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"}}
	if err := SaveTasks(path, tasks); err != nil {
		t.Fatalf("SaveTasks: %v", err)
	}

	loaded, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Description != "Env task" {
		t.Errorf("got %v, want 1 task with description 'Env task'", loaded)
	}
}

func TestAddTaskToEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	task, err := AddTask(path, "Buy groceries")
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("ID = %d, want 1", task.ID)
	}
	if task.Description != "Buy groceries" {
		t.Errorf("Description = %q, want %q", task.Description, "Buy groceries")
	}
	if task.Status != StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, StatusTodo)
	}
	if task.CreatedAt == "" {
		t.Error("CreatedAt is empty")
	}
	if task.UpdatedAt == "" {
		t.Error("UpdatedAt is empty")
	}
}

func TestAddTaskIncrementID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_, _ = AddTask(path, "First")
	task2, _ := AddTask(path, "Second")

	if task2.ID != 2 {
		t.Errorf("second task ID = %d, want 2", task2.ID)
	}

	tasks, _ := LoadTasks(path)
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
}

func TestAddTaskPersistsToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_, _ = AddTask(path, "Persist me")

	loaded, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("got %d tasks, want 1", len(loaded))
	}
	if loaded[0].Description != "Persist me" {
		t.Errorf("Description = %q, want %q", loaded[0].Description, "Persist me")
	}
}

func TestListAllTasks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
	})

	tasks, err := ListTasks(path, Status(""))
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
}

func TestListFilterByStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
		{ID: 3, Description: "C", Status: StatusTodo},
	})

	tasks, _ := ListTasks(path, StatusTodo)
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
	if tasks[0].Description != "A" || tasks[1].Description != "C" {
		t.Errorf("got %v, want [A, C]", tasks)
	}
}

func TestListEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	tasks, err := ListTasks(path, Status(""))
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d, want 0", len(tasks))
	}
}

func TestListNoMatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	tasks, _ := ListTasks(path, StatusDone)
	if len(tasks) != 0 {
		t.Fatalf("got %d, want 0", len(tasks))
	}
}

func TestUpdateTaskDescription(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, err := UpdateTask(path, 1, "New")
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Description != "New" {
		t.Errorf("Description = %q, want %q", updated.Description, "New")
	}
}

func TestUpdateTaskPreservesCreatedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, _ := UpdateTask(path, 1, "New")
	if updated.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt changed to %q, should be unchanged", updated.CreatedAt)
	}
}

func TestUpdateTaskSetsUpdatedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, _ := UpdateTask(path, 1, "New")
	if updated.UpdatedAt == "2026-01-01T00:00:00Z" {
		t.Error("UpdatedAt was not changed")
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := UpdateTask(path, 99, "New")
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestDeleteTask(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
	})

	if err := DeleteTask(path, 1); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	tasks, _ := LoadTasks(path)
	if len(tasks) != 1 {
		t.Fatalf("got %d, want 1", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("remaining ID = %d, want 2", tasks[0].ID)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	err := DeleteTask(path, 99)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestDeletePreservesOtherIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusTodo},
		{ID: 3, Description: "C", Status: StatusTodo},
	})

	_ = DeleteTask(path, 2)

	tasks, _ := LoadTasks(path)
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
	if tasks[0].ID != 1 || tasks[1].ID != 3 {
		t.Errorf("IDs = [%d, %d], want [1, 3]", tasks[0].ID, tasks[1].ID)
	}
}

func TestMarkTaskFromTodo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	marked, err := MarkTask(path, 1, StatusInProgress)
	if err != nil {
		t.Fatalf("MarkTask: %v", err)
	}
	if marked.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", marked.Status, StatusInProgress)
	}
	if marked.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt changed to %q", marked.CreatedAt)
	}
}

func TestMarkTaskFromDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusDone},
	})

	marked, _ := MarkTask(path, 1, StatusInProgress)
	if marked.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", marked.Status, StatusInProgress)
	}
}

func TestMarkTaskNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := MarkTask(path, 99, StatusInProgress)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMarkTaskInvalidStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := MarkTask(path, 1, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	_ = SaveTasks(path, []Task{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = AddTask(path, "task")
		}(i)
	}
	wg.Wait()

	tasks, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(tasks) != 10 {
		t.Errorf("got %d tasks, want 10", len(tasks))
	}
}
