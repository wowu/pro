package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

	req.Header.Set("Authorization", "token "+token)

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
	url := "https://api.github.com/user"
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
		return UserResponse{}, errors.New("Unknown response code: " + fmt.Sprint(resp.StatusCode))
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

func PullRequests(projectPath string, token string) ([]PullRequestResponse, error) {
	url := "https://api.github.com/repos/" + projectPath + "/pulls"
	resp, err := apiGet(url, token)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrorUnauthorized
	case http.StatusOK:
		var pullRequests []PullRequestResponse
		err = json.Unmarshal(resp.Body, &pullRequests)
		if err != nil {
			return nil, err
		}

		return pullRequests, nil
	default:
		return nil, errors.New("Unknown response code: " + fmt.Sprint(resp.StatusCode))
	}
}
