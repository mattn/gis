package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"code.google.com/p/goauth2/oauth"

	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
)

func main() {
	user, err := getUser()
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Name = "gis"
	app.Usage = "show git issue list"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Value: user,
			Usage: "target user",
		},
		cli.BoolFlag{
			Name:  "assignee, a",
			Usage: "filter issues based on their assignee",
		},
		cli.BoolFlag{
			Name:  "mentioned, m",
			Usage: "filter issues to those mentioned target user",
		},
		cli.IntFlag{
			Name:  "page, p",
			Value: 1,
			Usage: "page",
		},
	}
	app.Action = func(c *cli.Context) {
		token, err := getToken()
		if err != nil {
			log.Fatal(err)
		}
		owner, repo, err := getOwnerAndRepo()
		if err != nil {
			log.Fatal(err)
		}

		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: token},
		}
		client := github.NewClient(t.Client())

		opt := &github.IssueListByRepoOptions{
			ListOptions: github.ListOptions{
				Page: c.Int("page"),
			},
		}
		if c.Bool("assignee") {
			opt.Assignee = user
		}
		if c.Bool("mentioned") {
			opt.Mentioned = user
		}

		issues, _, err := client.Issues.ListByRepo(owner, repo, opt)
		if err != nil {
			log.Fatal(err)
		}

		for _, issue := range issues {
			fmt.Printf("%d: %s\n", *issue.Number, *issue.Title)
		}
	}
	app.Run(os.Args)
}

func getUser() (string, error) {
	return getGitConfigValue("user.name")
}

func getToken() (string, error) {
	return getGitConfigValue("gis.token")
}

func getOwnerAndRepo() (string, string, error) {
	url, err := getGitConfigValue("remote.origin.url")
	if err != nil {
		return "", "", err
	}

	re, err := regexp.Compile(`git@github\.com:(.+)/(.+)\.git`)
	if err != nil {
		return "", "", err
	}

	matches := re.FindStringSubmatch(url)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("can't parse remote.origin.url")
	}

	return matches[1], matches[2], nil
}

func getGitConfigValue(key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}