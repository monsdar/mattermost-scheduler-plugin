package main

import (
	"bytes"
	"encoding/json"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	//KVKEY is the key used for storing the data in the KVStorage
	KVKEY = "SchedulerData"
)

// ReadFromStorage reads IceBreakerData from the KVStore. Makes sure that data is inited for the given team and channel
func (p *Plugin) ReadFromStorage() SchedulerData {
	data := SchedulerData{}
	kvData, err := p.API.KVGet(KVKEY)
	if err != nil {
		//do nothing.. we'll return an empty IceBreakerData then...
	}
	if kvData != nil {
		json.Unmarshal(kvData, &data)
	}

	return data
}

// WriteToStorage writes the given data to storage
func (p *Plugin) WriteToStorage(data *SchedulerData) {
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(data)
	p.API.KVSet(KVKEY, reqBodyBytes.Bytes())
}

// ClearStorage removes all stored data from KVStorage
func (p *Plugin) ClearStorage() *model.AppError {
	return p.API.KVDelete(KVKEY)
}
