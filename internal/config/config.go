package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Url  string `json:"db_url"`
	User string `json:"current_user_name"`
}

func (c *Config) SetUser(name string) error {
	c.User = name
	err := Write(*c)
	if err != nil {
		return err
	}
	return nil
}

func Write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
	/* the third argument is the **file permission**. `0600` means only the owner
	 * can read/write it, which is sensible for a config file containing credentials.*/
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func getConfigFilePath() (string, error) {
	root, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, configFileName), nil

}
