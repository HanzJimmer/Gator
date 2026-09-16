package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	db_url            string `json:"db_url"`
	current_user_name string `json:"current_user_name"`
}

func Read() (Config, error) {
	homepath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	file, err := os.ReadFile(homepath + configFileName)
	if err != nil {
		return Config{}, err
	}

	var configInfo Config
	if err := json.Unmarshal(file, &configInfo); err != nil {
		return Config{}, err
	}
	return configInfo, nil
}

func (c Config) SetUser() {
	user, err := os.Hostname()
	if err != nil {
		return
	}
	c.current_user_name = user
	write(c)
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homedir, nil
}

func write(cfg Config) error {
	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	writePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	os.WriteFile(writePath+configFileName, jsonConfig, 0777)
	return nil
}
