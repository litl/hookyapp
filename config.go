package main

import (
	"github.com/BurntSushi/toml"
)

type CrashHandlerType string

const (
	FOGBUGZ CrashHandlerType = "fogbugz"
)

type config struct {
	BindAddress string               `toml:"bind_address"`
	BindPort    int                  `toml:"bind_port"`
	AppConfigs  map[string]appConfig `toml:"apps"`
}

type appConfig struct {
	Name                string                        `toml:"name"`
	HockeyAppId         string                        `toml:"hockeyapp_id"`
	HockeyApiToken      string                        `toml:"hockeyapp_api_token"`
	CrashHandlerConfigs map[string]crashHandlerConfig `toml:"crash_handlers"`
}

type crashHandlerConfig struct {
	HandlerType   CrashHandlerType `toml:"type"`
	HandlerConfig toml.Primitive   `toml:"config"`
}

func (appConfig *appConfig) buildApp() (*App, error) {
	crashHandlers := make([]CrashHandler, 0)

	for _, crashHandlerConfig := range appConfig.CrashHandlerConfigs {
		switch crashHandlerConfig.HandlerType {
		case FOGBUGZ:
			crashHandler, err := NewFogbuzCrashHandler(crashHandlerConfig.HandlerConfig)
			if err != nil {
				return nil, err
			}

			crashHandlers = append(crashHandlers, crashHandler)
		}
	}

	return &App{appConfig.Name, appConfig.HockeyAppId, appConfig.HockeyApiToken, crashHandlers}, nil
}

func ParseConfig(configFile string) (*HookyAppHandler, error) {
	var config config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, err
	}

	apps := make(map[string]*App)
	for _, appConfig := range config.AppConfigs {
		app, err := appConfig.buildApp()
		if err != nil {
			return nil, err
		}

		apps[appConfig.HockeyAppId] = app
	}

	return &HookyAppHandler{config.BindAddress, config.BindPort, apps}, nil
}
