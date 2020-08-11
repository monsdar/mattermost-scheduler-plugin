package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	commandScheduler       = "scheduler"
	commandSchedulerAdd    = commandScheduler + " add"
	commandSchedulerList   = commandScheduler + " list"
	commandSchedulerRemove = commandScheduler + " remove"
)

func (p *Plugin) registerCommands() error {
	commands := [...]model.Command{
		model.Command{
			Trigger:          commandScheduler,
			AutoComplete:     true,
			AutoCompleteDesc: "Display a help message",
		},
		model.Command{
			Trigger:          commandSchedulerAdd,
			AutoComplete:     true,
			AutoCompleteHint: "<cron>: <message>",
			AutoCompleteDesc: "Add a new scheduled message",
		},
		model.Command{
			Trigger:          commandSchedulerList,
			AutoComplete:     true,
			AutoCompleteDesc: "List all the schedules that have been made",
		},
		model.Command{
			Trigger:          commandSchedulerRemove,
			AutoComplete:     true,
			AutoCompleteHint: "<index>",
			AutoCompleteDesc: "Remove a scheduled message",
		},
	}

	for _, command := range commands {
		if err := p.API.RegisterCommand(&command); err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to register %s command", command.Trigger))
		}
	}

	return nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand
// API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	userCommands := map[string]func(args *model.CommandArgs) (*model.CommandResponse, *model.AppError){
		commandSchedulerList: func(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
			return p.executeCommandSchedulerList(args), nil
		},
		commandSchedulerRemove: func(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
			return p.executeCommandSchedulerRemove(args), nil
		},
		commandSchedulerAdd: func(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
			return p.executeCommandSchedulerAdd(args), nil
		},
	}

	trigger := strings.TrimPrefix(args.Command, "/")
	trigger = strings.TrimSuffix(trigger, " ")

	for key, value := range userCommands {
		if strings.HasPrefix(trigger, key) {
			return value(args)
		}
	}
	if trigger == commandScheduler {
		return p.executeCommandScheduler(args), nil
	}

	//return an error message when the command has not been detected at all
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         fmt.Sprintf("Unknown command: " + args.Command),
	}, nil
}

func (p *Plugin) executeCommandScheduler(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "This plugin schedules messages. Add a new one by calling `/scheduler add <cron>:<message>`",
	}
}

func (p *Plugin) executeCommandSchedulerRemove(args *model.CommandArgs) *model.CommandResponse {
	data := p.ReadFromStorage()

	index, errResponse := getIndex(args.Command, data.ScheduledMessages)
	if errResponse != nil {
		return errResponse
	}

	p.pluginCron.Remove(data.ScheduledMessages[index].CronID)

	//from https://stackoverflow.com/a/37335777/199513
	data.ScheduledMessages = append(data.ScheduledMessages[:index], data.ScheduledMessages[index+1:]...)
	p.WriteToStorage(&data)

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Scheduled messages removed",
	}
}

func (p *Plugin) executeCommandSchedulerList(args *model.CommandArgs) *model.CommandResponse {
	data := p.ReadFromStorage()

	if len(data.ScheduledMessages) == 0 {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "There are no scheduled messages...",
		}
	}

	message := "Scheduled Messages:\n"
	message = message + "| Index | TeamID | ChannelID | Author | Cron | Message |\n"
	message = message + "| :---- | :----- | :-------- | :----- | :--- | :------ |\n"
	for index, scheduledMsg := range data.ScheduledMessages {
		creator := scheduledMsg.Creator
		user, err := p.API.GetUser(creator)
		if err == nil {
			creator = user.GetDisplayName("")
		}
		channelName := scheduledMsg.ChannelID
		channel, err := p.API.GetChannel(channelName)
		if err == nil {
			channelName = channel.DisplayName
		}

		message = message + fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n", index, scheduledMsg.TeamID, channelName, creator, scheduledMsg.Cron, scheduledMsg.Message)
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         message,
	}
}

func (p *Plugin) executeCommandSchedulerAdd(args *model.CommandArgs) *model.CommandResponse {
	//check the user input and extract cron and message from it
	givenText := strings.TrimPrefix(args.Command, fmt.Sprintf("/%s", commandSchedulerAdd))
	givenText = strings.TrimPrefix(givenText, " ")
	fields := strings.Split(givenText, ":")

	if len(fields) < 2 {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Error: Please give your schedule message in the format <cron>: <message>",
		}
	}

	newMessage := ScheduledMessage{
		Creator:   args.UserId,
		ChannelID: args.ChannelId,
		TeamID:    args.RootId,
		Cron:      fields[0],
		Message:   fields[1],
	}

	entryID, err := p.pluginCron.AddFunc(newMessage.Cron, func() { p.postMessage(newMessage) })
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Error: Cannot start cron-job. Is your cron-syntax correct?",
		}
	}
	newMessage.CronID = entryID

	data := p.ReadFromStorage()
	data.ScheduledMessages = append(data.ScheduledMessages, newMessage)
	p.WriteToStorage(&data)

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Added your message!",
	}
}
