package task

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/internal/repository"
)


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
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer writer.Flush()

	strBuilder.WriteString("Name\tPath\tCommands\tIcon\tIcon Color\n")
	for _, task := range tasks {
		strBuilder.WriteString(task.Name + "\t")

		if strings.HasPrefix(task.Path, os.Getenv("HOME")) {
			task.Path = strings.Replace(task.Path, os.Getenv("HOME"), "~", 1)
		}
		strBuilder.WriteString(task.Path + "\t")
		strBuilder.WriteString(strings.Join(task.Cmds, ", ") + "\t")
		strBuilder.WriteString(task.Icon + "\t")
		strBuilder.WriteString(task.IconColor + "\n")
	}
	fmt.Fprintln(writer, strBuilder.String())
	return err
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