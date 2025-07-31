package cfg

import (
	"errors"
	"os"
	"path"
)

var CFG_FOLDER = path.Join(os.Getenv("HOME"), ".config/vsct-runner")
var CFG_FILE = path.Join(CFG_FOLDER, "config.json")

func CheckFile() bool {
	_, err := os.Stat(CFG_FILE)
	return !errors.Is(err, os.ErrNotExist)
}

func CreateFile() error {
	file, err := os.Create(CFG_FILE)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
