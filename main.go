package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: task-traker-cli <command> [arguments]")
		return
	}

	var err error
	switch os.Args[1] {
	case "add":
		err = cmdAdd()
	case "list":
		err = cmdList()
	case "update":
		err = cmdUpdate()
	case "delete":
		err = cmdDelete()
	case "mark-in-progress":
		err = cmdMark(StatusInProgress)
	case "mark-done":
		err = cmdMark(StatusDone)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\nUsage: task-traker-cli <command> [arguments]\nCommands: add, list, update, delete, mark-in-progress, mark-done\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cmdAdd() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("description required")
	}
	task, err := AddTask(TaskFilePath(), os.Args[2])
	if err != nil {
		return err
	}
	fmt.Printf("Task added successfully (ID: %d)\n", task.ID)
	return nil
}

func cmdList() error {
	status := Status("")
	if len(os.Args) >= 3 {
		status = Status(os.Args[2])
	}
	tasks, err := ListTasks(TaskFilePath(), status)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		fmt.Printf("%d [%s] %s\n", t.ID, t.Status, t.Description)
	}
	return nil
}

func cmdUpdate() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("id and description required")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("invalid ID %q", os.Args[2])
	}
	_, err = UpdateTask(TaskFilePath(), id, os.Args[3])
	if err != nil {
		return err
	}
	fmt.Printf("Task %d updated successfully\n", id)
	return nil
}

func cmdDelete() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("id required")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("invalid ID %q", os.Args[2])
	}
	if err := DeleteTask(TaskFilePath(), id); err != nil {
		return err
	}
	fmt.Printf("Task %d deleted successfully\n", id)
	return nil
}

func cmdMark(status Status) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("id required")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("invalid ID %q", os.Args[2])
	}
	_, err = MarkTask(TaskFilePath(), id, status)
	if err != nil {
		return err
	}
	fmt.Printf("Task %d marked as %s\n", id, status)
	return nil
}
