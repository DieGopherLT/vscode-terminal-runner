package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
	"github.com/charmbracelet/lipgloss"
)

var (
	ErrSetupCompleted = errors.New("setup already completed")
	ErrSetupFailed    = errors.New("setup failed")
)

// Setup initializes the CLI tool with welcome message and extension requirements.
func Setup() error {

	config, err := Load()
	if err != nil {
		return err
	}

	if config.IsSetupComplete {
		return ErrSetupCompleted
	}

	fmt.Print(GetWelcomeMessage())
	time.Sleep(2 * time.Second)
	fmt.Print(GetExtensionRequirement())

	var response string
	fmt.Scanln(&response)

	if response != "y" {
		return nil
	}

	// Invoke a process that executes the command installation
	command := exec.Command("code", "--install extension DiegoGopherLT.vstr-bridge")
	output, err := command.CombinedOutput()
	if err != nil {
		return err
	}

	styles.PrintInfo(string(output))

	brandNewConfig := models.Config{
		IsSetupComplete: true,
	}

	file, err := os.Create(ConfigurationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(brandNewConfig)
	if err != nil {
		return err
	}

	return nil
}

// GetWelcomeMessage returns a styled welcome message for new users.
func GetWelcomeMessage() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.VSCodeBlue).
		Bold(true).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.VSCodeBlue).
		Padding(1, 2).
		MarginBottom(1).
		Width(70)

	welcomeStyle := lipgloss.NewStyle().
		Foreground(styles.White).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.LightBlue).
		Background(lipgloss.Color("#1E1E1E")).
		Padding(1, 2).
		MarginBottom(2).
		Width(70)

	title := "🚀 VSCode Terminal Runner"
	welcomeText := `Automate your development workflow
Launch multiple projects with a single command

⚡ Perfect for microservices and full-stack setups`

	return titleStyle.Render(title) + "\n" + welcomeStyle.Render(welcomeText)
}

// GetExtensionRequirement returns information about the required VSCode extension.
func GetExtensionRequirement() string {
	requirementStyle := lipgloss.NewStyle().
		Foreground(styles.Warning).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Warning).
		Background(lipgloss.Color("#2D241B")).
		Padding(1, 2).
		MarginBottom(2).
		Width(70)

	linkStyle := lipgloss.NewStyle().
		Foreground(styles.VSCodeBlue).
		Underline(true).
		Bold(true)

	extensionText := "⚠️  REQUIRED: VSCode Extension\n\n" +
		"This CLI requires the VSTR-Bridge extension to work.\n\n" +
		"📦 Install from: " + linkStyle.Render("https://github.com/DieGopherLT/VSTR-Bridge") + "\n\n" +
		"🔍 Search in VSCode: " + linkStyle.Render("vstr-bridge") + "\n\n" +
		"Or install automatically: Do you want to install it now? (y/n): "

	return requirementStyle.Render(extensionText)
}
