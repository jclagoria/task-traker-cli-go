package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: task-cli <command> [arguments]")
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
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\nUsage: task-cli <command> [arguments]\nCommands: add, list, update, delete, mark-in-progress, mark-done\n", os.Args[1])
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
	task, err := AddTask(os.Args[2])
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
	tasks, err := ListTasks(status)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		fmt.Printf("%d [%s] %s\n", t.ID, t.Status, t.Description)
	}
	return nil
}

func parseID(minArgs int) (int, error) {
	if len(os.Args) < minArgs {
		return 0, fmt.Errorf("id required")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return 0, fmt.Errorf("invalid ID %q", os.Args[2])
	}
	return id, nil
}

func cmdUpdate() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("id and description required")
	}
	id, err := parseID(4)
	if err != nil {
		return err
	}
	_, err = UpdateTask(id, os.Args[3])
	if err != nil {
		return err
	}
	fmt.Printf("Task %d updated successfully\n", id)
	return nil
}

func cmdDelete() error {
	id, err := parseID(3)
	if err != nil {
		return err
	}
	if err := DeleteTask(id); err != nil {
		return err
	}
	fmt.Printf("Task %d deleted successfully\n", id)
	return nil
}

func cmdMark(status Status) error {
	id, err := parseID(3)
	if err != nil {
		return err
	}
	_, err = MarkTask(id, status)
	if err != nil {
		return err
	}
	fmt.Printf("Task %d marked as %s\n", id, status)
	return nil
}
