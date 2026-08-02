package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func useTempFile(t *testing.T) {
	t.Helper()
	taskFile = filepath.Join(t.TempDir(), "tasks.json")
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	useTempFile(t)
	tasks := []Task{
		{ID: 1, Description: "Buy milk", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
		{ID: 2, Description: "Walk dog", Status: StatusDone, CreatedAt: "2026-01-02T00:00:00Z", UpdatedAt: "2026-01-02T00:00:00Z"},
	}

	if err := saveTasks(tasks); err != nil {
		t.Fatalf("saveTasks: %v", err)
	}

	loaded, err := LoadTasks()
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
	useTempFile(t)
	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks on missing file: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(tasks))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	useTempFile(t)
	if err := os.WriteFile(taskFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks on empty file: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(tasks))
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	useTempFile(t)
	if err := os.WriteFile(taskFile, []byte("{bad json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadTasks()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "is corrupted") {
		t.Errorf("error %q should contain %q", err.Error(), "is corrupted")
	}
}

func TestAddTaskToEmptyList(t *testing.T) {
	useTempFile(t)
	task, err := AddTask("Buy groceries")
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
	useTempFile(t)
	_, _ = AddTask("First")
	task2, _ := AddTask("Second")

	if task2.ID != 2 {
		t.Errorf("second task ID = %d, want 2", task2.ID)
	}

	tasks, _ := LoadTasks()
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
}

func TestAddTaskPersistsToFile(t *testing.T) {
	useTempFile(t)
	_, _ = AddTask("Persist me")

	loaded, err := LoadTasks()
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
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
	})

	tasks, err := ListTasks(Status(""))
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
}

func TestListFilterByStatus(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
		{ID: 3, Description: "C", Status: StatusTodo},
	})

	tasks, _ := ListTasks(StatusTodo)
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
	if tasks[0].Description != "A" || tasks[1].Description != "C" {
		t.Errorf("got %v, want [A, C]", tasks)
	}
}

func TestListEmptyFile(t *testing.T) {
	useTempFile(t)
	tasks, err := ListTasks(Status(""))
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("got %d, want 0", len(tasks))
	}
}

func TestListNoMatch(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	tasks, _ := ListTasks(StatusDone)
	if len(tasks) != 0 {
		t.Fatalf("got %d, want 0", len(tasks))
	}
}

func TestListInvalidStatusFilter(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := ListTasks("banana")
	if err == nil {
		t.Fatal("expected error for invalid status filter, got nil")
	}
}

func TestUpdateTaskDescription(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, err := UpdateTask(1, "New")
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Description != "New" {
		t.Errorf("Description = %q, want %q", updated.Description, "New")
	}
}

func TestUpdateTaskPreservesCreatedAt(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, _ := UpdateTask(1, "New")
	if updated.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt changed to %q, should be unchanged", updated.CreatedAt)
	}
}

func TestUpdateTaskSetsUpdatedAt(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "Old", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	updated, _ := UpdateTask(1, "New")
	if updated.UpdatedAt == "2026-01-01T00:00:00Z" {
		t.Error("UpdatedAt was not changed")
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := UpdateTask(99, "New")
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestDeleteTask(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusDone},
	})

	if err := DeleteTask(1); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	tasks, _ := LoadTasks()
	if len(tasks) != 1 {
		t.Fatalf("got %d, want 1", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("remaining ID = %d, want 2", tasks[0].ID)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	err := DeleteTask(99)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestDeletePreservesOtherIDs(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
		{ID: 2, Description: "B", Status: StatusTodo},
		{ID: 3, Description: "C", Status: StatusTodo},
	})

	_ = DeleteTask(2)

	tasks, _ := LoadTasks()
	if len(tasks) != 2 {
		t.Fatalf("got %d, want 2", len(tasks))
	}
	if tasks[0].ID != 1 || tasks[1].ID != 3 {
		t.Errorf("IDs = [%d, %d], want [1, 3]", tasks[0].ID, tasks[1].ID)
	}
}

func TestMarkTaskFromTodo(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z"},
	})

	marked, err := MarkTask(1, StatusInProgress)
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
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusDone},
	})

	marked, _ := MarkTask(1, StatusInProgress)
	if marked.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", marked.Status, StatusInProgress)
	}
}

func TestMarkTaskNotFound(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := MarkTask(99, StatusInProgress)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMarkTaskInvalidStatus(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{
		{ID: 1, Description: "A", Status: StatusTodo},
	})

	_, err := MarkTask(1, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestConcurrentWrites(t *testing.T) {
	useTempFile(t)
	_ = saveTasks([]Task{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = AddTask("task")
		}(i)
	}
	wg.Wait()

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(tasks) != 10 {
		t.Errorf("got %d tasks, want 10", len(tasks))
	}
}
