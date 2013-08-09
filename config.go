package main

import (
	"github.com/BurntSushi/toml"
)

type CrashHandlerType string

const (
	FOGBUGZ CrashHandlerType = "fogbugz"
)

type ReleaseHandlerType string

const (
	EMAIL ReleaseHandlerType = "email"
)

type config struct {
	BindAddress string               `toml:"bind_address"`
	BindPort    int                  `toml:"bind_port"`
	AppConfigs  map[string]appConfig `toml:"apps"`
}

type appConfig struct {
	Name                  string                          `toml:"name"`
	HockeyAppId           string                          `toml:"hockeyapp_id"`
	HockeyApiToken        string                          `toml:"hockeyapp_api_token"`
	CrashHandlerConfigs   map[string]crashHandlerConfig   `toml:"crash_handlers"`
	ReleaseHandlerConfigs map[string]releaseHandlerConfig `toml:"release_handlers"`
}

type crashHandlerConfig struct {
	HandlerType   CrashHandlerType `toml:"type"`
	HandlerConfig toml.Primitive   `toml:"config"`
}

type releaseHandlerConfig struct {
	HandlerType   ReleaseHandlerType `toml:"type"`
	HandlerConfig toml.Primitive     `toml:"config"`
}

func (appConfig *appConfig) buildApp() (*App, error) {
	crashHandlers := make([]NotificationHandler, 0)

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

	releaseHandlers := make([]NotificationHandler, 0)

	for _, releaseHandlerConfig := range appConfig.ReleaseHandlerConfigs {
		switch releaseHandlerConfig.HandlerType {
		case EMAIL:
			releaseHandler, err := NewEmailReleaseHandler(releaseHandlerConfig.HandlerConfig)
			if err != nil {
				return nil, err
			}

			releaseHandlers = append(releaseHandlers, releaseHandler)
		}
	}

	return &App{appConfig.Name, appConfig.HockeyAppId, appConfig.HockeyApiToken, crashHandlers, releaseHandlers}, nil
}

func (handler *HookyAppHandler) ParseConfig(configFile string) error {
	var config config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return err
	}

	apps := make(map[string]*App)
	for _, appConfig := range config.AppConfigs {
		app, err := appConfig.buildApp()
		if err != nil {
			return err
		}

		apps[appConfig.HockeyAppId] = app
	}

	handler.bindAddress = config.BindAddress
	handler.bindPort = config.BindPort
	handler.apps = apps

	return nil
}
