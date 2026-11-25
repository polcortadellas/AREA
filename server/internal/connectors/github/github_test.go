package github_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/connectors/github"
	"github.com/polcortadellas/AREA/server/internal/node"
)

var creds = node.Credentials{AccessToken: "secret-token"}

// fakeGitHub answers like the REST API. The transport sends every request to
// it, so the code under test still builds https://api.github.com URLs.
type fakeGitHub struct {
	status  int
	reply   string
	method  string
	path    string
	query   url.Values
	headers http.Header
	body    map[string]any
}

func (f *fakeGitHub) transport(t *testing.T) http.RoundTripper {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.method, f.path, f.query, f.headers = r.Method, r.URL.Path, r.URL.Query(), r.Header
		f.body = nil
		_ = json.NewDecoder(r.Body).Decode(&f.body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.reply)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.github.com" {
			t.Errorf("request to %s, want api.github.com", r.URL.Host)
		}
		r = r.Clone(r.Context())
		r.URL.Scheme, r.URL.Host = target.Scheme, target.Host
		return http.DefaultTransport.RoundTrip(r)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRepoValidation(t *testing.T) {
	cases := map[string]bool{
		"octocat/hello-world":  true,
		"edu24epitech/area.v2": true,
		"octocat":              false,
		"octocat/hello/world":  false,
		"octocat/..":           false,
		"../etc":               false,
		"octo cat/repo":        false,
		"octocat/repo?x=1":     false,
		"{{ trigger.repo }}":   false,
	}
	for repo, ok := range cases {
		err := github.NewIssue{}.ValidateConfig(node.Values{"repo": repo})
		if ok != (err == nil) {
			t.Errorf("%q: got %v", repo, err)
		}
		if err != nil && !errors.Is(err, node.ErrInvalidConfig) {
			t.Errorf("%q: got %v, want ErrInvalidConfig", repo, err)
		}
	}
}

func TestCreateCommentConfig(t *testing.T) {
	cases := []struct {
		config node.Values
		ok     bool
	}{
		{node.Values{"repo": "{{ trigger.repo }}", "issue_number": "{{ trigger.number }}", "body": "Thanks {{ trigger.author }}"}, true},
		{node.Values{"repo": "octocat/hello", "issue_number": float64(7), "body": "hi"}, true},
		{node.Values{"repo": "octocat/hello", "issue_number": 7.5, "body": "hi"}, false},
		{node.Values{"repo": "octocat/hello", "issue_number": float64(0), "body": "hi"}, false},
		{node.Values{"repo": "not a repo", "issue_number": float64(7), "body": "hi"}, false},
		{node.Values{"repo": "octocat/hello", "issue_number": float64(7)}, false},
	}
	for _, tc := range cases {
		err := github.CreateComment{}.ValidateConfig(tc.config)
		if tc.ok != (err == nil) {
			t.Errorf("%v: got %v", tc.config, err)
		}
	}
}

const issues = `[
	{"id": 104, "number": 4, "title": "Fourth", "body": null, "html_url": "https://github.com/o/r/issues/4", "user": {"login": "bob"}},
	{"id": 103, "number": 3, "title": "A pull request", "html_url": "https://github.com/o/r/pull/3", "user": {"login": "bob"}, "pull_request": {}},
	{"id": 102, "number": 2, "title": "Second", "body": "Details", "html_url": "https://github.com/o/r/issues/2", "user": {"login": "alice"}},
	{"id": 101, "number": 1, "title": "First", "body": "Old", "html_url": "https://github.com/o/r/issues/1", "user": {"login": "alice"}}
]`

func TestNewIssue(t *testing.T) {
	f := &fakeGitHub{status: http.StatusOK, reply: issues}
	trigger := github.NewIssue{Transport: f.transport(t)}
	config := node.Values{"repo": "o/r"}

	events, next, err := trigger.Poll(t.Context(), config, "", creds)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 || next != "4" {
		t.Fatalf("first poll: %d events, cursor %q; want none and 4", len(events), next)
	}
	if f.path != "/repos/o/r/issues" || f.query.Get("sort") != "created" || f.query.Get("state") != "all" {
		t.Errorf("request = %s?%s", f.path, f.query.Encode())
	}
	if f.headers.Get("Authorization") != "Bearer secret-token" || f.headers.Get("User-Agent") == "" ||
		f.headers.Get("X-GitHub-Api-Version") == "" {
		t.Errorf("headers = %v", f.headers)
	}

	events, next, err = trigger.Poll(t.Context(), config, "1", creds)
	if err != nil {
		t.Fatal(err)
	}
	if next != "4" || len(events) != 2 {
		t.Fatalf("got %d events and cursor %q, want issues 2 and 4 (not the pull request) and cursor 4", len(events), next)
	}
	second, fourth := events[0], events[1]
	if second.ID != "102" || second.Outputs["number"] != float64(2) || second.Outputs["author"] != "alice" ||
		second.Outputs["body"] != "Details" || second.Outputs["repo"] != "o/r" {
		t.Errorf("oldest event first, got %+v", second)
	}
	if fourth.Outputs["body"] != "" || fourth.Outputs["url"] != "https://github.com/o/r/issues/4" {
		t.Errorf("issue without body: %+v", fourth)
	}

	if events, next, _ := trigger.Poll(t.Context(), config, "4", creds); len(events) != 0 || next != "4" {
		t.Errorf("nothing new: got %d events, cursor %q", len(events), next)
	}
}

func TestCreateComment(t *testing.T) {
	f := &fakeGitHub{status: http.StatusCreated, reply: `{"id": 555, "html_url": "https://github.com/o/r/issues/7#issuecomment-555"}`}
	reaction := github.CreateComment{Transport: f.transport(t)}

	out, err := reaction.Execute(t.Context(), node.Values{"repo": "o/r", "issue_number": float64(7), "body": "Thanks!"}, creds)
	if err != nil {
		t.Fatal(err)
	}
	if f.method != http.MethodPost || f.path != "/repos/o/r/issues/7/comments" || f.body["body"] != "Thanks!" {
		t.Errorf("request = %s %s %v", f.method, f.path, f.body)
	}
	if out["comment_id"] != "555" || out["comment_url"] != "https://github.com/o/r/issues/7#issuecomment-555" {
		t.Errorf("outputs = %v", out)
	}

	long := strings.Repeat("a", 70000)
	if _, err := reaction.Execute(t.Context(), node.Values{"repo": "o/r", "issue_number": "7", "body": long}, creds); err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(f.body["body"].(string))); got != 65536 {
		t.Errorf("body has %d characters, want 65536", got)
	}
}

func TestErrors(t *testing.T) {
	f := &fakeGitHub{status: http.StatusNotFound, reply: `{"message": "Not Found"}`}
	reaction := github.CreateComment{Transport: f.transport(t)}
	values := node.Values{"repo": "o/r", "issue_number": float64(7), "body": "hi"}

	if _, err := reaction.Execute(t.Context(), values, creds); err == nil || !strings.Contains(err.Error(), "404 Not Found") {
		t.Errorf("got %v, want the GitHub error", err)
	}
	if _, err := reaction.Execute(t.Context(), values, node.Credentials{}); err == nil || !strings.Contains(err.Error(), "not linked") {
		t.Errorf("without a token: got %v", err)
	}
	if _, _, err := (github.NewIssue{Transport: f.transport(t)}).Poll(t.Context(), node.Values{"repo": "o/r"}, "", node.Credentials{}); err == nil {
		t.Error("polling without a token should fail")
	}
	bad := node.Values{"repo": "o/r", "issue_number": "{{ x }}", "body": "hi"}
	if _, err := reaction.Execute(t.Context(), bad, creds); !errors.Is(err, node.ErrInvalidConfig) {
		t.Errorf("unresolved issue number: got %v", err)
	}
	if _, _, err := (github.NewIssue{}).Poll(t.Context(), node.Values{"repo": "o/r"}, "x", creds); err == nil {
		t.Error("an invalid cursor should fail")
	}
}
