package main

import (
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	//This is our cron-instance, triggering the right messages at the right time
	pluginCron *cron.Cron
}

//ScheduledMessage stores information about a message that has been scheduled with the plugin
type ScheduledMessage struct {
	Creator   string       `json:"creator"` //userID of the author
	TeamID    string       `json:"teamID"`
	ChannelID string       `json:"channelID"`
	Cron      string       `json:"cron"`
	Message   string       `json:"message"`
	CronID    cron.EntryID `json:"CronID"` //needed so we know which cron-job we need to stop
}

//SchedulerData contains all data necessary to be stored for the Scheduler Plugin
type SchedulerData struct {
	ScheduledMessages []ScheduledMessage `json:"ScheduledMessage"`
}

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate() error {
	//register all our commands
	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	if p.pluginCron != nil {
		p.pluginCron.Stop()
	}
	data := p.ReadFromStorage()
	p.pluginCron = cron.New(cron.WithSeconds())
	for index := range data.ScheduledMessages {
		entryID, err := p.pluginCron.AddFunc(data.ScheduledMessages[index].Cron, func() { p.postMessage(data.ScheduledMessages[index]) })
		if err == nil {
			data.ScheduledMessages[index].CronID = entryID
		}
	}
	p.WriteToStorage(&data)
	p.pluginCron.Start()

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	data := p.ReadFromStorage()
	for _, msg := range data.ScheduledMessages {
		p.pluginCron.Remove(msg.CronID)
		msg.CronID = -1
	}
	p.WriteToStorage(&data)
	p.pluginCron.Stop()
	p.pluginCron = nil
	return nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
