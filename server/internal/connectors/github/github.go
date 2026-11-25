// Package github is the GitHub connector. Users link their account with OAuth2
// (internal/oauth), and every call is made with that user's token.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/polcortadellas/AREA/server/internal/node"
)

var Service = node.Service{
	Name:        "github",
	Description: "GitHub repositories and issues.",
	OAuth2:      true,
}

const (
	apiURL  = "https://api.github.com"
	timeout = 10 * time.Second
	// GitHub's limit for an issue comment, in characters.
	maxComment = 65536
)

var errNotLinked = errors.New("github: the account is not linked")

var repoRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}/[A-Za-z0-9._-]{1,100}$`)

// The repository is put in the URL path, so only owner/name is accepted.
func checkRepo(repo string) error {
	_, name, _ := strings.Cut(repo, "/")
	if !repoRe.MatchString(repo) || name == "." || name == ".." {
		return fmt.Errorf("%w: repo must look like owner/name", node.ErrInvalidConfig)
	}
	return nil
}

// call sends one request to the GitHub REST API and decodes the answer into out.
func call(ctx context.Context, transport http.RoundTripper, token, method, path string, body, out any) error {
	if token == "" {
		return errNotLinked
	}
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiURL+path, payload)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "AREA")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := (&http.Client{Transport: transport, Timeout: timeout}).Do(req)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("github: reading the response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var apiErr struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &apiErr)
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("github: %d %s", resp.StatusCode, apiErr.Message)
	}
	if out != nil && json.Unmarshal(data, out) != nil {
		return errors.New("github: unexpected response")
	}
	return nil
}
