package incident

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/slack-go/slack"
)

type PostHandlerRequest struct {
	Body CreateIncidentRequest
}

type PostHandlerResponse struct {
}

func PostHandler(slackApi *slack.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ir := CreateIncidentRequest{}

		err := c.MustBindWith(&ir, binding.Form)
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

		// channel, err := createChannel(slackApi)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"Error": "Could not create channel", "Details": err,
		// 	})
		// 	return
		// }

		// err = addChannelParticipants(sr, slackApi, channel.ID)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"Error": "Could not add participants to channel", "Details": err,
		// 	})
		// 	return
		// }
	}
}

func commandRouting(slackApi *slack.Client, req PostHandlerRequest) error {
	text := req.Body.Text
	switch {
	case text == "create":
		return createIncident(slackApi, req)
	case strings.HasPrefix(text, "timeline"):
		timelineText := strings.TrimPrefix(text, "timeline ")
		return AddTimelineItem(slackApi, req.Body.ChannelId, req.Body.UserId, timelineText)
	case text == "update":
		return updateIncident(slackApi, req)
	case text == "":
		slackApi.PostEphemeral(req.Body.ChannelId, req.Body.UserId, slack.MsgOptionText("Please provide a command", false))
	default:
		slackApi.PostEphemeral(req.Body.ChannelId, req.Body.UserId, slack.MsgOptionText(fmt.Sprintf("Command `%s` not found", req.Body.Text), false))
	}

	return nil
}
