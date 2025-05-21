package internal

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type IncidentService struct {
	slackService *SlackService
	config       *Config
}

func NewIncidentService(slackService *SlackService, config *Config) *IncidentService {
	return &IncidentService{
		slackService: slackService,
		config:       config,
	}
}

func (s *IncidentService) CreateIncidentChannel(ctx context.Context, description string, severity Severity, status Status, incidentCommanderID string, commsRepresentativeID string, userIDs ...string) (*slack.Channel, error) {
	channelInt := 1
	var channel *slack.Channel
	var err error

	// Ensure commander and comms rep are in the userIDs to be invited, if they are set
	// The main userIDs array already includes the person who initiated the command + any from the multi-select
	activeUserIDs := userIDs
	if incidentCommanderID != "" {
		activeUserIDs = appendIfMissing(activeUserIDs, incidentCommanderID)
	}
	if commsRepresentativeID != "" {
		activeUserIDs = appendIfMissing(activeUserIDs, commsRepresentativeID)
	}

	for channelInt < 10 {
		channelName := "incident-" + time.Now().Format("20060102") + "-" + strconv.Itoa(channelInt)
		channel, err = s.slackService.CreateChannel(ctx, channelName, false)

		if err != nil {
			if strings.Contains(err.Error(), "name_taken") {
				channelInt++
				continue
			}
			return nil, fmt.Errorf("failed to create incident channel: %w", err)
		}

		if len(activeUserIDs) > 0 {
			err = s.slackService.InviteUsersToChannel(ctx, channel.ID, activeUserIDs...)
			if err != nil {
				return nil, fmt.Errorf("failed to invite users to incident channel: %w", err)
			}
		}

		// Construct topic string
		topicParts := []string{fmt.Sprintf("%s incident: %s", severity, description)}
		if incidentCommanderID != "" {
			topicParts = append(topicParts, fmt.Sprintf("Commander: <@%s>", incidentCommanderID))
		}
		if commsRepresentativeID != "" {
			topicParts = append(topicParts, fmt.Sprintf("Comms: <@%s>", commsRepresentativeID))
		}
		channelTopic := strings.Join(topicParts, " | ")

		err = s.slackService.SetChannelTopic(ctx, channel.ID, channelTopic)
		if err != nil {
			slog.WarnContext(ctx, "Failed to set channel topic (non-fatal)", "channelID", channel.ID, "error", err)
		}

		break
	}

	return channel, nil
}

// Helper function (can be defined at package level)
func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func (s *IncidentService) CreateTimeline(ctx context.Context, channelID, userName string, severity Severity, status Status) error {
	headerText := slack.NewTextBlockObject("mrkdwn", "*Incident Timeline*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	timelineText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s - Incident created by <@%s>. Severity: %s Status: %s",
		time.Now().UTC().Format("2006-01-02 15:04:05"), userName, severity, status), false, false)
	timelineSection := slack.NewSectionBlock(timelineText, nil, nil)

	blocks := []slack.Block{
		headerSection,
		timelineSection,
	}

	timestamp, err := s.slackService.PostMessage(ctx, channelID, blocks)
	if err != nil {
		return fmt.Errorf("failed to create timeline: %w", err)
	}

	err = s.slackService.AddPin(ctx, channelID, timestamp)
	if err != nil {
		return fmt.Errorf("failed to pin timeline message: %w", err)
	}

	return nil
}

func (s *IncidentService) AddTimelineItem(ctx context.Context, channelID, userName, message string) error {
	timelineItem, err := s.getTimelineMessage(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get timeline message: %w", err)
	}

	if timelineItem == nil {
		return fmt.Errorf("timeline message not found")
	}

	additionalText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s - %s by <@%s>",
		time.Now().UTC().Format("2006-01-02 15:04:05"), message, userName), false, false)
	additionalSection := slack.NewSectionBlock(additionalText, nil, nil)

	updatedBlocks := append(timelineItem.Message.Blocks.BlockSet, additionalSection)

	err = s.slackService.UpdateMessage(ctx, channelID, timelineItem.Message.Timestamp, updatedBlocks)
	if err != nil {
		return fmt.Errorf("failed to update timeline: %w", err)
	}

	return nil
}

func (s *IncidentService) getTimelineMessage(ctx context.Context, channelID string) (*slack.Item, error) {
	items, err := s.slackService.ListPins(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		slog.WarnContext(ctx, "No pins found in channel", "channelID", channelID)
		return nil, nil
	}

	for _, item := range items {
		if strings.Contains(item.Message.Text, "*Incident Timeline*") {
			return &item, nil
		}
	}

	return nil, nil
}

func (s *IncidentService) CreateActionItems(ctx context.Context, channelID, userName string) error {
	headerText := slack.NewTextBlockObject("mrkdwn", "*Action Items*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks := []slack.Block{
		headerSection,
	}

	timestamp, err := s.slackService.PostMessage(ctx, channelID, blocks)
	if err != nil {
		return fmt.Errorf("failed to create action items: %w", err)
	}

	err = s.slackService.AddPin(ctx, channelID, timestamp)
	if err != nil {
		return fmt.Errorf("failed to pin action items message: %w", err)
	}

	return nil
}

func (s *IncidentService) AddActionItem(ctx context.Context, channelID, userName, description string) error {
	actionItem, err := s.getActionItemMessage(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to get action items message: %w", err)
	}

	if actionItem == nil {
		return fmt.Errorf("action items message not found")
	}

	if strings.Contains(actionItem.Message.Text, description) {
		return nil
	}

	additionalText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s> - %s", userName, description), false, false)
	additionalSection := slack.NewSectionBlock(additionalText, nil, nil)

	updatedBlocks := append(actionItem.Message.Blocks.BlockSet, additionalSection)

	err = s.slackService.UpdateMessage(ctx, channelID, actionItem.Message.Timestamp, updatedBlocks)
	if err != nil {
		return fmt.Errorf("failed to update action items: %w", err)
	}

	return nil
}

func (s *IncidentService) getActionItemMessage(ctx context.Context, channelID string) (*slack.Item, error) {
	items, err := s.slackService.ListPins(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		slog.WarnContext(ctx, "No pins found in channel", "channelID", channelID)
		return nil, nil
	}

	for _, item := range items {
		if strings.Contains(item.Message.Text, "*Action Items*") {
			return &item, nil
		}
	}

	return nil, nil
}

func (s *IncidentService) CreateIncidentModal() slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject("plain_text", "Create an Incident", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Cancel", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Create Incident", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Enter the below details for the incident channel.", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	statusText := slack.NewTextBlockObject("plain_text", "status", false, false)
	statusPlaceholder := slack.NewTextBlockObject("plain_text", "Current Status...", false, false)
	statusOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "Investigating", false, false), slack.NewTextBlockObject("plain_text", "Incident is under investigation.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "Fixing", false, false), slack.NewTextBlockObject("plain_text", "A fix is being implemented.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Monitoring", false, false), slack.NewTextBlockObject("plain_text", "Fix implemented, monitoring for stability.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "Resolved", false, false), slack.NewTextBlockObject("plain_text", "The incident has been resolved.", false, false)),
	}
	statusSelection := slack.NewOptionsSelectBlockElement("static_select", statusPlaceholder, "status", statusOptions...)
	status := slack.NewInputBlock("status", statusText, nil, statusSelection)

	severityText := slack.NewTextBlockObject("plain_text", "severity", false, false)
	severityPlaceholder := slack.NewTextBlockObject("plain_text", "Select Severity...", false, false)
	severityOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "SEV-0", false, false), slack.NewTextBlockObject("plain_text", "Critical: High impact, affects many users.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "SEV-1", false, false), slack.NewTextBlockObject("plain_text", "Major: Significant impact, affects some users.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "SEV-2", false, false), slack.NewTextBlockObject("plain_text", "Moderate: Non-critical impact or user inconvenience.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "SEV-3", false, false), slack.NewTextBlockObject("plain_text", "Minor: Low priority, no immediate attention needed.", false, false)),
	}
	severitySelection := slack.NewOptionsSelectBlockElement("static_select", severityPlaceholder, "incident_severity", severityOptions...)
	severity := slack.NewInputBlock("incident_severity", severityText, nil, severitySelection)

	descriptionText := slack.NewTextBlockObject("plain_text", "description", false, false)
	descriptionHint := slack.NewTextBlockObject("plain_text", "Example: Increased latency in Carehub", false, false)
	descriptionPlaceholder := slack.NewTextBlockObject("plain_text", "Enter description of incident...", false, false)
	descriptionElement := slack.NewPlainTextInputBlockElement(descriptionPlaceholder, "description")
	description := slack.NewInputBlock("description", descriptionText, descriptionHint, descriptionElement)

	// Incident Commander field
	commanderText := slack.NewTextBlockObject("plain_text", "Incident Commander (Optional - select one)", false, false)
	commanderPlaceholder := slack.NewTextBlockObject("plain_text", "Select Incident Commander", false, false)
	commanderElement := &slack.SelectBlockElement{
		Type:        "users_select",
		Placeholder: commanderPlaceholder,
		ActionID:    "incident_commander",
	}
	commanderInput := slack.NewInputBlock("incident_commander", commanderText, nil, commanderElement)
	commanderInput.Optional = true

	// Comms Representative field
	commsRepText := slack.NewTextBlockObject("plain_text", "Comms Representative (Optional - select one)", false, false)
	commsRepPlaceholder := slack.NewTextBlockObject("plain_text", "Select Comms Representative", false, false)
	commsRepElement := &slack.SelectBlockElement{
		Type:        "users_select",
		Placeholder: commsRepPlaceholder,
		ActionID:    "comms_representative",
	}
	commsRepInput := slack.NewInputBlock("comms_representative", commsRepText, nil, commsRepElement)
	commsRepInput.Optional = true

	membersText := slack.NewTextBlockObject("plain_text", "incident_members", false, false)
	membersHint := slack.NewTextBlockObject("plain_text", "Example: @johndoe @janedoe", false, false)
	membersPlaceholder := slack.NewTextBlockObject("plain_text", "Enter people to add to incident channel...", false, false)
	membersSelection := slack.NewOptionsMultiSelectBlockElement(slack.MultiOptTypeUser, membersPlaceholder, "incident_members")
	members := slack.NewInputBlock("incident_members", membersText, membersHint, membersSelection)
	members.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			severity,
			status,
			description,
			commanderInput,
			commsRepInput,
			members,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.CallbackID = "create_incident_modal"

	return modalRequest
}

// UpdateIncidentModal now needs channelID to pre-fill commander/comms rep
func (s *IncidentService) UpdateIncidentModal(ctx context.Context, channelID string) slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject("plain_text", "Update an Incident", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Cancel", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Update Incident", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Enter the below details for the incident update.", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Status and Severity fields (as before)
	statusText := slack.NewTextBlockObject("plain_text", "status", false, false)
	statusPlaceholder := slack.NewTextBlockObject("plain_text", "Current Status...", false, false)
	statusOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "Investigating", false, false), slack.NewTextBlockObject("plain_text", "Incident is under investigation.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "Fixing", false, false), slack.NewTextBlockObject("plain_text", "A fix is being implemented.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "Monitoring", false, false), slack.NewTextBlockObject("plain_text", "Fix implemented, monitoring for stability.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "Resolved", false, false), slack.NewTextBlockObject("plain_text", "The incident has been resolved.", false, false)),
	}
	statusSelection := slack.NewOptionsSelectBlockElement("static_select", statusPlaceholder, "status", statusOptions...)
	status := slack.NewInputBlock("status", statusText, nil, statusSelection)

	severityText := slack.NewTextBlockObject("plain_text", "severity", false, false)
	severityPlaceholder := slack.NewTextBlockObject("plain_text", "Select Severity...", false, false)
	severityOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("1", slack.NewTextBlockObject("plain_text", "SEV-0", false, false), slack.NewTextBlockObject("plain_text", "Critical: High impact, affects many users.", false, false)),
		slack.NewOptionBlockObject("2", slack.NewTextBlockObject("plain_text", "SEV-1", false, false), slack.NewTextBlockObject("plain_text", "Major: Significant impact, affects some users.", false, false)),
		slack.NewOptionBlockObject("3", slack.NewTextBlockObject("plain_text", "SEV-2", false, false), slack.NewTextBlockObject("plain_text", "Moderate: Non-critical impact or user inconvenience.", false, false)),
		slack.NewOptionBlockObject("4", slack.NewTextBlockObject("plain_text", "SEV-3", false, false), slack.NewTextBlockObject("plain_text", "Minor: Low priority, no immediate attention needed.", false, false)),
	}
	severitySelection := slack.NewOptionsSelectBlockElement("static_select", severityPlaceholder, "incident_severity", severityOptions...)
	severity := slack.NewInputBlock("incident_severity", severityText, nil, severitySelection)

	// Fetch current topic to extract commander and comms rep for pre-filling
	currentCommanderID, currentCommsRepID := "", ""
	currentTopic, err := s.slackService.GetChannelTopic(ctx, channelID)
	if err == nil {
		if idx := strings.Index(currentTopic, " | Commander: <@"); idx != -1 {
			start := idx + len(" | Commander: <@")
			end := strings.Index(currentTopic[start:], ">")
			if end != -1 {
				currentCommanderID = currentTopic[start : start+end]
			}
		}
		if idx := strings.Index(currentTopic, " | Comms: <@"); idx != -1 {
			start := idx + len(" | Comms: <@")
			end := strings.Index(currentTopic[start:], ">")
			if end != -1 {
				currentCommsRepID = currentTopic[start : start+end]
			}
		}
	} else {
		slog.WarnContext(ctx, "Could not get current topic for UpdateIncidentModal prefill", "channelID", channelID, "error", err)
	}

	// Incident Commander field (optional)
	commanderText := slack.NewTextBlockObject("plain_text", "Incident Commander (Optional - select one)", false, false)
	commanderPlaceholder := slack.NewTextBlockObject("plain_text", "Select Incident Commander", false, false)
	commanderElement := &slack.SelectBlockElement{
		Type: "users_select", Placeholder: commanderPlaceholder, ActionID: "incident_commander",
	}
	if currentCommanderID != "" {
		commanderElement.InitialUser = currentCommanderID
	}
	commanderInput := slack.NewInputBlock("incident_commander", commanderText, nil, commanderElement)
	commanderInput.Optional = true

	// Comms Representative field (optional)
	commsRepText := slack.NewTextBlockObject("plain_text", "Comms Representative (Optional - select one)", false, false)
	commsRepPlaceholder := slack.NewTextBlockObject("plain_text", "Select Comms Representative", false, false)
	commsRepElement := &slack.SelectBlockElement{
		Type: "users_select", Placeholder: commsRepPlaceholder, ActionID: "comms_representative",
	}
	if currentCommsRepID != "" {
		commsRepElement.InitialUser = currentCommsRepID
	}
	commsRepInput := slack.NewInputBlock("comms_representative", commsRepText, nil, commsRepElement)
	commsRepInput.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			severity,
			status,
			commanderInput,
			commsRepInput,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.CallbackID = "update_incident_modal"
	modalRequest.PrivateMetadata = channelID // Keep passing channelID for submission handler

	return modalRequest
}

func (s *IncidentService) HelpMessage() *slack.Blocks {
	introText := slack.NewTextBlockObject("mrkdwn", "Hey there 👋 I'm Hal. I'm here to help you create and manage incidents in Slack.\nHere are the commands available to you:", false, false)
	introSection := slack.NewSectionBlock(introText, nil, nil)

	createText := slack.NewTextBlockObject("mrkdwn", "*🆕 Use `/incident create` (or `c`)*. I will ask you for some details, and create a new incident.", false, false)
	createSection := slack.NewSectionBlock(createText, nil, nil)

	updateText := slack.NewTextBlockObject("mrkdwn", "*🔄 Use `/incident update` (or `u`)*. Change the status or severity of an incident.", false, false)
	updateSection := slack.NewSectionBlock(updateText, nil, nil)

	actionItemText := slack.NewTextBlockObject("mrkdwn", "*🧹 Use `/incident action-item <description>` (or `ai <description>`)*. Adds an action item to the incident.", false, false)
	actionItemSection := slack.NewSectionBlock(actionItemText, nil, nil)

	timelineText := slack.NewTextBlockObject("mrkdwn", "*⏰ Use `/incident timeline <message>` (or `t <message>`)*. Adds an event to the incident timeline.", false, false)
	timelineSection := slack.NewSectionBlock(timelineText, nil, nil)

	resolveText := slack.NewTextBlockObject("mrkdwn", "*✅ Use `/incident resolve [optional message]` (or `r [optional message]`)*. Marks the incident as resolved and updates the channel topic.", false, false)
	resolveSection := slack.NewSectionBlock(resolveText, nil, nil)

	helpText := slack.NewTextBlockObject("mrkdwn", "*🤖 Use `/incident help` (or `h`)*. Show this menu again.", false, false)
	helpSection := slack.NewSectionBlock(helpText, nil, nil)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			introSection,
			createSection,
			updateSection,
			actionItemSection,
			timelineSection,
			resolveSection,
			helpSection,
		},
	}

	return &blocks
}

func (s *IncidentService) CreateIncident(ctx context.Context, triggerID string) error {
	modal := s.CreateIncidentModal()
	return s.slackService.OpenView(ctx, triggerID, modal)
}

func (s *IncidentService) UpdateIncident(ctx context.Context, triggerID, channelID string) error {
	// Pass channelID to UpdateIncidentModal for pre-filling purposes
	modal := s.UpdateIncidentModal(ctx, channelID)
	// PrivateMetadata is still set here, which is fine and used by the submission handler.
	modal.PrivateMetadata = channelID
	return s.slackService.OpenView(ctx, triggerID, modal)
}

func (s *IncidentService) ResolveIncident(ctx context.Context, channelID, userID, userName, resolutionMessage string) error {
	// Add to timeline
	timelineMsg := fmt.Sprintf("Incident resolved by <@%s>.", userName)
	if resolutionMessage != "" {
		timelineMsg = fmt.Sprintf("Incident resolved by <@%s>. Resolution: %s", userName, resolutionMessage)
	}
	err := s.AddTimelineItem(ctx, channelID, userID, timelineMsg)
	if err != nil {
		// Log the error but don't block topic update if timeline fails for some reason
		slog.ErrorContext(ctx, "Failed to add resolved item to timeline", "channelID", channelID, "error", err)
	}

	// Update channel topic
	currentTopic, err := s.slackService.GetChannelTopic(ctx, channelID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get channel topic for resolving incident", "channelID", channelID, "error", err)
		// Proceed without updating topic if fetching fails
	} else {
		originalDescription := ""
		commanderPart := ""
		commsPart := ""

		// Topic format: "<SEV-X/Resolved> incident: <description> [| Commander: <@...>] [| Comms: <@...>]"
		// Extract original description first
		coreTopic := currentTopic
		if idx := strings.Index(coreTopic, " | Commander:"); idx != -1 {
			coreTopic = coreTopic[:idx]
		}
		if idx := strings.Index(coreTopic, " | Comms:"); idx != -1 {
			coreTopic = coreTopic[:idx]
		}

		if strings.Contains(coreTopic, " incident: ") {
			parts := strings.SplitN(coreTopic, " incident: ", 2)
			if len(parts) == 2 {
				originalDescription = parts[1]
			}
		} else if strings.HasPrefix(coreTopic, "Resolved: ") { // If already resolved, description is after "Resolved: "
			originalDescription = strings.TrimPrefix(coreTopic, "Resolved: ")
		} else if coreTopic != "" { // Fallback for other formats
			originalDescription = coreTopic
			slog.WarnContext(ctx, "Could not parse original description from topic for resolve, using full core topic", "channelID", channelID, "coreTopic", coreTopic)
		} else {
			originalDescription = "Description unavailable"
			slog.WarnContext(ctx, "Original topic was empty or unparsable for resolve", "channelID", channelID)
		}

		// Preserve commander and comms if they exist in the original topic
		if idx := strings.Index(currentTopic, " | Commander: "); idx != -1 {
			tempPart := currentTopic[idx+len(" | Commander: "):]
			if commsIdx := strings.Index(tempPart, " | Comms: "); commsIdx != -1 {
				commanderPart = fmt.Sprintf(" | Commander: %s", tempPart[:commsIdx])
			} else {
				commanderPart = fmt.Sprintf(" | Commander: %s", tempPart)
			}
		}
		if idx := strings.Index(currentTopic, " | Comms: "); idx != -1 {
			commsPart = fmt.Sprintf(" | Comms: %s", currentTopic[idx+len(" | Comms: "):])
		}

		newTopic := fmt.Sprintf("Resolved: %s%s%s", originalDescription, commanderPart, commsPart)
		err = s.slackService.SetChannelTopic(ctx, channelID, newTopic)
		if err != nil {
			slog.WarnContext(ctx, "Failed to set channel topic to resolved", "channelID", channelID, "newTopic", newTopic, "error", err)
		}
	}

	// Send ephemeral confirmation
	err = s.slackService.PostEphemeralMessage(ctx, channelID, userID, []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", ":white_check_mark: Incident marked as resolved.", false, false), nil, nil),
	})
	if err != nil {
		slog.WarnContext(ctx, "Failed to send resolve confirmation message", "channelID", channelID, "userID", userID, "error", err)
	}

	return nil // Overall command success even if some non-critical parts fail (logged as warnings)
}
