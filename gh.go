package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func gh(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

type ghRoundTripper struct{}

func (g ghRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("gh round tripper: %w", err)
	}
	var requestBody struct {
		Query     string
		Variables map[string]interface{}
	}
	if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
		return nil, fmt.Errorf("gh round tripper: %w", err)
	}
	if requestBody.Query == "" {
		return nil, fmt.Errorf("gh round tripper: no query provided")
	}
	args := []string{"api", "graphql", "-f", "query=" + requestBody.Query}
	for k, v := range requestBody.Variables {
		args = append(args, "-F")
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	response, err := gh(args...)
	if err != nil {
		return nil, fmt.Errorf("gh round tripper: %w", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
	}, nil
}
