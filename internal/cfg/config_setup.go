package cfg

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

	// Display welcome message with a brief pause for better UX
	fmt.Print(getWelcomeMessage())
	time.Sleep(1500 * time.Millisecond)

	// Check if extension is already installed
	if IsExtensionInstalled() {
		styles.PrintSuccess("VSTR-Bridge extension is already installed!")
		return completeSetup()
	}

	// Show extension requirement information
	fmt.Print(getExtensionRequirement())

	// Get user choice with enhanced options
	choice := getInstallationChoice()

	switch choice {
	case "y", "yes", "":
		// Install the extension
		if err := installExtension(); err != nil {
			styles.PrintError(fmt.Sprintf("Failed to install extension: %v", err))
			styles.PrintInfo("You can install it manually from: https://github.com/DieGopherLT/VSTR-Bridge")
			return err
		}
		return completeSetup()

	default:
		styles.PrintWarning("The VSTR-Bridge extension is required for this CLI to work.")
		styles.PrintInfo("You can:")
		styles.PrintInfo("  • Run 'vstr setup' again when ready to install")
		styles.PrintInfo("  • Install manually from: https://github.com/DieGopherLT/VSTR-Bridge")
		styles.PrintInfo("  • Search 'vstr-bridge' in VSCode extensions")
		return nil
	}
}

// GetWelcomeMessage returns a styled welcome message for new users.
func getWelcomeMessage() string {
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

	title := "VSCode Terminal Runner"
	welcomeText := `Automate your development workflow
Launch multiple projects with a single command

Perfect for microservices and full-stack setups`

	return titleStyle.Render(title) + "\n" + welcomeStyle.Render(welcomeText)
}

// GetExtensionRequirement returns information about the required VSCode extension with minimal inline styling.
func getExtensionRequirement() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Warning).
		Bold(true)

	linkStyle := lipgloss.NewStyle().
		Foreground(styles.VSCodeBlue).
		Underline(true)

	accentStyle := lipgloss.NewStyle().
		Foreground(styles.LightGray)

	var message strings.Builder
	message.WriteString("\n")
	message.WriteString(headerStyle.Render("▲ Extension Required"))
	message.WriteString("\n\n")
	message.WriteString(accentStyle.Render("This CLI requires the VSTR-Bridge extension to work with VSCode."))
	message.WriteString("\n\n")
	message.WriteString(accentStyle.Render("Install manually: ") + linkStyle.Render("https://github.com/DieGopherLT/VSTR-Bridge"))
	message.WriteString("\n")
	message.WriteString(accentStyle.Render("Search in VSCode: ") + linkStyle.Render("vstr-bridge"))
	message.WriteString("\n\n")

	return message.String()
}

// completeSetup marks the setup as complete and saves the configuration.
func completeSetup() error {
	config := models.Config{
		IsSetupComplete: true,
	}

	file, err := os.Create(ConfigurationFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	styles.PrintSuccess("Setup completed successfully!")
	styles.PrintInfo("You can now use 'vstr' to manage your development workflow.")

	return nil
}

// IsExtensionInstalled checks if the VSCode extension is already installed.
func IsExtensionInstalled() bool {
	cmd := exec.Command("code", "--list-extensions")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	installedExtensions := string(output)
	return strings.Contains(strings.ToLower(installedExtensions), "diegopherlt.vstr-bridge")
}

// installExtension downloads the latest VSIX from GitHub Releases and installs it.
func installExtension() error {
	styles.PrintProgress("Fetching latest VSTR-Bridge release...")

	vsixPath, err := downloadLatestExtension()
	if err != nil {
		return fmt.Errorf("downloading extension: %w", err)
	}
	defer os.Remove(vsixPath)

	styles.PrintProgress("Installing VSTR-Bridge extension...")

	cmd := exec.Command("code", "--install-extension", vsixPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w\nOutput: %s", err, string(output))
	}

	styles.PrintSuccess("Extension installed successfully!")
	styles.PrintInfo("Please restart VSCode to activate the extension.")
	return nil
}

type githubRelease struct {
	Assets []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func downloadLatestExtension() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/DieGopherLT/VSTR-Bridge/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "vstr-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decoding release: %w", err)
	}

	var asset *releaseAsset
	for i := range release.Assets {
		if strings.HasSuffix(release.Assets[i].Name, ".vsix") {
			asset = &release.Assets[i]
			break
		}
	}
	if asset == nil {
		return "", fmt.Errorf("no .vsix asset found in latest release")
	}

	downloadResp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", asset.Name, err)
	}
	defer downloadResp.Body.Close()

	tmpFile, err := os.CreateTemp("", "vstr-bridge-*.vsix")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, downloadResp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("writing extension file: %w", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}

// getInstallationChoice prompts the user for installation choice with better UX.
func getInstallationChoice() string {
	promptStyle := lipgloss.NewStyle().
		Foreground(styles.VSCodeBlue).
		Bold(true)

	optionsStyle := lipgloss.NewStyle().
		Foreground(styles.LightGray)

	fmt.Print(promptStyle.Render("Would you like to install the extension now?"))
	fmt.Print(" ")
	fmt.Print(optionsStyle.Render("[Y/n]: "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "n"
	}

	return strings.ToLower(strings.TrimSpace(input))
}
