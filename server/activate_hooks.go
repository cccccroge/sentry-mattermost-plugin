package main

import (
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

func (p *Plugin) OnActivate() error {
	if p.client == nil {
		p.client = pluginapi.NewClient(p.API, p.Driver)
	}

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	p.initializeAPI()

	return nil
}
