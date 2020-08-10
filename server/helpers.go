package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

func getIndeces(command string, givenArray []ScheduledMessage) ([]int, *model.CommandResponse) {
	commandFields := strings.Fields(command)
	indeces := []int{}

	for _, field := range commandFields {
		index, err := strconv.Atoi(field)
		if err != nil {
			//do nothing... The word we got is not a valid index, but perhaps the next fits...
		}
		if len(givenArray) <= index {
			return []int{}, &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         fmt.Sprintf("Error: Your given index of %d is not valid", index),
			}
		}
		indeces = append(indeces, index)
	}

	if len(indeces) == 0 {
		return []int{}, &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Error: Please enter a valid index",
		}
	}

	return indeces, nil
}

func (p *Plugin) postMessage(msg ScheduledMessage) *model.CommandResponse {
	post := &model.Post{
		ChannelId: msg.ChannelID,
		RootId:    msg.TeamID,
		UserId:    msg.Creator,
		Message:   msg.Message,
	}

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
