package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"wowu/pro/cfg"
	"wowu/pro/gitlab"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	giturls "github.com/whilp/git-urls"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
)

func auth(provider string) {
	fmt.Println("Generate your token at https://gitlab.com/-/profile/personal_access_tokens?name=PR+opener&scopes=api")
	fmt.Println("The only required scope is 'api'")

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

func open(repoPath string, print bool) {
	r, err := git.PlainOpen(repoPath)
	handleError(err)

	remotes, err := r.Remotes()
	handleError(err)
	if len(remotes) == 0 {
		fmt.Println("No remotes found")
		os.Exit(0)
	}

	// check if there is a remote named origin
	origin, err := r.Remote("origin")
	handleError(err)

	// get current head
	head, err := r.Head()
	handleError(err)

	if !head.Name().IsBranch() {
		fmt.Println("Not on a branch")
		os.Exit(0)
	}

	// get current branch name
	branch := head.Name().Short()
	fmt.Printf("Current branch: %s\n", color.GreenString(branch))

	originURL := origin.Config().URLs[0]

	gitURL, err := giturls.Parse(originURL)
	handleError(err)

	if branch == "master" || branch == "main" || branch == "trunk" {
		fmt.Println("Looks like you are on the main branch. Opening home page.")

		homeUrl := fmt.Sprintf("https://%s/%s", gitURL.Host, gitURL.Path)

		if print {
			fmt.Println(homeUrl)
		} else {
			fmt.Println(homeUrl)
			openBrowser(homeUrl)
		}

		os.Exit(0)
	}

	type RemoteType int
	const (
		Gitlab RemoteType = 0
		Github            = 1
	)

	var remoteType RemoteType
	switch gitURL.Host {
	case "gitlab.com":
		remoteType = Gitlab
	case "github.com":
		remoteType = Github
	default:
		fmt.Println("Unknown remote type")
		os.Exit(1)
	}

	if remoteType == Github {
		fmt.Println("Github is not supported yet")
		os.Exit(0)
	}

	path := gitURL.Path

	// remove trailing ".git" from path if it exists
	if path[len(path)-4:] == ".git" {
		path = path[:len(path)-4]
	}

	gitlabToken := cfg.Get().GitlabToken

	if gitlabToken == "" {
		color.Red("Gitlab token is not set. Run `pro auth gitlab` to set it.")
		os.Exit(1)
	}

	mergeRequests, err := gitlab.MergeRequests(path, gitlabToken)
	handleError(err)

	// find merge request for current branch
	currentMergeRequestIndex := slices.IndexFunc(mergeRequests, func(mr gitlab.MergeRequestResponse) bool {
		return mr.SourceBranch == branch && mr.State == "opened"
	})

	if currentMergeRequestIndex == -1 {
		fmt.Println("No open merge request found for current branch")

		fmt.Printf("Create pull request at https://gitlab.com/%s/merge_requests/new?merge_request%%5Bsource_branch%%5D=%s\n", path, branch)

		os.Exit(0)
	}

	currentMergeRequest := mergeRequests[currentMergeRequestIndex]

	url := currentMergeRequest.WebUrl

	if print {
		fmt.Println(url)
	} else {
		fmt.Println("Opening " + url)
		openBrowser(url)
	}
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Println("Error opening browser:", err)
		os.Exit(1)
	}
}

// Print error and exit
func handleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
