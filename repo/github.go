package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ghReadmeURL = `https://api.%s/repos/%s/%s/readme`
	ghQuery     = `query {
	repository(name: "%s", owner: "%s") {
		defaultBranchRef {
			name
			target {
				... on Commit {
					oid
					pushedDate
				}
			}
		}
	}
}
`
	ghRateLimitQuery = `query { rateLimit { remaining resetAt } }`
)

type ErrGitHubGQL struct {
	Msg string
}

func (e *ErrGitHubGQL) Error() string {
	return fmt.Sprintf("github query error: %s", e.Msg)
}

type ErrGitHubGQLRateLimit struct {
	ResetAt time.Time
}

func (e *ErrGitHubGQLRateLimit) Error() string {
	return "github graphql rate limit reached"
}

type GitHub struct {
	ProjectName  string
	ProjectOwner string
	Host         string
	AuthToken    string
}

// RateLimit response
type ghRateLimit struct {
	RateLimit struct {
		Remaining int       `json:"remaining"`
		ResetAt   time.Time `json:"resetAt"`
	} `json:"rateLimit"`
}

// GraphQL response for the repo info
type ghResp struct {
	Repository struct {
		ID               string `json:"id"`
		DefaultBranchRef struct {
			Name   string `json:"name"`
			Target struct {
				Oid        string    `json:"oid"`
				PushedDate time.Time `json:"pushedDate"`
			} `json:"target"`
		} `json:"defaultBranchRef"`
	} `json:"repository"`
}

func (r *GitHub) Readme() (string, error) {
	var res ghResp

	url := fmt.Sprintf("https://api.%s/graphql", r.Host)
	q := GraphQLQuery{
		Query: fmt.Sprintf(ghQuery, r.ProjectName, r.ProjectOwner),
	}

	if err := QueryGraphQL(url, r.AuthToken, q, &res); err != nil {
		return "", err
	}

	// rootRef := res.Repository.DefaultBranchRef.Name
	// lastCommitHash := res.Repository.DefaultBranchRef.Target.Oid
	// authoredDate := res.Repository.DefaultBranchRef.Target.PushedDate

	readmeAPIResp, err := http.Get(fmt.Sprintf(ghReadmeURL, r.Host, r.ProjectOwner, r.ProjectName))
	if err != nil {
		return "", err
	}
	defer readmeAPIResp.Body.Close()

	m := make(map[string]interface{})
	if err = json.NewDecoder(readmeAPIResp.Body).Decode(&m); err != nil {
		return "", err
	}

	if s, ok := m["message"]; ok && s == "Not Found" {
		return "", ErrNoFile
	}

	if m["encoding"] == "base64" {
		rs := m["content"].(string)
		b, err := base64.StdEncoding.DecodeString(rs)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	readmeResp, err := http.Get(m["download_url"].(string))
	if err != nil {
		return "", err
	}
	defer readmeResp.Body.Close()
	b, err := ioutil.ReadAll(readmeResp.Body)

	// TODO: implement size checks and timeouts, don't want to deal with unresponsive remotes
	return string(b), err
}

func (r *GitHub) RateLimit() (remaining int, resetAt time.Time, err error) {
	var res ghRateLimit

	url := fmt.Sprintf("https://api.%s/graphql", r.Host)

	q := GraphQLQuery{
		Query: ghRateLimitQuery,
	}

	if err = QueryGraphQL(url, r.AuthToken, q, &res); err != nil {
		return
	}

	remaining = res.RateLimit.Remaining
	resetAt = res.RateLimit.ResetAt

	return
}
