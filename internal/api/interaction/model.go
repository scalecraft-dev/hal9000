package interaction

import "github.com/slack-go/slack"

type InteractionModal struct {
	Payload slack.InteractionCallback `form:"payload" json:"payload"`
}
