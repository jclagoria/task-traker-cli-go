# Task Tracker CLI

A simple command line interface (CLI) to track your tasks and manage your to-do list. Tasks are stored in a JSON file in the current directory.

Project from [roadmap.sh](https://roadmap.sh/projects/task-tracker).

## Installation

```bash
git clone https://github.com/jclagoria/task-traker-cli-go.git
cd task-traker-cli-go
go build -o task-cli .
```

## Usage

### Adding a new task

```bash
task-cli add "Buy groceries"
# Output: Task added successfully (ID: 1)
```

### Updating a task

```bash
task-cli update 1 "Buy groceries and cook dinner"
```

### Deleting a task

```bash
task-cli delete 1
```

### Marking a task as in progress or done

```bash
task-cli mark-in-progress 1
task-cli mark-done 1
```

### Listing all tasks

```bash
task-cli list
```

### Listing tasks by status

```bash
task-cli list done
task-cli list todo
task-cli list in-progress
```

## Task Properties

Each task has the following properties:

- **id** - Unique identifier (auto-incremented)
- **description** - Short description of the task
- **status** - `todo`, `in-progress`, or `done`
- **createdAt** - Timestamp when the task was created (RFC3339)
- **updatedAt** - Timestamp when the task was last updated (RFC3339)

## Storage

Tasks are stored in `tasks.json` in the current directory. The file is created automatically when you add your first task.

To use a custom file path, set the `TASK_FILE` environment variable:

```bash
TASK_FILE=/path/to/tasks.json task-cli add "My task"
```

## Development

```bash
# Run tests
go test ./...

# Lint
golangci-lint run

# Build
go build -o task-cli .
```
