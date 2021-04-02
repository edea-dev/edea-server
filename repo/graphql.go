package repo

// SPDX-License-Identifier: EUPL-1.2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// GraphQLQuery helper struct
type GraphQLQuery struct {
	Query         string      `json:"query"`
	OperationName string      `json:"operationName,omitempty"`
	Variables     interface{} `json:"variables,omitempty"`
}

// GraphQLResponse struct
type GraphQLResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Message   string `json:"message,omitempty"`
		Locations []struct {
			Line   int `json:"line,omitempty"`
			Column int `json:"column,omitempty"`
		} `json:"locations,omitempty"`
	} `json:"errors,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (e *GraphQLResponse) Error() string {
	var sb strings.Builder
	sb.WriteString("graphql query error: ")
	for i, msg := range e.Errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(msg.Message)
		sb.WriteString(" at ")
		for j, l := range msg.Locations {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%d:%d", l.Line, l.Column))
		}
	}
	return sb.String()
}

// QueryGraphQL runs a GraphQL with the given parameters
func QueryGraphQL(url string, token string, q GraphQLQuery, res interface{}) error {
	b, err := json.Marshal(q)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	r := &GraphQLResponse{Data: res}

	b, _ = ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

	if err = json.NewDecoder(bytes.NewReader(b)).Decode(r); err != nil {
		return err
	}

	if len(r.Errors) > 0 || len(r.Message) > 0 {
		r.Data = nil
		return r
	}

	return nil
}
