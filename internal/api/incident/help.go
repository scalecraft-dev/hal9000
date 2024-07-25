package incident

import "github.com/slack-go/slack"

func HelpMessage() *slack.Blocks {
	introText := slack.NewTextBlockObject("mrkdwn", "Hey there 👋 I'm Hal. I'm here to help you create and manage incidents in Slack.\nHere are the commands available to you:", false, false)
	introSection := slack.NewSectionBlock(introText, nil, nil)

	createText := slack.NewTextBlockObject("mrkdwn", "*🆕 Use `/incident create`*. I will ask you for some details, and create a new incident.", false, false)
	createSection := slack.NewSectionBlock(createText, nil, nil)

	updateText := slack.NewTextBlockObject("mrkdwn", "*𝌡 Use `/incident update`*. Change the status or severity of an incident.", false, false)
	updateSection := slack.NewSectionBlock(updateText, nil, nil)

	actionItemText := slack.NewTextBlockObject("mrkdwn", "*🧹 Use `/incident action-item <description of the action-item>`*. Adds an action item to the incident.", false, false)
	actionItemSection := slack.NewSectionBlock(actionItemText, nil, nil)

	timelineText := slack.NewTextBlockObject("mrkdwn", "*⏰ Use `/incident timeline <timeline update>`*. Adds an event to the incident timeline.", false, false)
	timelineSection := slack.NewSectionBlock(timelineText, nil, nil)

	helpText := slack.NewTextBlockObject("mrkdwn", "*🤖 Use `/incident help` command*. Show this menu again.", false, false)
	helpSection := slack.NewSectionBlock(helpText, nil, nil)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			introSection,
			createSection,
			updateSection,
			actionItemSection,
			timelineSection,
			helpSection,
		},
	}

	return &blocks
}
