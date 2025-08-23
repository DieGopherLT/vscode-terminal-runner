/*
Copyright © 2025 Diego López Torres diego.lpz.trrs.dev@gmail.com
*/
package main

import (
	"github.com/DieGopherLT/vscode-terminal-runner/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
