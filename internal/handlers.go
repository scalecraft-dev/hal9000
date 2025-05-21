package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func SlackAuthMiddleware(signingSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/health" {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("Failed to read request body", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		timestamp := c.GetHeader("X-Slack-Request-Timestamp")
		signature := c.GetHeader("X-Slack-Signature")

		_, err = validateTimestamp(timestamp)
		if err != nil {
			slog.Warn("Invalid timestamp in Slack request", "timestamp", timestamp, "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid timestamp"})
			return
		}

		expectedSignature := computeSignature(signingSecret, timestamp, bodyBytes)

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			slog.Warn("Invalid Slack signature", "provided", signature, "expected", expectedSignature)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Slack signature"})
			return
		}

		c.Set("x-valid-slack-request", true)
		c.Next()
	}
}

func validateTimestamp(timestamp string) (int64, error) {
	if timestamp == "" {
		return 0, fmt.Errorf("empty timestamp")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Check if timestamp is older than 5 minutes
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute).Unix()
	if ts < fiveMinutesAgo {
		return 0, fmt.Errorf("timestamp too old")
	}

	return ts, nil
}

func computeSignature(signingSecret, timestamp string, body []byte) string {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(baseString))
	signature := fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))
	return signature
}

func HealthHandler(slackService *SlackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		err := slackService.HealthCheck(ctx)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"slack":  "unavailable",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"slack":  "available",
		})
	}
}

func IncidentHandler(incidentService *IncidentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("x-valid-slack-request") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Invalid request",
			})
			return
		}

		var req SlackCommandRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		ctx := c.Request.Context()
		slackClient := c.MustGet("slackApi").(*slack.Client)

		commandArgs := strings.SplitN(req.Text, " ", 2)
		command := strings.ToLower(commandArgs[0])
		args := ""
		if len(commandArgs) > 1 {
			args = commandArgs[1]
		}

		var err error // Declare error variable once for the handler

		switch command {
		case "create", "c":
			err = incidentService.CreateIncident(ctx, req.TriggerId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open dialog", "details": err.Error()})
				return
			}

		case "update", "u":
			err = incidentService.UpdateIncident(ctx, req.TriggerId, req.ChannelId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open update dialog", "details": err.Error()})
				return
			}

		case "timeline", "t":
			if args == "" {
				_, postErr := slackClient.PostEphemeral(req.ChannelId, req.UserId,
					slack.MsgOptionText("Usage: /incident timeline <message> (or /incident t <message>)", false))
				if postErr != nil {
					slog.ErrorContext(ctx, "Failed to send timeline usage message", "error", postErr)
				}
				c.Status(http.StatusOK) // Acknowledge command even if usage message fails
				return
			}
			err = incidentService.AddTimelineItem(ctx, req.ChannelId, req.UserId, args)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add timeline item", "details": err.Error()})
				return
			}

		case "action-item", "ai":
			if args == "" {
				_, postErr := slackClient.PostEphemeral(req.ChannelId, req.UserId,
					slack.MsgOptionText("Usage: /incident action-item <description> (or /incident ai <description>)", false))
				if postErr != nil {
					slog.ErrorContext(ctx, "Failed to send action-item usage message", "error", postErr)
				}
				c.Status(http.StatusOK)
				return
			}
			err = incidentService.AddActionItem(ctx, req.ChannelId, req.UserId, args)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add action item", "details": err.Error()})
				return
			}

		case "resolve", "r":
			err = incidentService.ResolveIncident(ctx, req.ChannelId, req.UserId, req.UserName, args)
			if err != nil {
				// ResolveIncident itself logs specific errors and returns nil for overall success
				// However, if it were to return a critical error, handle it here.
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve incident", "details": err.Error()})
				return
			}
			// Ephemeral confirmation is handled within ResolveIncident

		case "help", "h":
			blocks := incidentService.HelpMessage()
			err = incidentService.slackService.PostEphemeralMessage(ctx, req.ChannelId, req.UserId, blocks.BlockSet)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to send help message via PostEphemeralMessage", "channelID", req.ChannelId, "userID", req.UserId, "original_error", err)
				if strings.Contains(err.Error(), "not_in_channel") {
					c.JSON(http.StatusOK, gin.H{
						"response_type": "ephemeral",
						"text":          "Oops! I can't display the help message here because I'm not a member of this channel. Please add me to this channel, try `/incident help` in a channel I'm in, or send me a direct message with `help`.",
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not send help message", "details": err.Error()})
				}
				return // Return after handling the error response
			}

		case "": // Handles the case where only /incident is typed
			_, err = slackClient.PostEphemeral(req.ChannelId, req.UserId,
				slack.MsgOptionText("Please provide a command. Use `/incident help` (or `/incident h`) for details.", false))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not send message", "details": err.Error()})
				return
			}

		default:
			_, err = slackClient.PostEphemeral(req.ChannelId, req.UserId,
				slack.MsgOptionText(fmt.Sprintf("Command `%s` not found. Use `/incident help` (or `/incident h`) for details.", command), false))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not send message", "details": err.Error()})
				return
			}
		}

		c.Status(http.StatusOK)
	}
}

func InteractionHandler(incidentService *IncidentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("x-valid-slack-request") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Invalid request",
			})
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "Failed to read request body in InteractionHandler", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		parsedForm, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "Failed to parse form data from request body", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Malformed form data"})
			return
		}

		payloadFormParam := parsedForm.Get("payload")
		if payloadFormParam == "" {
			slog.ErrorContext(c.Request.Context(), "Missing 'payload' in form data for interaction")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing payload in form data"})
			return
		}

		var interaction slack.InteractionCallback
		if err := json.Unmarshal([]byte(payloadFormParam), &interaction); err != nil {
			slog.ErrorContext(c.Request.Context(), "Failed to unmarshal interaction callback from payload", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interaction payload format"})
			return
		}

		ctx := c.Request.Context()
		slackClient := c.MustGet("slackApi").(*slack.Client)

		slog.InfoContext(ctx, "Received interaction",
			"type", interaction.Type,
			"callbackID", interaction.View.CallbackID,
			"userID", interaction.User.ID)

		switch interaction.Type {
		case slack.InteractionTypeViewSubmission:
			switch interaction.View.CallbackID {
			case "create_incident_modal":
				description := interaction.View.State.Values["description"]["description"].Value
				status := interaction.View.State.Values["status"]["status"].SelectedOption.Text.Text
				users := interaction.View.State.Values["incident_members"]["incident_members"].SelectedUsers
				severityText := interaction.View.State.Values["incident_severity"]["incident_severity"].SelectedOption.Text.Text
				severity := Severity(severityText)

				var incidentCommanderID string
				if commanderState, ok := interaction.View.State.Values["incident_commander"]["incident_commander"]; ok {
					if commanderState.SelectedUser != "" { // For single user select, it's SelectedUser
						incidentCommanderID = commanderState.SelectedUser
					} else if len(commanderState.SelectedUsers) > 0 { // Fallback if it's behaving like multi-user select
						incidentCommanderID = commanderState.SelectedUsers[0]
					}
				}

				var commsRepresentativeID string
				if commsRepState, ok := interaction.View.State.Values["comms_representative"]["comms_representative"]; ok {
					if commsRepState.SelectedUser != "" { // For single user select
						commsRepresentativeID = commsRepState.SelectedUser
					} else if len(commsRepState.SelectedUsers) > 0 { // Fallback for multi-user select behavior
						commsRepresentativeID = commsRepState.SelectedUsers[0]
					}
				}

				usersToInvite := append(users, interaction.User.ID)
				if incidentCommanderID != "" {
					usersToInvite = appendIfMissing(usersToInvite, incidentCommanderID)
				}
				if commsRepresentativeID != "" {
					usersToInvite = appendIfMissing(usersToInvite, commsRepresentativeID)
				}

				channel, err := incidentService.CreateIncidentChannel(
					ctx,
					description,
					severity,
					Status(status),
					incidentCommanderID,
					commsRepresentativeID,
					usersToInvite...,
				)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to create incident channel", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create incident channel",
						"details": err.Error(),
					})
					return
				}

				err = incidentService.CreateTimeline(
					ctx,
					channel.ID,
					interaction.User.Name,
					severity,
					Status(status),
				)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to create timeline", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create timeline",
						"details": err.Error(),
					})
					return
				}

				err = incidentService.CreateActionItems(
					ctx,
					channel.ID,
					interaction.User.Name,
				)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to create action items", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create action items",
						"details": err.Error(),
					})
					return
				}

				if severity == SeveritySev1 || severity == SeveritySev2 {
					err = incidentService.AddActionItem(
						ctx,
						channel.ID,
						interaction.User.Name,
						"Create incident postmortem",
					)
					if err != nil {
						slog.WarnContext(ctx, "Failed to add postmortem action item", "error", err)
					}
				}

				blocks := incidentService.HelpMessage()
				_, err = slackClient.PostEphemeral(channel.ID, interaction.User.ID, slack.MsgOptionBlocks(blocks.BlockSet...))
				if err != nil {
					slog.WarnContext(ctx, "Failed to send help message", "error", err)
				}

			case "update_incident_modal":
				newStatus := interaction.View.State.Values["status"]["status"].SelectedOption.Text.Text
				newSeverityText := interaction.View.State.Values["incident_severity"]["incident_severity"].SelectedOption.Text.Text
				newSeverity := Severity(newSeverityText)
				channelID := interaction.View.PrivateMetadata

				var newCommanderID string
				if commanderState, ok := interaction.View.State.Values["incident_commander"]["incident_commander"]; ok {
					if commanderState.SelectedUser != "" {
						newCommanderID = commanderState.SelectedUser
					} else if len(commanderState.SelectedUsers) > 0 { // Fallback for multi-user select behavior
						newCommanderID = commanderState.SelectedUsers[0]
					}
				}

				var newCommsRepID string
				if commsRepState, ok := interaction.View.State.Values["comms_representative"]["comms_representative"]; ok {
					if commsRepState.SelectedUser != "" {
						newCommsRepID = commsRepState.SelectedUser
					} else if len(commsRepState.SelectedUsers) > 0 { // Fallback for multi-user select behavior
						newCommsRepID = commsRepState.SelectedUsers[0]
					}
				}

				// Get current channel topic to extract original description and old commander/comms for comparison
				currentTopic, err := incidentService.slackService.GetChannelTopic(ctx, channelID)
				originalDescription := "Description unavailable" // Fallback
				oldCommanderID := ""
				oldCommsRepID := ""

				if err != nil {
					slog.ErrorContext(ctx, "Failed to get channel topic for update", "channelID", channelID, "error", err)
				} else {
					// Parse existing topic
					coreTopic := currentTopic
					if idx := strings.Index(coreTopic, " | Commander: <@"); idx != -1 {
						start := idx + len(" | Commander: <@")
						end := strings.Index(coreTopic[start:], ">")
						if end != -1 {
							oldCommanderID = coreTopic[start : start+end]
						}
						coreTopic = coreTopic[:idx] // Remove for description parsing
					}
					if idx := strings.Index(coreTopic, " | Comms: <@"); idx != -1 {
						start := idx + len(" | Comms: <@")
						end := strings.Index(coreTopic[start:], ">")
						if end != -1 {
							oldCommsRepID = coreTopic[start : start+end]
						}
						coreTopic = coreTopic[:idx] // Remove for description parsing
					}
					parts := strings.SplitN(coreTopic, " incident: ", 2)
					if len(parts) == 2 {
						originalDescription = parts[1]
					} else if strings.HasPrefix(coreTopic, "Resolved: ") {
						originalDescription = strings.TrimPrefix(coreTopic, "Resolved: ")
					} else if coreTopic != "" {
						originalDescription = coreTopic // Fallback
					}
				}

				// Construct new topic
				topicPrefix := fmt.Sprintf("%s incident:", newSeverity)
				if strings.HasPrefix(originalDescription, "Resolved: ") { // If it was resolved, keep it resolved, but update severity for record? Or prevent this state? For now, assume severity update implies active.
					topicPrefix = fmt.Sprintf("%s incident:", newSeverity)
				} else if strings.HasPrefix(currentTopic, "Resolved: ") { // If original topic was resolved, new topic should also indicate resolved state, but with potentially new description/roles
					topicPrefix = "Resolved:"
				}

				topicParts := []string{fmt.Sprintf("%s %s", topicPrefix, originalDescription)}
				if newCommanderID != "" {
					topicParts = append(topicParts, fmt.Sprintf("Commander: <@%s>", newCommanderID))
				}
				if newCommsRepID != "" {
					topicParts = append(topicParts, fmt.Sprintf("Comms: <@%s>", newCommsRepID))
				}
				newTopic := strings.Join(topicParts, " | ")

				if newTopic != currentTopic { // Only update if there's a change
					err = incidentService.slackService.SetChannelTopic(ctx, channelID, newTopic)
					if err != nil {
						slog.WarnContext(ctx, "Failed to update channel topic", "channelID", channelID, "newTopic", newTopic, "error", err)
					}
				}

				// Invite new commander/comms rep if they changed and are not empty
				usersToInvite := []string{}
				if newCommanderID != "" && newCommanderID != oldCommanderID {
					usersToInvite = appendIfMissing(usersToInvite, newCommanderID)
				}
				if newCommsRepID != oldCommsRepID {
					usersToInvite = appendIfMissing(usersToInvite, newCommsRepID)
				}
				if len(usersToInvite) > 0 {
					err = incidentService.slackService.InviteUsersToChannel(ctx, channelID, usersToInvite...)
					if err != nil {
						slog.WarnContext(ctx, "Failed to invite new commander/comms to channel", "channelID", channelID, "users", usersToInvite, "error", err)
					}
				}

				// Add timeline item for the update (status, severity, and potentially roles)
				updateMessages := []string{fmt.Sprintf("Incident updated. Status: %s, Severity: %s", newStatus, newSeverity)}
				if newCommanderID != oldCommanderID {
					if newCommanderID != "" {
						updateMessages = append(updateMessages, fmt.Sprintf("Incident Commander changed to <@%s>.", newCommanderID))
					} else {
						updateMessages = append(updateMessages, "Incident Commander removed.")
					}
				}
				if newCommsRepID != oldCommsRepID {
					if newCommsRepID != "" {
						updateMessages = append(updateMessages, fmt.Sprintf("Comms Representative changed to <@%s>.", newCommsRepID))
					} else {
						updateMessages = append(updateMessages, "Comms Representative removed.")
					}
				}
				timelineMessage := strings.Join(updateMessages, " ")

				if newSeverity == SeveritySev1 || newSeverity == SeveritySev2 {
					err = incidentService.AddActionItem(
						ctx,
						channelID,
						interaction.User.Name,
						"Create incident postmortem",
					)
					if err != nil {
						slog.WarnContext(ctx, "Failed to add postmortem action item", "error", err)
					}
				}

				err = incidentService.AddTimelineItem(
					ctx,
					channelID,
					interaction.User.Name,
					timelineMessage, // Use the consolidated timeline message
				)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to add timeline item", "error", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to add timeline item",
						"details": err.Error(),
					})
					return
				}
			}
		}

		c.Status(http.StatusOK)
	}
}
