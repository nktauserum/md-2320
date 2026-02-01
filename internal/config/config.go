package config

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed config.yml
var yml_config []byte

type Config struct {
	AUTHORIZED_USERS         []int64 `yaml:"authorized_users"`
	TELEGRAM_TOKEN           string  `yaml:"telegram_token"`
	QBITTORRENT_API_URL      string  `yaml:"qbittorrent_api_url"`
	QBITTORRENT_API_USERNAME string  `yaml:"qbittorrent_api_username"`
	QBITTORRENT_API_PASSWORD string  `yaml:"qbittorrent_api_password"`
	DOWNLOAD_FOLDER          string  `yaml:"download_folder"`
}

func LoadConfig() (*Config, error) {
	var config *Config
	err := yaml.Unmarshal(yml_config, &config)
	return config, err
}
