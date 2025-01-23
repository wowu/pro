package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrNotFound = errors.New("not found")

type ApiResponse struct {
	StatusCode int
	Body       []byte
}

func apiGet(url string, token string) (ApiResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ApiResponse{}, err
	}

	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ApiResponse{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ApiResponse{}, err
	}

	return ApiResponse{resp.StatusCode, body}, nil
}

type UserResponse struct {
	ID int `json:"id"`
}

func User(token string) (UserResponse, error) {
	url := "https://api.github.com/user"
	resp, err := apiGet(url, token)
	if err != nil {
		return UserResponse{}, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return UserResponse{}, ErrUnauthorized
	case http.StatusOK:
		var user UserResponse
		err = json.Unmarshal(resp.Body, &user)
		if err != nil {
			return UserResponse{}, err
		}

		return user, nil
	default:
		return UserResponse{}, errors.New("unknown response code: " + fmt.Sprint(resp.StatusCode) + " " + string(resp.Body))
	}
}

type PullRequestResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
	Head  struct {
		Ref string `json:"ref"`
	} `json:"head"`
	HtmlURL string `json:"html_url"`
}

func FindPullRequest(projectPath string, token string, branch string) (PullRequestResponse, error) {
	userOrOrg := strings.Split(projectPath, "/")[0]
	url := "https://api.github.com/repos/" + projectPath + "/pulls?state=open&head=" + userOrOrg + ":" + url.QueryEscape(branch)

	resp, err := apiGet(url, token)
	if err != nil {
		return PullRequestResponse{}, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return PullRequestResponse{}, ErrUnauthorized
	case http.StatusOK:
		var pullRequests []PullRequestResponse
		err = json.Unmarshal(resp.Body, &pullRequests)
		if err != nil {
			return PullRequestResponse{}, err
		}

		if len(pullRequests) == 0 {
			return PullRequestResponse{}, ErrNotFound
		}

		return pullRequests[0], nil
	default:
		return PullRequestResponse{}, errors.New("unknown response code: " + fmt.Sprint(resp.StatusCode) + " " + string(resp.Body))
	}
}

func GetRemoteBranches(projectPath string, token string) ([]string, error) {
  url := "https://api.github.com/repos/" + projectPath + "/git/refs/heads"

  resp, err := apiGet(url, token)
  if err != nil {
    return nil, err
  }

  switch resp.StatusCode {
  case http.StatusUnauthorized:
    return nil, ErrUnauthorized
  case http.StatusOK:
    var branches []struct {
      Ref string `json:"ref"`
    }
    err = json.Unmarshal(resp.Body, &branches)
    if err != nil {
      return nil, err
    }

    var branchNames []string
    for _, branch := range branches {
      branchNames = append(branchNames, strings.TrimPrefix(branch.Ref, "refs/heads/"))
    }

    return branchNames, nil
  default:
    return nil, errors.New("unknown response code: " + fmt.Sprint(resp.StatusCode) + " " + string(resp.Body))
  }
}
