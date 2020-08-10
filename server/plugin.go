package main

import (
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
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
	Creator   string `json:"creator"` //userID of the author
	TeamID    string `json:"teamID"`
	ChannelID string `json:"channelID"`
	Cron      string `json:"cron"`
	Message   string `json:"message"`
}

//SchedulerData contains all data necessary to be stored for the Scheduler Plugin
type SchedulerData struct {
	ScheduledMessages []ScheduledMessage `json:"ScheduledMessage"`
}

// OnActivate is invoked when the plugin is activated.
//
// This demo implementation logs a message to the demo channel whenever the plugin is activated.
// It also creates a demo bot account
func (p *Plugin) OnActivate() error {
	//register all our commands
	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	err := p.ClearStorage()
	if err != nil {
		return errors.Wrap(err, "failed to clear storage")
	}

	data := p.ReadFromStorage()
	p.pluginCron = cron.New()
	for _, msg := range data.ScheduledMessages {
		p.pluginCron.AddFunc(msg.Cron, func() { p.postMessage(msg) })
	}
	p.pluginCron.Start()

	return nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
