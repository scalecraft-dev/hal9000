package incident

import (
	"github.com/slack-go/slack"
)

func updateIncident(slackApi *slack.Client, req PostHandlerRequest) error {
	modal := updateIncidentModal()
	modal.PrivateMetadata = req.Body.ChannelId

	_, err := slackApi.OpenView(req.Body.TriggerId, modal)

	if err != nil {
		return err
	}

	return nil
}
