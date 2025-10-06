package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	giturls "github.com/whilp/git-urls"
	"github.com/wowu/pro/config"
	"github.com/wowu/pro/provider/github"
	"github.com/wowu/pro/provider/gitlab"
	"github.com/wowu/pro/repository"
)

func List(repoPath string, print bool, copy bool) {
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

	projectPath := strings.TrimPrefix(gitURL.Path, "/")
	projectPath = strings.TrimSuffix(projectPath, ".git")

	var prTitles []string
	var prUrls []string

	// Append repository homepage
	prTitles = append(prTitles, fmt.Sprintf("Repository homepage (%s)", projectPath))
	prUrls = append(prUrls, fmt.Sprintf("https://%s/%s", gitURL.Host, projectPath))

	switch gitURL.Host {
	case "github.com":
		prs := getGitHubOpenPullRequests(projectPath)
		for _, pr := range prs {
			prTitles = append(prTitles, fmt.Sprintf("%s (#%d)", pr.Title, pr.Number))
			prUrls = append(prUrls, pr.HtmlURL)
		}
	case "gitlab.com":
		mrs := getGitLabOpenMergeRequests(projectPath)
		for _, mr := range mrs {
			prTitles = append(prTitles, fmt.Sprintf("%s (!%d)", mr.Title, mr.IID))
			prUrls = append(prUrls, mr.WebUrl)
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown remote type")
		os.Exit(1)
	}

	if len(prTitles) == 0 {
		fmt.Println("No open pull/merge requests found.")
		return
	}

	idx, err := fuzzyfinder.Find(
		prTitles,
		func(i int) string {
			return prTitles[i]
		},
	)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return
		}
		handleError(err, "Fuzzyfinder failed")
	}

	if print {
		color.Blue(prUrls[idx])
	} else if copy {
		copyToClipboard(prUrls[idx])
		fmt.Fprintln(os.Stderr, "Copied to clipboard: "+color.BlueString(prUrls[idx]))
	} else {
		fmt.Fprintln(os.Stderr, "Opening "+color.BlueString(prUrls[idx]))
		openBrowser(prUrls[idx])
	}
}

func getGitHubOpenPullRequests(projectPath string) []github.PullRequestResponse {
	githubToken := config.Get().GitHubToken
	if githubToken == "" {
		color.Red("GitHub token is not set. Run `pro auth github` to set it.")
		os.Exit(1)
	}
	prs, err := github.ListOpenPullRequests(projectPath, githubToken)
	if err != nil {
		if errors.Is(err, github.ErrUnauthorized) {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get pull requests: %s", err.Error()))
			fmt.Fprintln(os.Stderr, "Token may be expired or deleted. Run `pro auth github` to connect GitHub again.")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get pull requests: %s", err.Error()))
			os.Exit(1)
		}
	}
	return prs
}

func getGitLabOpenMergeRequests(projectPath string) []gitlab.MergeRequestResponse {
	gitlabToken := config.Get().GitLabToken
	if gitlabToken == "" {
		color.Red("GitLab token is not set. Run `pro auth gitlab` to set it.")
		os.Exit(1)
	}
	mrs, err := gitlab.ListOpenMergeRequests(projectPath, gitlabToken)
	if err != nil {
		if errors.Is(err, gitlab.ErrUnauthorized) || errors.Is(err, gitlab.ErrTokenExpired) {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get merge requests: %s", err.Error()))
			fmt.Fprintln(os.Stderr, "Connect GitLab again with `pro auth gitlab`.")
			os.Exit(1)
		} else if errors.Is(err, gitlab.ErrProjectNotFound) {
			fmt.Fprintln(os.Stderr, color.RedString("Project not found: %s", projectPath))
			fmt.Fprintln(os.Stderr, "Maybe it was renamed or deleted? Change remote URL and try again.")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.RedString("Unable to get merge requests: %s", err.Error()))
			os.Exit(1)
		}
	}
	return mrs
}
