package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitHubToken string `yaml:"github_token"`
	GitLabToken string `yaml:"gitlab_token"`
}

// Read config file and return config object
func Get() Config {
	// check if file exists
	_, err := os.Stat(configfile())
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}
		} else {
			fmt.Fprintln(os.Stderr, "Unable to stat config file:", err)
			os.Exit(1)
		}
	}

	data, err := os.ReadFile(configfile())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to read config file:", err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to unmarshal config file:", err)
		os.Exit(1)
	}

	return config
}

func Save(config Config) {
	// Make sure the config directory exists
	configdir, _ := filepath.Split(configfile())
	err := os.MkdirAll(configdir, 0750)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to create config directory:", err)
		os.Exit(1)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to marshal config:", err)
		os.Exit(1)
	}

	err = os.WriteFile(configfile(), data, 0600)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to write config file:", err)
		os.Exit(1)
	}
}

func configdir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get home directory:", err)
		os.Exit(1)
	}

	return filepath.Join(home, ".config")
}

func configfile() string {
	return filepath.Join(configdir(), "pro", "config.yml")
}
