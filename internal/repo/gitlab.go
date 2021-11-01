package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/*
 * Get repo file:
 * "https://gitlab.example.com/api/v4/projects/13083/repository/files/app%2Fmodels%2Fkey%2Erb/raw?ref=master"
 *
 * Get default branch:
 * "https://gitlab.example.com/api/v4/projects/5/repository/branches"
 *
 * should also work with the repo path instead of id
 */

type GitLab struct {
	FullPath string
	Host     string
}

type glResp struct {
	Project struct {
		ID          string `json:"id"`
		Respository struct {
			RootRef string `json:"rootRef"`
			Tree    struct {
				LastCommit struct {
					AuthoredDate time.Time `json:"authoredDate"`
					SHA          string    `json:"sha"`
				} `json:"lastCommit"`
			} `json:"tree"`
		} `json:"repository"`
	} `json:"project"`
}

var gitlabQuery = `{
	project(fullPath: "%s") {
		id
		repository {
			rootRef
			tree {
				lastCommit {
					authoredDate
					sha
				}
			}
		}
	}
}
`

func (r *GitLab) Readme() (string, error) {
	var res glResp

	url := fmt.Sprintf("https://%s/api/graphql", r.Host)
	q := GraphQLQuery{
		Query: fmt.Sprintf(gitlabQuery, r.FullPath),
	}
	if err := QueryGraphQL(url, "", q, &res); err != nil {
		return "", err
	}

	fmt.Println(res.Project.ID)
	if res.Project.ID == "" {
		return "", fmt.Errorf("no project id")
	}

	/*
		id, err := strconv.Atoi(res.Project.ID[21:]) // gid://gitlab/Project/{id}
		if err != nil {
			return "", err
		}

		rootRef := res.Project.Respository.RootRef
		lastCommitHash := res.Project.Respository.Tree.LastCommit.SHA
		authoredData := res.Project.Respository.Tree.LastCommit.AuthoredDate
	*/

	readmeURL := fmt.Sprintf("https://%s/%s/-/raw/%s/README.md", r.Host, r.FullPath, res.Project.Respository.RootRef)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	readmeResp, err := client.Get(readmeURL)
	if err != nil {
		return "", err
	}
	defer readmeResp.Body.Close()

	if readmeResp.StatusCode == http.StatusSeeOther ||
		readmeResp.StatusCode == http.StatusPermanentRedirect {
		return "", ErrNoFile // redirect of an empty repository
	}

	if readmeResp.StatusCode != http.StatusOK {
		fmt.Println(readmeResp.StatusCode)
		return "", ErrUnexpectedResponse
	}

	b, err := ioutil.ReadAll(readmeResp.Body)

	if len(b) == 0 {
		return "", ErrNoFile
	}

	// TODO: implement size checks and timeouts, don't want to deal with unresponsive remotes

	return string(b), nil
}
