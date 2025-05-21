package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/slack-go/slack"
)

type SlackService struct {
	client *slack.Client
	config *Config
}

func NewSlackService(client *slack.Client, config *Config) *SlackService {
	return &SlackService{
		client: client,
		config: config,
	}
}

func (s *SlackService) GetClient() *slack.Client {
	return s.client
}

func (s *SlackService) CreateChannel(ctx context.Context, name string, isPrivate bool) (*slack.Channel, error) {
	channel, err := s.client.CreateConversation(slack.CreateConversationParams{
		ChannelName: name,
		IsPrivate:   isPrivate,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create Slack channel", "error", err)
		return nil, fmt.Errorf("failed to create Slack channel: %w", err)
	}
	return channel, nil
}

func (s *SlackService) InviteUsersToChannel(ctx context.Context, channelID string, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}

	_, err := s.client.InviteUsersToConversation(channelID, userIDs...)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to invite users to Slack channel", "error", err)
		return fmt.Errorf("failed to invite users to Slack channel: %w", err)
	}
	return nil
}

func (s *SlackService) SetChannelTopic(ctx context.Context, channelID, topic string) error {
	_, err := s.client.SetTopicOfConversation(channelID, topic)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to set Slack channel topic", "error", err)
		return fmt.Errorf("failed to set Slack channel topic: %w", err)
	}
	return nil
}

func (s *SlackService) PostMessage(ctx context.Context, channelID string, blocks []slack.Block) (string, error) {
	_, timestamp, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to post message to Slack channel", "error", err)
		return "", fmt.Errorf("failed to post message to Slack channel: %w", err)
	}
	return timestamp, nil
}

func (s *SlackService) PostEphemeralMessage(ctx context.Context, channelID, userID string, blocks []slack.Block) error {
	_, err := s.client.PostEphemeral(
		channelID,
		userID,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to post ephemeral message", "error", err)
		return fmt.Errorf("failed to post ephemeral message: %w", err)
	}
	return nil
}

func (s *SlackService) UpdateMessage(ctx context.Context, channelID, timestamp string, blocks []slack.Block) error {
	_, _, _, err := s.client.UpdateMessage(
		channelID,
		timestamp,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update message", "error", err)
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

func (s *SlackService) AddPin(ctx context.Context, channelID, timestamp string) error {
	err := s.client.AddPin(channelID, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add pin", "error", err)
		return fmt.Errorf("failed to add pin: %w", err)
	}
	return nil
}

func (s *SlackService) ListPins(ctx context.Context, channelID string) ([]slack.Item, error) {
	items, _, err := s.client.ListPins(channelID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list pins", "error", err)
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}
	return items, nil
}

func (s *SlackService) GetChannelTopic(ctx context.Context, channelID string) (string, error) {
	info, err := s.client.GetConversationInfo(&slack.GetConversationInfoInput{ChannelID: channelID, IncludeLocale: false, IncludeNumMembers: false})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get conversation info", "channelID", channelID, "error", err)
		return "", fmt.Errorf("failed to get conversation info for channel %s: %w", channelID, err)
	}
	if info.Topic.Value == "" {
		slog.WarnContext(ctx, "Channel topic is not set or accessible", "channelID", channelID)
		return "", nil // Return empty string if topic value is empty
	}
	return info.Topic.Value, nil
}

func (s *SlackService) OpenView(ctx context.Context, triggerID string, view slack.ModalViewRequest) error {
	_, err := s.client.OpenView(triggerID, view)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to open modal view", "error", err)
		return fmt.Errorf("failed to open modal view: %w", err)
	}
	return nil
}

func (s *SlackService) ValidateSlackRequest(signature, timestamp, body string) bool {
	if signature == "" || timestamp == "" || body == "" {
		return false
	}
	return true
}

func (s *SlackService) HealthCheck(ctx context.Context) error {
	_, err := s.client.AuthTest()
	if err != nil {
		return errors.New("slack API is unavailable")
	}
	return nil
}
