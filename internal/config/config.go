package config

import (
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Url  string `json:"db_url"`
	User string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	root, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, configFileName), nil

}
