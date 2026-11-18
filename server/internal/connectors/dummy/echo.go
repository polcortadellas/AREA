package dummy

import (
	"context"

	"github.com/polcortadellas/AREA/server/internal/node"
)

type Echo struct{}

func (Echo) Metadata() node.Metadata {
	return node.Metadata{
		Service:     Service.Name,
		Name:        "echo",
		Description: "Returns the message it receives.",
	}
}

func (Echo) InputSchema() node.Schema {
	return node.Schema{
		{Key: "message", Type: node.TypeString, Required: true, Description: "Text to return."},
	}
}

func (Echo) OutputSchema() node.Schema {
	return node.Schema{
		{Key: "message", Type: node.TypeString, Required: true, Description: "The same text."},
	}
}

func (e Echo) ValidateConfig(config node.Values) error {
	return e.InputSchema().Validate(config)
}

func (Echo) Execute(_ context.Context, inputs node.Values, _ node.Credentials) (node.Values, error) {
	return node.Values{"message": inputs["message"]}, nil
}
