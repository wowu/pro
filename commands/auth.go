package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"wowu/pro/cfg"
	"wowu/pro/github"
	"wowu/pro/gitlab"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func Auth(provider string) {
	switch provider {
	case "gitlab":
		authgitlab()
	case "github":
		authgithub()
	default:
		fmt.Println("unknown provider")
		os.Exit(1)
	}
}

func authgitlab() {
	fmt.Println("Generate your token at https://gitlab.com/-/profile/personal_access_tokens?name=PR+opener&scopes=api")
	fmt.Println("The only required scope is 'api'")

	// Ask for token
	fmt.Print("Token: ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	handleError(err)

	token := strings.TrimSpace(string(byteToken))

	if token == "" {
		color.Red("Token is empty. Try again")
		os.Exit(1)
	}

	// Check if token is valid by fetching user info
	_, err = gitlab.User(token)
	if err != nil {
		switch err {
		case gitlab.ErrorUnauthorized:
			color.Red("Token is invalid. Try again")
			os.Exit(1)
		default:
			fmt.Println(err)
			os.Exit(1)
		}
	}

	config := cfg.Get()
	config.GitlabToken = token
	cfg.Save(config)

	color.Green("Saved.")
}

func authgithub() {
	fmt.Println("Generate personal access token at https://github.com/settings/tokens/new")
	fmt.Println("The only required scope is 'repo'")

	// ask for token
	fmt.Print("Token: ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	handleError(err)

	token := strings.TrimSpace(string(byteToken))

	if token == "" {
		color.Red("Token is empty. Try again")
		os.Exit(1)
	}

	// Check if token is valid by fetching user info
	_, err = github.User(token)
	if err != nil {
		switch err {
		case github.ErrorUnauthorized:
			color.Red("Token is invalid. Try again")
			os.Exit(1)
		default:
			fmt.Println(err)
			os.Exit(1)
		}
	}

	config := cfg.Get()
	config.GitHubToken = token
	cfg.Save(config)

	color.Green("Saved.")
}
