package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/polcortadellas/AREA/server/internal/api"
	"github.com/polcortadellas/AREA/server/internal/connectors"
	"github.com/polcortadellas/AREA/server/internal/node"
)

type aboutDoc struct {
	Client struct {
		Host string `json:"host"`
	} `json:"client"`
	Server struct {
		SchemaVersion int            `json:"schema_version"`
		CurrentTime   int64          `json:"current_time"`
		Services      []aboutService `json:"services"`
	} `json:"server"`
}

type aboutService struct {
	Name      string      `json:"name"`
	OAuth2    bool        `json:"oauth2"`
	Actions   []aboutNode `json:"actions"`
	Reactions []aboutNode `json:"reactions"`
}

type aboutNode struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Inputs  node.Schema `json:"inputs"`
	Outputs node.Schema `json:"outputs"`
}

func getAbout(t *testing.T, reg *node.Registry) aboutDoc {
	t.Helper()
	rec := httptest.NewRecorder()
	api.About(reg)(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/about.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var doc aboutDoc
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return doc
}

func findService(t *testing.T, doc aboutDoc, name string) aboutService {
	t.Helper()
	for _, s := range doc.Server.Services {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("service %q missing from %+v", name, doc.Server.Services)
	return aboutService{}
}

func TestAbout(t *testing.T) {
	reg := node.NewRegistry()
	connectors.Register(reg)

	before := time.Now().Unix()
	doc := getAbout(t, reg)

	if doc.Client.Host != "192.0.2.1" { // httptest's default RemoteAddr
		t.Errorf("client.host = %q", doc.Client.Host)
	}
	if doc.Server.SchemaVersion != 1 {
		t.Errorf("schema_version = %d", doc.Server.SchemaVersion)
	}
	if now := time.Now().Unix(); doc.Server.CurrentTime < before || doc.Server.CurrentTime > now {
		t.Errorf("current_time = %d, want between %d and %d", doc.Server.CurrentTime, before, now)
	}

	dummy := findService(t, doc, "dummy")
	if len(dummy.Actions) != 1 || len(dummy.Reactions) != 1 {
		t.Fatalf("dummy: got %d actions and %d reactions", len(dummy.Actions), len(dummy.Reactions))
	}
	counter := dummy.Actions[0]
	if counter.Type != "dummy.counter" || counter.Inputs == nil || len(counter.Inputs) != 0 {
		t.Errorf("counter = %+v, want type dummy.counter and inputs []", counter)
	}
	if len(counter.Outputs) != 1 || counter.Outputs[0].Key != "count" || counter.Outputs[0].Type != node.TypeNumber {
		t.Errorf("counter outputs = %+v", counter.Outputs)
	}
	if echo := dummy.Reactions[0]; echo.Type != "dummy.echo" || len(echo.Inputs) != 1 || echo.Inputs[0].Key != "message" {
		t.Errorf("echo = %+v", echo)
	}
}

func TestAboutEmptyRegistry(t *testing.T) {
	doc := getAbout(t, node.NewRegistry())
	if doc.Server.Services == nil || len(doc.Server.Services) != 0 {
		t.Errorf("services = %v, want []", doc.Server.Services)
	}
}

// A node defined only here shows up without any change to the handler.
func TestAboutListsNewNode(t *testing.T) {
	reg := node.NewRegistry()
	reg.RegisterService(node.Service{Name: "mail", Description: "Test service.", OAuth2: true})
	reg.RegisterReaction(sendMail{})

	mail := findService(t, getAbout(t, reg), "mail")
	if !mail.OAuth2 || len(mail.Actions) != 0 || len(mail.Reactions) != 1 {
		t.Fatalf("mail = %+v", mail)
	}
	send := mail.Reactions[0]
	if send.Name != "send" || send.Type != "mail.send" {
		t.Errorf("reaction = %+v", send)
	}
	if len(send.Inputs) != 2 || send.Inputs[0].Key != "to" || !send.Inputs[0].Required {
		t.Errorf("inputs = %+v", send.Inputs)
	}
	if len(send.Outputs) != 1 || send.Outputs[0].Key != "id" {
		t.Errorf("outputs = %+v", send.Outputs)
	}
}

type sendMail struct{}

func (sendMail) Metadata() node.Metadata {
	return node.Metadata{Service: "mail", Name: "send", Description: "Sends an e-mail."}
}

func (sendMail) InputSchema() node.Schema {
	return node.Schema{
		{Key: "to", Type: node.TypeString, Required: true, Description: "Recipient."},
		{Key: "body", Type: node.TypeString, Description: "Text of the e-mail."},
	}
}

func (sendMail) OutputSchema() node.Schema {
	return node.Schema{{Key: "id", Type: node.TypeString, Required: true, Description: "Message ID."}}
}

func (s sendMail) ValidateConfig(config node.Values) error { return s.InputSchema().Validate(config) }

func (sendMail) Execute(context.Context, node.Values, node.Credentials) (node.Values, error) {
	return node.Values{"id": "1"}, nil
}
