package incident

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

type PostHandlerRequest struct {
	Body CreateIncidentRequest
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
		ir := CreateIncidentRequest{}

		err := c.ShouldBind(&ir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Request body invalid. Error:": err.Error(),
			})
			return
		}

		req := PostHandlerRequest{
			Body: ir,
		}

		err = commandRouting(slackApi, req)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": "Could not open dialog", "Details": err,
			})
			return
		}
	}
}

func commandRouting(slackApi *slack.Client, req PostHandlerRequest) error {
	text := req.Body.Text
	switch {
	case text == "create":
		return createIncident(slackApi, req)
	case strings.HasPrefix(text, "timeline "):
		timelineText := strings.TrimPrefix(text, "timeline ")
		return AddTimelineItem(slackApi, req.Body.ChannelId, req.Body.UserId, timelineText)
	case text == "update":
		return updateIncident(slackApi, req)
	case strings.HasPrefix(text, "action-item "):
		actionItemText := strings.TrimPrefix(text, "action-item ")
		return AddActionItem(slackApi, req.Body.ChannelId, req.Body.UserId, actionItemText)
	case text == "help":
		slackApi.PostEphemeral(req.Body.ChannelId, req.Body.UserId, slack.MsgOptionBlocks(HelpMessage().BlockSet...))
	case text == "":
		slackApi.PostEphemeral(req.Body.ChannelId, req.Body.UserId, slack.MsgOptionText("Please provide a command. Use `/incident help` for details on available commands.", false))
	default:
		slackApi.PostEphemeral(req.Body.ChannelId, req.Body.UserId, slack.MsgOptionText(fmt.Sprintf("Command `%s` not found. Use `/incident help` for details on available commands.", req.Body.Text), false))
	}

	return nil
}
