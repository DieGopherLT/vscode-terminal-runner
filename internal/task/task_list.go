package task

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
)


// listAllTasks displays all tasks in a formatted table.
func listAllTasks() error {
	tasks, err := repository.ReadTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	var strBuilder strings.Builder
	writer := tabwriter.NewWriter(&strBuilder, 0, 0, 1, ' ', 0)

	fmt.Fprintln(writer, "Name\tPath\tCommands\tIcon\tIcon Color")
	for _, task := range tasks {
		formattedPath := task.Path
		if strings.HasPrefix(task.Path, os.Getenv("HOME")) {
			formattedPath = strings.Replace(task.Path, os.Getenv("HOME"), "~", 1)
		}
		
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			task.Name,
			formattedPath,
			strings.Join(task.Cmds, ", "),
			task.Icon,
			task.IconColor,
		)
	}
	
	writer.Flush()
	fmt.Print(strBuilder.String())
	return nil
}

func listAllTaskNames() error {
	tasks, err := repository.ReadTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	var strBuilder strings.Builder
	strBuilder.WriteString("Task Names:\n")
	for _, task := range tasks {
		strBuilder.WriteString("- " + task.Name + "\n")
	}
	fmt.Println(strBuilder.String())
	return nil
}

// FindByName retrieves a task by its name from the saved tasks
func FindByName(name string) (*models.Task, error) {
	return repository.FindTaskByName(name)
}