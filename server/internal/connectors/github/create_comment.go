package github

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type CreateComment struct {
	// Nil uses http.DefaultTransport; tests point it at a local server.
	Transport http.RoundTripper
}

func (CreateComment) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "create_comment",
		Description: "Posts a comment on an issue.",
	}
}

func (CreateComment) InputSchema() node.Schema {
	return node.Schema{
		{Key: "repo", Type: node.TypeString, Required: true, Description: "Repository, as owner/name, e.g. {{ trigger.repo }}."},
		{Key: "issue_number", Type: node.TypeNumber, Required: true, Description: "Issue number, e.g. {{ trigger.number }}."},
		{Key: "body", Type: node.TypeString, Required: true, Description: "Comment text, Markdown allowed. Cut at 65536 characters."},
	}
}

func (CreateComment) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "comment_id", Type: node.TypeString, Required: true, Description: "ID of the comment."},
		{Key: "comment_url", Type: node.TypeString, Required: true, Description: "Link to the comment."},
	}
}

// References are only checked once they are resolved, in Execute.
func (c CreateComment) ValidateConfig(config node.Values) error {
	if err := c.InputSchema().Validate(config); err != nil {
		return err
	}
	if repo, ok := config["repo"].(string); ok && !strings.Contains(repo, "{{") {
		if err := checkRepo(repo); err != nil {
			return err
		}
	}
	if n, ok := config["issue_number"].(float64); ok {
		if _, err := issueNumber(n); err != nil {
			return err
		}
	}
	return nil
}

func (c CreateComment) Execute(ctx context.Context, inputs node.Values, creds node.Credentials) (node.Values, error) {
	repo, _ := inputs["repo"].(string)
	if err := checkRepo(repo); err != nil {
		return nil, err
	}
	number, err := issueNumber(inputs["issue_number"])
	if err != nil {
		return nil, err
	}
	body, _ := inputs["body"].(string)
	if r := []rune(body); len(r) > maxComment {
		body = string(r[:maxComment-1]) + "…"
	}

	var comment struct {
		ID      int64  `json:"id"`
		HTMLURL string `json:"html_url"`
	}
	path := "/repos/" + repo + "/issues/" + strconv.Itoa(number) + "/comments"
	if err := call(ctx, c.Transport, creds.AccessToken, http.MethodPost, path, map[string]string{"body": body}, &comment); err != nil {
		return nil, err
	}
	return node.Values{"comment_id": strconv.FormatInt(comment.ID, 10), "comment_url": comment.HTMLURL}, nil
}

// issueNumber accepts a positive whole number, as JSON or as text.
func issueNumber(v any) (int, error) {
	var f float64
	switch n := v.(type) {
	case float64:
		f = n
	case string:
		var err error
		if f, err = strconv.ParseFloat(n, 64); err != nil {
			return 0, fmt.Errorf("%w: issue_number must be a number", node.ErrInvalidConfig)
		}
	default:
		return 0, fmt.Errorf("%w: issue_number must be a number", node.ErrInvalidConfig)
	}
	if f < 1 || f != math.Trunc(f) || f > math.MaxInt32 {
		return 0, fmt.Errorf("%w: issue_number must be a positive whole number", node.ErrInvalidConfig)
	}
	return int(f), nil
}
