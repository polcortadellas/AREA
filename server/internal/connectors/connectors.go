package connectors

import (
	"github.com/polcortadellas/AREA/server/internal/connectors/discord"
	"github.com/polcortadellas/AREA/server/internal/connectors/dummy"
	"github.com/polcortadellas/AREA/server/internal/node"
)

// Adding a node only needs its own file and one line here.
func Register(r *node.Registry) {
	r.RegisterService(dummy.Service)
	r.RegisterTrigger(dummy.Counter{})
	r.RegisterReaction(dummy.Echo{})

	r.RegisterService(discord.Service)
	r.RegisterReaction(discord.SendMessage{})
	r.RegisterReaction(discord.CreateForumThread{})
}
