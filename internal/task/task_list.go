package task

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

func readTasks() ([]Task, error) {
	file, err := os.OpenFile(TasksSaveFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsonContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var content TaskSaveFileContent
	if len(jsonContent) > 0 {
		if err = json.Unmarshal(jsonContent, &content); err != nil {
			return nil, err
		}
	}

	return content.Tasks, nil
}

func listAllTasks() error {
	tasks, err := readTasks()
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
	tasks, err := readTasks()
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
