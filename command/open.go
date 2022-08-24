package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wowu/pro/config"
	"github.com/wowu/pro/provider/github"
	"github.com/wowu/pro/provider/gitlab"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	giturls "github.com/whilp/git-urls"
)

func Open(repoPath string, print bool, copy bool) {
	repository, err := findRepo(repoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Unable to find git repository in given directory or any of parent directories."))
		fmt.Fprintln(os.Stderr, "Please make sure you are in the project directory.")
		os.Exit(1)
	}

	// check if there is a remote named origin
	origin, err := repository.Remote("origin")
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("No remote named \"origin\" found."))
		fmt.Fprintln(os.Stderr, "Please make sure you have a remote named \"origin\".")
		os.Exit(1)
	}

	// get current head
	head, err := repository.Head()
	handleError(err, "Unable to get repository head")

	if !head.Name().IsBranch() {
		fmt.Fprintln(os.Stderr, color.RedString("No active branch found."))
		fmt.Fprintln(os.Stderr, "Switch to a branch and try again.")
		os.Exit(0)
	}

	// get current branch name
	branch := head.Name().Short()
	fmt.Fprintf(os.Stderr, "Current branch: %s\n", color.GreenString(branch))

	originURL := origin.Config().URLs[0]

	gitURL, err := giturls.Parse(originURL)
	handleError(err, "Unable to parse origin URL")

	if branch == "master" || branch == "main" || branch == "trunk" || branch == "develop" {
		fmt.Fprintln(os.Stderr, "Looks like you are on the main branch. Opening home page.")

		homeUrl := fmt.Sprintf("https://%s/%s", gitURL.Host, strings.TrimPrefix(gitURL.Path, "/"))
		homeUrl = strings.TrimSuffix(homeUrl, ".git")

		if print {
			color.Blue(homeUrl)
		} else if copy {
			copyToClipboard(homeUrl)
			fmt.Fprintln(os.Stderr, "Copied to clipboard: "+color.BlueString(homeUrl))
		} else {
			color.Blue(homeUrl)
			openBrowser(homeUrl)
		}

		os.Exit(0)
	}

	projectPath := strings.TrimPrefix(gitURL.Path, "/")
	projectPath = strings.TrimSuffix(projectPath, ".git")

	var url string
	var exists bool
	var requestType string
	switch gitURL.Host {
	case "gitlab.com":
		exists, url = getGitLabUrl(branch, projectPath, print)
		requestType = "merge request"
	case "github.com":
		exists, url = getGitHubUrl(branch, projectPath, print)
		requestType = "pull request"
	default:
		fmt.Fprintln(os.Stderr, "Unknown remote type")
		os.Exit(1)
	}

	if !exists {
		fmt.Fprintf(os.Stderr, "No open %s found for current branch\n", requestType)
		fmt.Fprintf(os.Stderr, "Create %s at ", requestType)
		color.Blue(url)

		if copy {
			copyToClipboard(url)
			fmt.Fprintln(os.Stderr, "URL copied to clipboard.")
		}

		os.Exit(0)
	}

	if print {
		color.Blue(url)
	} else if copy {
		copyToClipboard(url)
		fmt.Fprintln(os.Stderr, "Copied to clipboard: "+color.BlueString(url))
	} else {
		fmt.Fprintln(os.Stderr, "Opening "+color.BlueString(url))
		openBrowser(url)
	}
}

// Find git repository in given directory or parent directories.
func findRepo(path string) (*git.Repository, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	repository, err := git.PlainOpen(absolutePath)

	if err == nil {
		return repository, nil
	}

	if errors.Is(err, git.ErrRepositoryNotExists) {
		// Base case - we've reached the root of the filesystem
		if absolutePath == "/" {
			return nil, errors.New("no git repository found")
		}

		// Recurse to parent directory
		return findRepo(filepath.Dir(absolutePath))
	}

	return nil, err
}

// Returns merge request URL if it exists for given branch, otherwise returns URL to create new one.
func getGitLabUrl(branch string, projectPath string, print bool) (exists bool, url string) {
	gitlabToken := config.Get().GitLabToken

	if gitlabToken == "" {
		color.Red("GitLab token is not set. Run `pro auth gitlab` to set it.")
		os.Exit(1)
	}

	mergeRequest, err := gitlab.FindMergeRequest(projectPath, gitlabToken, branch)
	if err != nil {
		if errors.Is(err, gitlab.ErrNotFound) {
			return false, fmt.Sprintf("https://gitlab.com/%s/merge_requests/new?merge_request%%5Bsource_branch%%5D=%s", projectPath, branch)
		} else if errors.Is(err, gitlab.ErrUnauthorized) || errors.Is(err, gitlab.ErrTokenExpired) {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get merge requests: %s", err.Error()))
			fmt.Fprintln(os.Stderr, "Connect GitLab again with `pro auth gitlab`.")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get merge requests: %s", err.Error()))
			os.Exit(1)
		}
	}

	return true, mergeRequest.WebUrl
}

// Returns pull request URL if it exists for given branch, otherwise returns URL to create new one.
func getGitHubUrl(branch string, projectPath string, print bool) (exists bool, url string) {
	githubToken := config.Get().GitHubToken

	if githubToken == "" {
		fmt.Fprintln(os.Stderr, color.RedString("GitHub token is not set. Run `pro auth github` to set it."))
		os.Exit(1)
	}

	pullRequest, err := github.FindPullRequest(projectPath, githubToken, branch)
	if err != nil {
		if errors.Is(err, github.ErrNotFound) {
			return false, fmt.Sprintf("https://github.com/%s/pull/new/%s", projectPath, branch)
		} else if errors.Is(err, github.ErrUnauthorized) {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get pull requests: %s", err.Error()))
			fmt.Fprintln(os.Stderr, "Token may be expired or deleted. Run `pro auth github` to connect GitHub again.")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get pull requests: %s", err.Error()))
			os.Exit(1)
		}
	}

	return true, pullRequest.HtmlURL
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
		fmt.Fprintln(os.Stderr, color.RedString("Unable to open browser: %s", err.Error()))
		os.Exit(1)
	}
}

func copyToClipboard(url string) {
	err := clipboard.WriteAll(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Unable to copy to clipboard: %s", err.Error()))
		os.Exit(1)
	}
}
