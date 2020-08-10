# Mattermost Scheduler Plugin
This plugin let's users schedule messages with cron-syntax.

Examples:
* `/scheduler 0 0 12 * * *: Hey, it's time to get lunch!`
* `/scheduler @midnight: Another day another dollar :)` 
* `/scheduler 0 0 0 25 DEC ?: Happy XMas!` 
* `/scheduler 0 0 0 1 APR ?: /kick @henning`

## Features
* Schedule any messages you want, including slash commands from other plugins
* Cron-Syntax is implemented using [Rob Figueiredos cron library](https://pkg.go.dev/github.com/robfig/cron?tab=doc)

## Contribute
This plugin is based on the [mattermost-plugin-starter-template](https://github.com/mattermost/mattermost-plugin-starter-template). See there on how to set everything up and test the plugin.

## Attributions
The icecube logo is licensed under Creative Commons: `ice cube by 23 icons from the Noun Project`
