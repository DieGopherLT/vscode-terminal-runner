package cfg

import (
	"errors"
	"os"
	"path"
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
