package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	homepath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.ReadFile(homepath)
	if err != nil {
		return Config{}, err
	}

	var configInfo Config
	if err := json.Unmarshal(file, &configInfo); err != nil {
		return Config{}, err
	}
	return configInfo, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	return write(*c)
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(homedir, configFileName)
	return fullPath, nil
}

func write(cfg *Config) error {
	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	writePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(writePath, jsonConfig, 0644); err != nil {
		return fmt.Errorf("error occured when writing file: %w", err)
	}
	return nil
}
