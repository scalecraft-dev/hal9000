package incident

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func CreateTimeline(slackApi *slack.Client, channelID string, userName string, severity string, status string) error {
	headerText := slack.NewTextBlockObject("mrkdwn", "*Incident Timeline*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	timelineText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s - Incident created by <@%s>. Severity: %s Status: %s", time.Now().UTC().Format("2006-01-02 15:04:05"), userName, severity, status), false, false)
	timelineSection := slack.NewSectionBlock(timelineText, nil, nil)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			timelineSection,
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

func AddTimelineItem(slackApi *slack.Client, channelID string, userName string, timelineText string) error {

	timelineItem, err := GetTimelineMessage(slackApi, channelID)
	if err != nil {
		return err
	}

	additionalText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s - %s by <@%s>", time.Now().UTC().Format("2006-01-02 15:04:05"), timelineText, userName), false, false)
	additionalSection := slack.NewSectionBlock(additionalText, nil, nil)

	timelineItem.Message.Blocks.BlockSet = append(timelineItem.Message.Blocks.BlockSet, additionalSection)
	_, _, _, err = slackApi.UpdateMessage(
		channelID,
		timelineItem.Message.Timestamp,
		slack.MsgOptionBlocks(timelineItem.Message.Blocks.BlockSet...),
		slack.MsgOptionAsUser(true),
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetTimelineMessage(slackApi *slack.Client, channelID string) (*slack.Item, error) {
	items, _, err := slackApi.ListPins(channelID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		log.Printf("no pins found in channel %s", channelID)
	}

	for _, item := range items {
		if strings.Contains(item.Message.Text, "*Incident Timeline*") {
			return &item, nil
		}
	}

	return nil, nil
}
