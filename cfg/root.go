package cfg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
		}
	}

	data, err := ioutil.ReadFile(configfile())
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	return config
}

func Save(config Config) {
	// Make sure the config directory exists
	err := os.MkdirAll(path.Dir(configfile()), 0750)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := yaml.Marshal(config)

	ioutil.WriteFile(configfile(), data, 0600)
}

func configdir() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, ".config")
}

func configfile() string {
	return filepath.Join(configdir(), "pro", "config.yml")
}
