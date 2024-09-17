package gh

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// Run the GitHub CLI with args.
// Returns the output from stdout when the invocation exits successfully.
func Run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// NewGraphQLRoundTripper returns a http.RoundTripper that executes GraphQL queries through the GitHub CLI.
func NewGraphQLRoundTripper() http.RoundTripper {
	return &graphQLRoundTripper{}
}

type graphQLRoundTripper struct{}

// RoundTrip implements http.RoundTripper.
func (g graphQLRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("gh GraphQL round tripper: %w", err)
	}
	var requestBody struct {
		Query     string
		Variables map[string]interface{}
	}
	if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
		return nil, fmt.Errorf("gh GraphQL round tripper: %w", err)
	}
	if requestBody.Query == "" {
		return nil, fmt.Errorf("gh GraphQL round tripper: no query provided")
	}
	args := []string{"api", "graphql", "-f", "query=" + requestBody.Query}
	for k, v := range requestBody.Variables {
		args = append(args, "-F")
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	response, err := Run(args...)
	if err != nil {
		return nil, fmt.Errorf("gh GraphQL round tripper: %w", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
	}, nil
}
