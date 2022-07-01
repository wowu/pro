package gitlab

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var ErrorUnauthorized = errors.New("Unauthorized")
var ErrorNotFound = errors.New("Not found")

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

	body, err := ioutil.ReadAll(resp.Body)
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
		return UserResponse{}, ErrorUnauthorized
	case http.StatusOK:
		var user UserResponse
		err = json.Unmarshal(resp.Body, &user)
		if err != nil {
			return UserResponse{}, err
		}

		return user, nil
	default:
		return UserResponse{}, errors.New("Unknown response code")
	}
}

type MergeRequestResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	WebUrl       string `json:"web_url"`
}

func MergeRequests(projectPath string, token string) ([]MergeRequestResponse, error) {
	url := "https://gitlab.com/api/v4/projects/" + url.QueryEscape(projectPath) + "/merge_requests"
	resp, err := apiGet(url, token)
	if err != nil {
		return []MergeRequestResponse{}, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return []MergeRequestResponse{}, ErrorUnauthorized
	case http.StatusNotFound:
		return []MergeRequestResponse{}, ErrorNotFound
	case http.StatusOK:
		var mergeRequests []MergeRequestResponse
		err = json.Unmarshal(resp.Body, &mergeRequests)
		if err != nil {
			return []MergeRequestResponse{}, err
		}

		return mergeRequests, nil
	default:
		return []MergeRequestResponse{}, errors.New("Unknown response code")
	}
}
