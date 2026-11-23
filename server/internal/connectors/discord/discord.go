// Package discord posts to Discord channels through webhooks. The webhook URL
// is the credential, so the service does not use OAuth2.
package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/polcortadellas/AREA/server/internal/node"
)

var Service = node.Service{
	Name:        "discord",
	Description: "Posts messages to Discord channels through webhooks.",
}

// Discord limits, in characters.
const (
	maxContent     = 2000
	maxThreadName  = 100
	maxTitle       = 256
	maxDescription = 4096
)

const timeout = 10 * time.Second

var webhookPath = regexp.MustCompile(`^/api/webhooks/[0-9]+/[A-Za-z0-9_-]+$`)

// Only Discord webhook URLs are accepted, so a workflow cannot make the server
// call an internal address (SSRF). It runs when the workflow is saved and
// again before every call.
func checkWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || (u.Host != "discord.com" && u.Host != "discordapp.com") ||
		u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" ||
		!webhookPath.MatchString(u.EscapedPath()) {
		// The URL holds the webhook token, so it never goes into an error.
		return fmt.Errorf("%w: webhook_url must look like https://discord.com/api/webhooks/<id>/<token>", node.ErrInvalidConfig)
	}
	return nil
}

func validate(s node.Schema, config node.Values) error {
	if err := s.Validate(config); err != nil {
		return err
	}
	raw, _ := config["webhook_url"].(string)
	return checkWebhookURL(raw)
}

type message struct {
	Content         string          `json:"content,omitempty"`
	Embeds          []embed         `json:"embeds,omitempty"`
	ThreadName      string          `json:"thread_name,omitempty"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

type embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// An empty list stops text coming from a trigger (an issue title, say) from
// pinging @everyone or other users.
type allowedMentions struct {
	Parse []string `json:"parse"`
}

type posted struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
}

func post(ctx context.Context, transport http.RoundTripper, webhookURL string, msg message) (posted, error) {
	if err := checkWebhookURL(webhookURL); err != nil {
		return posted{}, err
	}
	msg.AllowedMentions.Parse = []string{}
	body, err := json.Marshal(msg)
	if err != nil {
		return posted{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL+"?wait=true", bytes.NewReader(body))
	if err != nil {
		return posted{}, errors.New("discord: cannot build the request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		// A redirect could point anywhere, so it is not followed.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		// *url.Error repeats the URL, and with it the token.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return posted{}, fmt.Errorf("discord: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return posted{}, fmt.Errorf("discord: reading the response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var apiErr struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &apiErr)
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		return posted{}, fmt.Errorf("discord: %d %s", resp.StatusCode, apiErr.Message)
	}
	var out posted
	if err := json.Unmarshal(data, &out); err != nil || out.ID == "" {
		return posted{}, errors.New("discord: unexpected response")
	}
	return out, nil
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
