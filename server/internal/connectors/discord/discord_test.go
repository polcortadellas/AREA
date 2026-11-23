package discord_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/polcortadellas/AREA/server/internal/connectors/discord"
	"github.com/polcortadellas/AREA/server/internal/node"
)

const hook = "https://discord.com/api/webhooks/123/secret-token"

func TestWebhookURLAllowlist(t *testing.T) {
	cases := map[string]bool{
		hook: true,
		"https://discordapp.com/api/webhooks/123/secret-token": true,
		"http://discord.com/api/webhooks/123/secret-token":     false,
		"https://evil.com/api/webhooks/123/secret-token":       false,
		"https://discord.com.evil.com/api/webhooks/123/tok":    false,
		"https://discord.com@evil.com/api/webhooks/123/tok":    false,
		"https://discord.com:8443/api/webhooks/123/tok":        false,
		"https://discord.com/api/webhooks/123/../../users/@me": false,
		"https://discord.com/api/webhooks/123/tok?wait=false":  false,
		"https://discord.com/api/webhooks/123/tok/slack":       false,
		"https://discord.com/api/v10/channels/123/messages":    false,
		"http://localhost:8080/ping":                           false,
		"http://postgres:5432":                                 false,
		"{{ trigger.url }}":                                    false,
		"https://discord.com/api/webhooks/{{ trigger.id }}/x":  false,
	}
	for raw, ok := range cases {
		err := discord.SendMessage{}.ValidateConfig(node.Values{"webhook_url": raw, "content": "hi"})
		if ok && err != nil {
			t.Errorf("%s: unexpected error %v", raw, err)
		}
		if !ok && !errors.Is(err, node.ErrInvalidConfig) {
			t.Errorf("%s: got %v, want ErrInvalidConfig", raw, err)
		}
		if err != nil && strings.Contains(err.Error(), "secret-token") {
			t.Errorf("%s: the error leaks the token: %v", raw, err)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	if err := (discord.SendMessage{}).ValidateConfig(node.Values{"webhook_url": hook}); !errors.Is(err, node.ErrInvalidConfig) {
		t.Errorf("missing content: got %v", err)
	}
	forum := discord.CreateForumThread{}
	if err := forum.ValidateConfig(node.Values{"webhook_url": hook, "content": "x"}); !errors.Is(err, node.ErrInvalidConfig) {
		t.Errorf("missing thread_name: got %v", err)
	}
	mapped := node.Values{"webhook_url": hook, "thread_name": "{{ trigger.title }}", "content": "Opened by {{ trigger.author }}"}
	if err := forum.ValidateConfig(mapped); err != nil {
		t.Errorf("{{ }} references in text fields: %v", err)
	}
}

// fakeDiscord answers like the webhook API. The transport sends every request
// to it, so the URLs in the tests still go through the allowlist.
type fakeDiscord struct {
	status   int
	reply    string
	location string
	requests int
	path     string
	query    string
	body     map[string]any
}

func (f *fakeDiscord) transport(t *testing.T) http.RoundTripper {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests++
		f.path, f.query = r.URL.Path, r.URL.RawQuery
		f.body = nil
		_ = json.NewDecoder(r.Body).Decode(&f.body)
		if f.location != "" {
			w.Header().Set("Location", f.location)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.reply)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r = r.Clone(r.Context())
		r.URL.Scheme, r.URL.Host = target.Scheme, target.Host
		return http.DefaultTransport.RoundTrip(r)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSendMessage(t *testing.T) {
	f := &fakeDiscord{status: http.StatusOK, reply: `{"id":"111","channel_id":"222"}`}
	send := discord.SendMessage{Transport: f.transport(t)}

	out, err := send.Execute(t.Context(), node.Values{"webhook_url": hook, "content": "**New issue** @everyone"}, node.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	if out["message_id"] != "111" {
		t.Errorf("outputs = %v", out)
	}
	if f.path != "/api/webhooks/123/secret-token" || f.query != "wait=true" {
		t.Errorf("request = %s?%s", f.path, f.query)
	}
	if f.body["content"] != "**New issue** @everyone" || f.body["embeds"] != nil {
		t.Errorf("body = %v", f.body)
	}
	if mentions, _ := f.body["allowed_mentions"].(map[string]any); mentions == nil || len(mentions["parse"].([]any)) != 0 {
		t.Errorf("allowed_mentions = %v, want no mentions", f.body["allowed_mentions"])
	}

	if _, err := send.Execute(t.Context(), node.Values{"webhook_url": hook, "content": "Body", "title": "Title"}, node.Credentials{}); err != nil {
		t.Fatal(err)
	}
	embeds, _ := f.body["embeds"].([]any)
	if len(embeds) != 1 || f.body["content"] != nil {
		t.Fatalf("body = %v, want one embed", f.body)
	}
	if e := embeds[0].(map[string]any); e["title"] != "Title" || e["description"] != "Body" {
		t.Errorf("embed = %v", e)
	}

	long := strings.Repeat("a", 2500)
	if _, err := send.Execute(t.Context(), node.Values{"webhook_url": hook, "content": long}, node.Credentials{}); err != nil {
		t.Fatal(err)
	}
	if got := []rune(f.body["content"].(string)); len(got) != 2000 {
		t.Errorf("content has %d characters, want 2000", len(got))
	}
}

func TestCreateForumThread(t *testing.T) {
	f := &fakeDiscord{status: http.StatusOK, reply: `{"id":"111","channel_id":"333"}`}
	forum := discord.CreateForumThread{Transport: f.transport(t)}

	out, err := forum.Execute(t.Context(), node.Values{"webhook_url": hook, "thread_name": "Bug report", "content": "Details"}, node.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	if out["thread_id"] != "333" || out["message_id"] != "111" {
		t.Errorf("outputs = %v", out)
	}
	if f.body["thread_name"] != "Bug report" || f.body["content"] != "Details" {
		t.Errorf("body = %v", f.body)
	}
}

func TestErrors(t *testing.T) {
	values := node.Values{"webhook_url": hook, "content": "hi"}

	f := &fakeDiscord{status: http.StatusNotFound, reply: `{"message":"Unknown Webhook","code":10015}`}
	_, err := discord.SendMessage{Transport: f.transport(t)}.Execute(t.Context(), values, node.Credentials{})
	if err == nil || !strings.Contains(err.Error(), "404 Unknown Webhook") {
		t.Errorf("got %v, want the Discord error", err)
	}

	f = &fakeDiscord{status: http.StatusFound, location: "http://127.0.0.1:1/internal"}
	_, err = discord.SendMessage{Transport: f.transport(t)}.Execute(t.Context(), values, node.Credentials{})
	if err == nil || f.requests != 1 {
		t.Errorf("redirect: err = %v after %d requests, want an error and no second request", err, f.requests)
	}

	down := roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("connection refused") })
	_, err = discord.SendMessage{Transport: down}.Execute(t.Context(), values, node.Credentials{})
	if err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Errorf("network error = %v, want an error without the token", err)
	}

	_, err = discord.SendMessage{Transport: down}.Execute(t.Context(), node.Values{"webhook_url": "http://localhost/x", "content": "hi"}, node.Credentials{})
	if !errors.Is(err, node.ErrInvalidConfig) {
		t.Errorf("bad URL at run time: got %v, want ErrInvalidConfig", err)
	}
}
