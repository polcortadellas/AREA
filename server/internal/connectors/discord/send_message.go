package discord

import (
	"context"
	"net/http"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type SendMessage struct {
	// Nil uses http.DefaultTransport; tests point it at a local server.
	Transport http.RoundTripper
}

func (SendMessage) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "send_message",
		Description: "Posts a message to a Discord channel. With a title it is shown as an embed card.",
	}
}

func (SendMessage) InputSchema() node.Schema {
	return node.Schema{
		{Key: "webhook_url", Type: node.TypeString, Required: true, Description: "Webhook URL of the channel (channel settings > Integrations > Webhooks). Cannot use {{ }} references."},
		{Key: "content", Type: node.TypeString, Required: true, Description: "Message text, Discord markdown allowed. Cut at 2000 characters."},
		{Key: "title", Type: node.TypeString, Description: "Optional title. When set, the message is sent as an embed card."},
	}
}

func (SendMessage) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "message_id", Type: node.TypeString, Required: true, Description: "ID of the posted message."},
	}
}

func (s SendMessage) ValidateConfig(config node.Values) error {
	return validate(s.InputSchema(), config)
}

func (s SendMessage) Execute(ctx context.Context, inputs node.Values, _ node.Credentials) (node.Values, error) {
	webhookURL, _ := inputs["webhook_url"].(string)
	content, _ := inputs["content"].(string)
	title, _ := inputs["title"].(string)

	var msg message
	if title != "" {
		msg.Embeds = []embed{{Title: truncate(title, maxTitle), Description: truncate(content, maxDescription)}}
	} else {
		msg.Content = truncate(content, maxContent)
	}

	out, err := post(ctx, s.Transport, webhookURL, msg)
	if err != nil {
		return nil, err
	}
	return node.Values{"message_id": out.ID}, nil
}
