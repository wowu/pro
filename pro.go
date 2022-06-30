package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	giturls "github.com/whilp/git-urls"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
)

func main() {
	app := &cli.App{
		Name:  "pro",
		Usage: "Pull Request opener",
		Commands: []*cli.Command{
			{
				Name:      "auth",
				ArgsUsage: "[gitlab|github]",
				Usage:     "Login to GitLab or GitHub",
				UsageText: fmt.Sprintf("pro auth gitlab\npro login github"),
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						fmt.Println("Please specify provider (github or gitlab)")
						os.Exit(1)
					}

					provider := c.Args().Get(0)

					if provider != "gitlab" && provider != "github" {
						fmt.Println("Please specify provider (github or gitlab)")
						os.Exit(1)
					}

					if provider == "github" {
						fmt.Println("Github provider is not supported yet")
						os.Exit(0)
					}

					fmt.Println("Generate your token at https://gitlab.com/-/profile/personal_access_tokens?name=PR+opener&scopes=api")
					fmt.Println("The only required scope is 'api'")

					// ask for token
					fmt.Print("Token: ")
					byteToken, err := term.ReadPassword(int(syscall.Stdin))
					fmt.Println()
					if err != nil {
						return err
					}

					token := strings.TrimSpace(string(byteToken))

					if token == "" {
						color.Red("Token is empty. Try again")
						os.Exit(1)
					}

					color.Green("Saved.")

					return nil
				},
			},
		},
		Action: func(*cli.Context) error {
			cli.HandleExitCoder(errors.New("not an exit coder, though"))

			r, err := git.PlainOpen(".")
			if err != nil {
				fmt.Println(err)
				return err
			}

			remotes, err := r.Remotes()
			if err != nil {
				fmt.Println(err)
				return err
			}
			if len(remotes) == 0 {
				fmt.Println("No remotes found")
				return nil
			}

			// check if there is a remote named origin
			origin, err := r.Remote("origin")
			if err != nil {
				fmt.Println(err)
				return err
			}

			// get current head
			head, err := r.Head()
			if err != nil {
				fmt.Println(err)
				return err
			}

			if !head.Name().IsBranch() {
				fmt.Println("Not on a branch")
				return nil
			}

			// get the branch name
			branch := head.Name().Short()

			fmt.Printf("Current branch: %s\n", branch)

			originURL := origin.Config().URLs[0]

			gitURL, err := giturls.Parse(originURL)
			if err != nil {
				fmt.Println(err)
				return err
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
				return nil
			}

			if remoteType == Github {
				fmt.Println("Github is not supported yet")
				return nil
			}

			path := gitURL.Path

			// remove trailing ".git" from path if it exists
			if path[len(path)-4:] == ".git" {
				path = path[:len(path)-4]
			}

			req, err := http.NewRequest("GET", fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/merge_requests", url.QueryEscape(path)), nil)
			if err != nil {
				fmt.Println(err)
				return err
			}
			req.Header.Set("PRIVATE-TOKEN", os.Getenv("PR_TOKEN"))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return err
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			if resp.StatusCode != 200 {
				fmt.Println(string(body))
				return nil
			}

			type MergeRequest struct {
				ID           int    `json:"id"`
				Title        string `json:"title"`
				State        string `json:"state"`
				SourceBranch string `json:"source_branch"`
				WebUrl       string `json:"web_url"`
			}

			var mergeRequests []MergeRequest
			json.Unmarshal(body, &mergeRequests)

			// find merge request for current branch
			currentMergeRequestIndex := slices.IndexFunc(mergeRequests, func(mr MergeRequest) bool {
				return mr.SourceBranch == branch && mr.State == "opened"
			})

			if currentMergeRequestIndex == -1 {
				fmt.Println("No open merge request found for current branch")

				fmt.Printf("Create pull request at https://gitlab.com/%s/merge_requests/new?merge_request%%5Bsource_branch%%5D=%s", path, branch)

				return nil
			}

			currentMergeRequest := mergeRequests[currentMergeRequestIndex]

			url := currentMergeRequest.WebUrl

			fmt.Println("Opening " + url)

			openbrowser(url)

			return nil
		},
	}

	app.Run(os.Args)
}

func openbrowser(url string) {
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
		log.Fatal(err)
	}
}
