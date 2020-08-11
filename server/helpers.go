package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

func getIndex(command string, givenArray []ScheduledMessage) (int, *model.CommandResponse) {
	commandFields := strings.Fields(command)

	for _, field := range commandFields {
		index, err := strconv.Atoi(field)
		if err != nil {
			continue //the field we got is not a valid index, let's check the next fields...
		}
		if (len(givenArray) <= index) || (index < 0) {
			return -1, &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         fmt.Sprintf("Error: Your given index of %d is not valid", index),
			}
		}
		return index, nil
	}

	return -1, &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Error: Please enter a valid index",
	}
}

func (p *Plugin) postMessage(msg ScheduledMessage) *model.CommandResponse {
	post := &model.Post{
		ChannelId: msg.ChannelID,
		RootId:    msg.TeamID,
		UserId:    msg.Creator,
		Message:   msg.Message,
	}

	//TODO: This posts every given text as simple text, even when the text should be a command like `/topic Test123`. How can I post a command?
	if _, err := p.API.CreatePost(post); err != nil {
		const errorMessage = "Error: Failed to create scheduled post"
		p.API.LogError(errorMessage, "err", err.Error())
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         errorMessage,
		}
	}

	return &model.CommandResponse{}
}
