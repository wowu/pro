package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	giturls "github.com/whilp/git-urls"
	"github.com/wowu/pro/config"
	"github.com/wowu/pro/provider/github"
	"github.com/wowu/pro/provider/gitlab"
	"github.com/wowu/pro/repository"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
)

func Open(repoPath string, print bool, copy bool) {
	repo, err := repository.FindInParents(repoPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Unable to find git repository in given directory or any of parent directories."))
		fmt.Fprintln(os.Stderr, "Please make sure you are in the project directory.")
		os.Exit(1)
	}

	originURL, err := repo.OriginUrl()
	if err != nil {
		if errors.Is(err, repository.ErrNoRemoteOrigin) {
			fmt.Fprintln(os.Stderr, color.RedString("No remote named \"origin\" found."))
			fmt.Fprintln(os.Stderr, "Please make sure you have a remote named \"origin\".")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get origin URL: %s", err.Error()))
		}
	}

	gitURL, err := giturls.Parse(originURL)
	handleError(err, "Unable to parse origin URL")

	branch, err := repo.CurrentBranchName()
	if err != nil {
		if errors.Is(err, repository.ErrNoActiveBranch) {
			fmt.Fprintln(os.Stderr, color.RedString("No active branch found."))
			fmt.Fprintln(os.Stderr, "Switch to a branch and try again.")
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get current branch: %s", err.Error()))
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "Current branch: %s\n", color.GreenString(branch))

	if branch == "master" || branch == "main" || branch == "trunk" || branch == "develop" || branch == "dev" {
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
		exists, url = getGitLabUrl(branch, projectPath)
		requestType = "merge request"
	case "github.com":
		exists, url = getGitHubUrl(branch, projectPath)
		requestType = "pull request"
	default:
		fmt.Fprintln(os.Stderr, "Unknown remote type")
		os.Exit(1)
	}

	if !exists {
		fmt.Fprintf(os.Stderr, "No open %s found for current branch. Opening create page.\n", requestType)
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

// Returns merge request URL if it exists for given branch, otherwise returns URL to create new one.
func getGitLabUrl(branch string, projectPath string) (exists bool, url string) {
	gitlabToken := config.Get().GitLabToken

	if gitlabToken == "" {
		color.Red("GitLab token is not set. Run `pro auth gitlab` to set it.")
		os.Exit(1)
	}

	mergeRequest, err := gitlab.FindMergeRequest(projectPath, gitlabToken, branch)
	if err != nil {
		if errors.Is(err, gitlab.ErrMergeRequestNotFound) {
      branches, err := gitlab.GetRemoteBranches(projectPath, gitlabToken)
      if err != nil {
        fmt.Fprintln(os.Stderr, color.RedString("Unable to get branches: %s", err.Error()))
        os.Exit(1)
      }

      // Check if the branch exists in the remote repository
      for _, b := range branches {
        if b == branch {
          return false, fmt.Sprintf("https://gitlab.com/%s/merge_requests/new?merge_request%%5Bsource_branch%%5D=%s", projectPath, branch)
        }
      }

      fmt.Fprintln(os.Stderr, color.RedString("Branch \"%s\" not found in the remote repository. Push the branch to create a merge request.", branch))
      os.Exit(1)
		} else if errors.Is(err, gitlab.ErrProjectNotFound) {
			fmt.Fprintln(os.Stderr, color.RedString("Project \"%s\" not found.", projectPath))
			fmt.Fprintln(os.Stderr, "Maybe it was renamed or deleted? Change remote URL and try again.")
			os.Exit(1)
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
func getGitHubUrl(branch string, projectPath string) (exists bool, url string) {
	githubToken := config.Get().GitHubToken

	if githubToken == "" {
		fmt.Fprintln(os.Stderr, color.RedString("GitHub token is not set. Run `pro auth github` to set it."))
		os.Exit(1)
	}

	pullRequest, err := github.FindPullRequest(projectPath, githubToken, branch)
	if err != nil {
		if errors.Is(err, github.ErrNotFound) {
      branches, err := github.GetRemoteBranches(projectPath, githubToken)
      if err != nil {
        fmt.Fprintln(os.Stderr, color.RedString("Unable to get branches: %s", err.Error()))
        os.Exit(1)
      }

      // Check if the branch exists in the remote repository
      for _, b := range branches {
        if b == branch {
          return false, fmt.Sprintf("https://github.com/%s/pull/new/%s", projectPath, branch)
        }
      }

      fmt.Fprintln(os.Stderr, color.RedString("Branch \"%s\" not found in the remote repository. Push the branch to create a pull request.", branch))
      os.Exit(1)
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
