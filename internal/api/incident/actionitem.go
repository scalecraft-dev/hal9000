package incident

import (
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
)

func CreateActionItems(slackApi *slack.Client, channelID string, userName string) error {
	headerText := slack.NewTextBlockObject("mrkdwn", "*Action Items*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
		},
	}

	_, timestamp, err := slackApi.PostMessage(
		channelID,
		slack.MsgOptionBlocks(blocks.BlockSet...),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		return err
	}

	err = slackApi.AddPin(channelID, slack.ItemRef{Channel: channelID, Timestamp: timestamp})
	if err != nil {
		return err
	}

	return nil
}

func AddActionItem(slackApi *slack.Client, channelID string, userName string, actionItemText string) error {

	actionItem, err := GetActionItemMessage(slackApi, channelID)
	if err != nil {
		return err
	}

	if strings.Contains(actionItem.Message.Text, actionItemText) {
		return nil
	}

	additionalText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s> - %s", userName, actionItemText), false, false)
	additionalSection := slack.NewSectionBlock(additionalText, nil, nil)

	actionItem.Message.Blocks.BlockSet = append(actionItem.Message.Blocks.BlockSet, additionalSection)
	_, _, _, err = slackApi.UpdateMessage(
		channelID,
		actionItem.Message.Timestamp,
		slack.MsgOptionBlocks(actionItem.Message.Blocks.BlockSet...),
		slack.MsgOptionAsUser(true),
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetActionItemMessage(slackApi *slack.Client, channelID string) (*slack.Item, error) {
	items, _, err := slackApi.ListPins(channelID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		log.Printf("no pins found in channel %s", channelID)
	}

	for _, item := range items {
		if strings.Contains(item.Message.Text, "*Action Items*") {
			return &item, nil
		}
	}

	return nil, nil
}
