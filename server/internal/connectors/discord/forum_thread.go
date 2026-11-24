package discord

import (
	"context"
	"net/http"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type CreateForumThread struct {
	// Nil uses http.DefaultTransport; tests point it at a local server.
	Transport http.RoundTripper
}

func (CreateForumThread) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "create_forum_thread",
		Description: "Creates a post in a Discord forum channel.",
	}
}

func (CreateForumThread) InputSchema() node.Schema {
	return node.Schema{
		{Key: "webhook_url", Type: node.TypeString, Required: true, Description: "Webhook URL of a forum channel (channel settings > Integrations > Webhooks). Cannot use {{ }} references."},
		{Key: "thread_name", Type: node.TypeString, Required: true, Description: "Title of the post. Cut at 100 characters."},
		{Key: "content", Type: node.TypeString, Required: true, Description: "First message of the post, Discord markdown allowed. Cut at 2000 characters."},
	}
}

func (CreateForumThread) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "thread_id", Type: node.TypeString, Required: true, Description: "ID of the new post."},
		{Key: "message_id", Type: node.TypeString, Required: true, Description: "ID of its first message."},
	}
}

func (c CreateForumThread) ValidateConfig(config node.Values) error {
	return validate(c.InputSchema(), config)
}

func (c CreateForumThread) Execute(ctx context.Context, inputs node.Values, _ node.Credentials) (node.Values, error) {
	webhookURL, _ := inputs["webhook_url"].(string)
	name, _ := inputs["thread_name"].(string)
	content, _ := inputs["content"].(string)

	out, err := post(ctx, c.Transport, webhookURL, message{
		ThreadName: truncate(name, maxThreadName),
		Content:    truncate(content, maxContent),
	})
	if err != nil {
		return nil, err
	}
	// In a forum, the message is posted in the new thread, so its channel is the thread.
	return node.Values{"thread_id": out.ChannelID, "message_id": out.ID}, nil
}
