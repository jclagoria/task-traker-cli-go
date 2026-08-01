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

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: description required")
			os.Exit(1)
		}
		task, err := AddTask(TaskFilePath(), os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task added successfully (ID: %d)\n", task.ID)
	case "list":
		status := ""
		if len(os.Args) >= 3 {
			status = os.Args[2]
		}
		tasks, err := ListTasks(TaskFilePath(), status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		for _, t := range tasks {
			fmt.Printf("%d [%s] %s\n", t.ID, t.Status, t.Description)
		}
	case "update":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Error: id and description required")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID %q\n", os.Args[2])
			os.Exit(1)
		}
		_, err = UpdateTask(TaskFilePath(), id, os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task %d updated successfully\n", id)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: id required")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID %q\n", os.Args[2])
			os.Exit(1)
		}
		if err := DeleteTask(TaskFilePath(), id); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task %d deleted successfully\n", id)
	case "mark-in-progress", "mark-done":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: id required")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid ID %q\n", os.Args[2])
			os.Exit(1)
		}
		status := "in-progress"
		if os.Args[1] == "mark-done" {
			status = "done"
		}
		_, err = MarkTask(TaskFilePath(), id, status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Task %d marked as %s\n", id, status)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}
