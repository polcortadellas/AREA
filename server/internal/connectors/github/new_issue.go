package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/polcortadellas/AREA/server/internal/node"
)

// NewIssue fires once per issue opened in a repository. The cursor is the
// number of the newest issue seen. The first poll only records it, so a new
// workflow does not fire for the issues that already exist.
type NewIssue struct {
	node.NoWebhook
	// Nil uses http.DefaultTransport; tests point it at a local server.
	Transport http.RoundTripper
}

func (NewIssue) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "new_issue",
		Description: "A new issue is opened in a repository.",
	}
}

func (NewIssue) InputSchema() node.Schema {
	return node.Schema{
		{Key: "repo", Type: node.TypeString, Required: true, Description: "Repository, as owner/name."},
	}
}

func (NewIssue) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "repo", Type: node.TypeString, Required: true, Description: "Repository, as owner/name."},
		{Key: "number", Type: node.TypeNumber, Required: true, Description: "Issue number."},
		{Key: "title", Type: node.TypeString, Required: true, Description: "Issue title."},
		{Key: "body", Type: node.TypeString, Description: "Issue text, empty if there is none."},
		{Key: "author", Type: node.TypeString, Required: true, Description: "Username of the issue author."},
		{Key: "url", Type: node.TypeString, Required: true, Description: "Link to the issue."},
	}
}

func (n NewIssue) ValidateConfig(config node.Values) error {
	if err := n.InputSchema().Validate(config); err != nil {
		return err
	}
	repo, _ := config["repo"].(string)
	return checkRepo(repo)
}

type issue struct {
	ID      int64   `json:"id"`
	Number  int     `json:"number"`
	Title   string  `json:"title"`
	Body    *string `json:"body"`
	HTMLURL string  `json:"html_url"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
	// GitHub lists pull requests as issues too; they carry this field.
	PullRequest *struct{} `json:"pull_request"`
}

func (n NewIssue) Poll(ctx context.Context, config node.Values, cursor string, creds node.Credentials) ([]node.Event, string, error) {
	repo, _ := config["repo"].(string)
	if err := checkRepo(repo); err != nil {
		return nil, "", err
	}
	last := 0
	if cursor != "" {
		var err error
		if last, err = strconv.Atoi(cursor); err != nil {
			return nil, "", fmt.Errorf("github.new_issue: invalid cursor %q: %w", cursor, err)
		}
	}

	// Newest first. More than 50 issues between two polls would lose the oldest ones.
	var issues []issue
	path := "/repos/" + repo + "/issues?state=all&sort=created&direction=desc&per_page=50"
	if err := call(ctx, n.Transport, creds.AccessToken, http.MethodGet, path, nil, &issues); err != nil {
		return nil, "", err
	}

	newest := last
	var events []node.Event
	for i := len(issues) - 1; i >= 0; i-- {
		is := issues[i]
		if is.PullRequest != nil || is.Number <= last {
			continue
		}
		newest = max(newest, is.Number)
		if cursor == "" {
			continue
		}
		body := ""
		if is.Body != nil {
			body = *is.Body
		}
		events = append(events, node.Event{
			ID: strconv.FormatInt(is.ID, 10),
			Outputs: node.Values{
				"repo":   repo,
				"number": float64(is.Number),
				"title":  is.Title,
				"body":   body,
				"author": is.User.Login,
				"url":    is.HTMLURL,
			},
		})
	}
	return events, strconv.Itoa(newest), nil
}
