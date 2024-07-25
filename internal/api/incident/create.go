package incident

import (
	"log"

	"github.com/slack-go/slack"
)

func createIncident(slackApi *slack.Client, req PostHandlerRequest) error {
	modal := createIncidentModal()

	_, err := slackApi.OpenView(req.Body.TriggerId, modal)

	if err != nil {
		log.Printf("Error opening dialog: %s", err)
		return err
	}

	return nil
}
