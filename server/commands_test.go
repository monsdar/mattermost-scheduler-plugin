package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveSchedule_fail(t *testing.T) {
	t.Run("No index given", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Error: Please enter a valid index", result.Text)
	})
	t.Run("No messages", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove 0",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Error: Your given index of 0 is not valid", result.Text)
	})
	t.Run("Negative Index", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{},
			ScheduledMessage{},
			ScheduledMessage{},
		}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove -1",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Error: Your given index of -1 is not valid", result.Text)
	})
	t.Run("Index out of bounds", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{},
			ScheduledMessage{},
			ScheduledMessage{},
		}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove 3",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Error: Your given index of 3 is not valid", result.Text)
	})
}
func TestRemoveSchedule_success(t *testing.T) {
	t.Run("Remove message", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{
				Message: "Index 0",
			},
			ScheduledMessage{
				Message: "Index 1",
			},
			ScheduledMessage{
				Message: "Index 2",
			},
		}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		schedulerDataAfter := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{
				Message: "Index 0",
			},
			ScheduledMessage{
				Message: "Index 2",
			},
		}}
		reqBodyBytesAfter := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytesAfter).Encode(schedulerDataAfter)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		api.On("KVSet", mock.AnythingOfType("string"), reqBodyBytesAfter.Bytes()).Return(nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove 1",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Scheduled messages removed", result.Text)
	})
	t.Run("Remove message with some random string in the middle of the command", func(t *testing.T) {
		schedulerData := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{
				Message: "Index 0",
			},
			ScheduledMessage{
				Message: "Index 1",
			},
			ScheduledMessage{
				Message: "Index 2",
			},
		}}
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(schedulerData)

		schedulerDataAfter := &SchedulerData{ScheduledMessages: []ScheduledMessage{
			ScheduledMessage{
				Message: "Index 0",
			},
			ScheduledMessage{
				Message: "Index 2",
			},
		}}
		reqBodyBytesAfter := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytesAfter).Encode(schedulerDataAfter)

		plugin := &Plugin{}
		api := &plugintest.API{}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{Username: "TestUser"}, nil)
		api.On("KVGet", mock.AnythingOfType("string")).Return(reqBodyBytes.Bytes(), nil)
		api.On("KVSet", mock.AnythingOfType("string"), reqBodyBytesAfter.Bytes()).Return(nil)
		plugin.SetAPI(api)

		args := &model.CommandArgs{
			Command:   "/scheduler remove yeah 1",
			ChannelId: "TestChannel",
			TeamId:    "TestTeam",
			UserId:    "TestUser",
		}

		result, _ := plugin.ExecuteCommand(nil, args)
		assert.Equal(t, "Scheduled messages removed", result.Text)
	})
}
