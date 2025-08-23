package cfg

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/DieGopherLT/vscode-terminal-runner/internal/models"
)

var (
	ConfigurationFile string
)

func init() {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic("could not determine user config directory: " + err.Error())
	}

	ConfigurationFile = path.Join(cfgDir, "vscode-terminal-runner", "config.json")

	if _, err := os.Stat(ConfigurationFile); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path.Dir(ConfigurationFile), 0755); err != nil {
			panic("could not create config directory: " + err.Error())
		}
		if _, err := os.Create(ConfigurationFile); err != nil {
			panic("could not create config file: " + err.Error())
		}
	}
}

func Load() (models.Config, error) {
	file, err := os.Open(ConfigurationFile)
	if err != nil {
		return models.Config{}, err
	}
	defer file.Close()

	var config models.Config

	// Get file information to check if file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return models.Config{}, err
	}

	// If file is empty, return default config
	if fileInfo.Size() == 0 {
		return models.Config{}, nil
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return models.Config{}, err
	}

	return config, nil
}
