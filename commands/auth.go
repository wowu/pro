package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"wowu/pro/config"
	"wowu/pro/providers/github"
	"wowu/pro/providers/gitlab"

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
	fmt.Println("Generate your token at " + color.BlueString("https://gitlab.com/-/profile/personal_access_tokens?name=pro+cli&scopes=read_api"))
	fmt.Println()
	fmt.Println("The only required scope is 'read_api'")
	fmt.Println()

	// Ask for token
	fmt.Print("Token: ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	handleError(err, "Error while reading token")

	token := strings.TrimSpace(string(byteToken))

	if token == "" {
		color.Red("Token is empty. Try again")
		os.Exit(1)
	}

	// Check if token is valid by fetching user info
	_, err = gitlab.User(token)
	if err != nil {
		switch err {
		case gitlab.ErrUnauthorized:
			color.Red("Token is invalid. Try again")
			os.Exit(1)
		default:
			fmt.Println(err)
			os.Exit(1)
		}
	}

	conf := config.Get()
	conf.GitLabToken = token
	config.Save(conf)

	color.Green("Saved.")
}

func authgithub() {
	fmt.Println("Generate personal access token at " + color.BlueString("https://github.com/settings/tokens/new?description=pro+cli&scopes=repo"))
	fmt.Println()
	fmt.Println("The only required scope is 'repo'")
	color.Yellow("It's recommended to set expiration to \"No expiration\"")
	fmt.Println()

	// ask for token
	fmt.Print("Token: ")
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	handleError(err, "Error while reading token")

	token := strings.TrimSpace(string(byteToken))

	if token == "" {
		color.Red("Token is empty. Try again")
		os.Exit(1)
	}

	// Check if token is valid by fetching user info
	_, err = github.User(token)
	if err != nil {
		switch err {
		case github.ErrUnauthorized:
			color.Red("Token is invalid. Try again")
			os.Exit(1)
		default:
			fmt.Println(err)
			os.Exit(1)
		}
	}

	conf := config.Get()
	conf.GitHubToken = token
	config.Save(conf)

	color.Green("Saved.")
}
