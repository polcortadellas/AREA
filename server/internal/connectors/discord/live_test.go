package discord_test

import (
	"os"
	"testing"
	"time"

	"github.com/polcortadellas/AREA/server/internal/connectors/discord"
	"github.com/polcortadellas/AREA/server/internal/node"
)

// The live tests post to real channels and are skipped unless the webhook of
// a text channel (DISCORD_TEST_WEBHOOK_URL) or of a forum channel
// (DISCORD_TEST_FORUM_WEBHOOK_URL) is set.

func liveWebhook(t *testing.T, env string) string {
	t.Helper()
	url := os.Getenv(env)
	if url == "" {
		t.Skip(env + " is not set")
	}
	return url
}

func TestLiveSendMessage(t *testing.T) {
	webhook := liveWebhook(t, "DISCORD_TEST_WEBHOOK_URL")
	out, err := discord.SendMessage{}.Execute(t.Context(), node.Values{
		"webhook_url": webhook,
		"title":       "AREA test " + time.Now().Format(time.DateTime),
		"content":     "Sent by `TestLiveSendMessage` with **markdown**. This @everyone must not ping anyone.",
	}, node.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("outputs: %v", out)
}

func TestLiveCreateForumThread(t *testing.T) {
	webhook := liveWebhook(t, "DISCORD_TEST_FORUM_WEBHOOK_URL")
	out, err := discord.CreateForumThread{}.Execute(t.Context(), node.Values{
		"webhook_url": webhook,
		"thread_name": "AREA test " + time.Now().Format(time.DateTime),
		"content":     "Post created by `TestLiveCreateForumThread`.",
	}, node.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("outputs: %v", out)
}
