package incident

import (
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

func CreateIncidentChannel(slackApi *slack.Client, description string, users ...string) (*slack.Channel, error) {
	channelInt := 1
	channel := &slack.Channel{}
	err := error(nil)
	for channelInt < 10 {
		channelName := "incident-" + time.Now().Format("20060102") + "-" + strconv.Itoa(channelInt)
		channel, err = slackApi.CreateConversation(slack.CreateConversationParams{
			ChannelName: channelName,
			IsPrivate:   false,
		})

		if err != nil {
			if err.Error() == "name_taken" {
				channelInt++
				continue
			}
			return nil, err
		}
		err = addChannelParticipants(slackApi, channel.ID, users...)
		if err != nil {
			return nil, err
		}

		slackApi.SetTopicOfConversation(channel.ID, description)
		break
	}
	return channel, nil
}

func addChannelParticipants(slackApi *slack.Client, channelId string, users ...string) error {
	_, err := slackApi.InviteUsersToConversation(channelId, users...)
	if err != nil {
		return err
	}
	return nil
}
