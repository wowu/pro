package gitlab

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrProjectNotFound = errors.New("project not found")
var ErrMergeRequestNotFound = errors.New("merge request not found")
var ErrTokenExpired = errors.New("token expired")

type ApiResponse struct {
	StatusCode int
	Body       []byte
}

func apiGet(url string, token string) (ApiResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ApiResponse{}, err
	}

	req.Header.Set("PRIVATE-TOKEN", token)

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
	url := "https://gitlab.com/api/v4/user"
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
		return UserResponse{}, errors.New("unknown response code")
	}
}

type MergeRequestResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	WebUrl       string `json:"web_url"`
}

func FindMergeRequest(projectPath string, token string, branch string) (MergeRequestResponse, error) {
	url := "https://gitlab.com/api/v4/projects/" + url.QueryEscape(projectPath) + "/merge_requests?state=opened&source_branch=" + url.QueryEscape(branch)
	resp, err := apiGet(url, token)
	if err != nil {
		return MergeRequestResponse{}, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		var body map[string]interface{}
		err = json.Unmarshal(resp.Body, &body)
		if err != nil {
			return MergeRequestResponse{}, err
		}

		if body["error_description"] == "Token is expired. You can either do re-authorization or token refresh." {
			return MergeRequestResponse{}, ErrTokenExpired
		} else {
			return MergeRequestResponse{}, ErrUnauthorized
		}
	case http.StatusNotFound:
		return MergeRequestResponse{}, ErrProjectNotFound
	case http.StatusOK:
		var mergeRequests []MergeRequestResponse
		err = json.Unmarshal(resp.Body, &mergeRequests)
		if err != nil {
			return MergeRequestResponse{}, err
		}

		if len(mergeRequests) == 0 {
			return MergeRequestResponse{}, ErrMergeRequestNotFound
		}

		return mergeRequests[0], nil
	default:
		return MergeRequestResponse{}, errors.New("unknown response code")
	}
}

func GetRemoteBranches(projectPath string, token string) ([]string, error) {
  url := "https://gitlab.com/api/v4/projects/" + url.QueryEscape(projectPath) + "/repository/branches"
  resp, err := apiGet(url, token)
  if err != nil {
    return nil, err
  }

  switch resp.StatusCode {
  case http.StatusUnauthorized:
    return nil, ErrUnauthorized
  case http.StatusOK:
    var branches []struct {
      Name string `json:"name"`
    }
    err = json.Unmarshal(resp.Body, &branches)
    if err != nil {
      return nil, err
    }

    var branchNames []string
    for _, b := range branches {
      branchNames = append(branchNames, b.Name)
    }

    return branchNames, nil
  default:
    return nil, errors.New("unknown response code")
  }
}
