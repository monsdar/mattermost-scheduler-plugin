package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
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
		Text:         "This plugin schedules messages. Add a new one by calling `/scheduler <cron>:<message>`",
	}
}


TODO: Unittest um Remove ab ExecuteCommand zu checken... Scheint noch nicht zu funktionieren -.-
func (p *Plugin) executeCommandSchedulerRemove(args *model.CommandArgs) *model.CommandResponse {
	data := p.ReadFromStorage()

	indeces, errResponse := getIndeces(args.Command, data.ScheduledMessages)
	if errResponse != nil {
		return errResponse
	}

	for _, index := range indeces {
		//from https://stackoverflow.com/a/37335777/199513
		data.ScheduledMessages = append(data.ScheduledMessages[:index], data.ScheduledMessages[index+1:]...)
	}
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
	for index, scheduledMsg := range data.ScheduledMessages {
		creator := scheduledMsg.Creator
		user, err := p.API.GetUser(creator)
		if err == nil {
			creator = user.GetDisplayName("")
		}
		message = message + fmt.Sprintf("%d.\t%s (%s):\t%s\n", index, scheduledMsg.Cron, creator, scheduledMsg.Message)
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         message,
	}
}

TODO: Scheint als wenn Commands nur als Text gepostet werden - sollen aber ja excuted werden als wenn der User die schreibt
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

	//check if the cron-syntax is valid
	_, err := cron.Parse(newMessage.Cron)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Error: Your cron-syntax does not seem to be correct",
		}
	}

	p.pluginCron.AddFunc(newMessage.Cron, func() { p.postMessage(newMessage) })

	data := p.ReadFromStorage()
	data.ScheduledMessages = append(data.ScheduledMessages, newMessage)
	p.WriteToStorage(&data)

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "Added your message!",
	}
}
