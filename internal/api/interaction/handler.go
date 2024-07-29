package interaction

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/Imagine-Pediatrics/hal/internal/api/incident"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

type PostHandlerRequest struct {
	Body InteractionModal
}

type PostHandlerResponse struct {
}

func PostHandler(slackApi *slack.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("x-valid-slack-request") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Invalid request",
			})
			return
		}
		interaction := InteractionModal{}

		err := c.ShouldBind(&interaction)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Request body invalid. Error:": err.Error(),
			})
			return
		}

		req := &PostHandlerRequest{
			Body: interaction,
		}

		response, err := Interaction(slackApi, req)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"internal service error loading data.": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func Interaction(slackApi *slack.Client, req *PostHandlerRequest) (*PostHandlerResponse, error) {
	switch req.Body.Payload.Type {
	case slack.InteractionTypeViewSubmission:
		switch req.Body.Payload.View.CallbackID {
		case "create_incident_modal":
			description := req.Body.Payload.View.State.Values["description"]["description"].Value
			status := req.Body.Payload.View.State.Values["status"]["status"].SelectedOption.Text.Text
			users := req.Body.Payload.View.State.Values["incident_members"]["incident_members"].SelectedUsers
			severity := req.Body.Payload.View.State.Values["incident_severity"]["incident_severity"].SelectedOption.Text.Text
			users = append(users, req.Body.Payload.User.ID)

			channel, err := incident.CreateIncidentChannel(
				slackApi,
				fmt.Sprintf("%s incident: %s", severity, description),
				users...,
			)
			if err != nil {
				return nil, err
			}

			err = incident.CreateTimeline(
				slackApi,
				channel.ID,
				req.Body.Payload.User.Name,
				severity,
				status,
			)
			if err != nil {
				return nil, err
			}

			err = incident.CreateActionItems(
				slackApi,
				channel.ID,
				req.Body.Payload.User.Name,
			)

			if severity == "SEV-1" || severity == "SEV-2" {
				err = incident.AddActionItem(
					slackApi,
					channel.ID,
					req.Body.Payload.User.Name,
					"Create incident postmortem",
				)
				if err != nil {
					log.Printf("Error adding action item: %s", err)
					return nil, err
				}
			}
			slackApi.PostEphemeral(channel.ID, req.Body.Payload.User.ID, slack.MsgOptionBlocks(incident.HelpMessage().BlockSet...))

		case "update_incident_modal":
			status := req.Body.Payload.View.State.Values["status"]["status"].SelectedOption.Text.Text
			severity := req.Body.Payload.View.State.Values["incident_severity"]["incident_severity"].SelectedOption.Text.Text
			channelID := req.Body.Payload.View.PrivateMetadata

			if severity == "SEV-1" || severity == "SEV-2" {
				err := incident.AddActionItem(
					slackApi,
					channelID,
					req.Body.Payload.User.Name,
					"Create incident postmortem",
				)
				if err != nil {
					log.Printf("Error adding action item: %s", err)
					return nil, err
				}
			}

			err := incident.AddTimelineItem(
				slackApi,
				channelID,
				req.Body.Payload.User.Name,
				fmt.Sprintf("Incident updated. Status: %s, Severity: %s", status, severity),
			)
			if err != nil {
				log.Printf("Error adding timeline item: %s", err)
				return nil, err
			}
		}
	default:
		slog.Debug(fmt.Sprintf("unknown interaction type: %s", req.Body.Payload.Type))
	}

	return &PostHandlerResponse{}, nil
}
