package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/litl/hookyapp/fogbugz"
	"log"
)

type FogbugzCrashHandlerConfig struct {
	Host     string `toml:"host"`
	Email    string `toml:"email"`
	Password string `toml:"password"`
	Project  string `toml:"project"`
	Area     string `toml:"area"`
}

func (config *FogbugzCrashHandlerConfig) GetHost() string {
	return config.Host
}

func (config *FogbugzCrashHandlerConfig) GetEmail() string {
	return config.Email
}

func (config *FogbugzCrashHandlerConfig) GetPassword() string {
	return config.Password
}

type FogbugzCrashHandler struct {
	session *fogbugz.Session
	project string
	area    string
}

func NewFogbuzCrashHandler(configPrimitive toml.Primitive) (*FogbugzCrashHandler, error) {
	var config FogbugzCrashHandlerConfig
	err := toml.PrimitiveDecode(configPrimitive, &config)
	if err != nil {
		return nil, err
	}

	session, err := fogbugz.NewSession(&config)
	if err != nil {
		return nil, err
	}

	return &FogbugzCrashHandler{session, config.Project, config.Area}, nil
}

const CONTENT_PATTERN = `
See HockeyApp at %s
User: %s

%s
`

func (handler FogbugzCrashHandler) HandleCrash(app *App, notification HockeyNotification) error {
	title := fmt.Sprintf("Crash at %s:%s - %s - %s", notification.CrashReason.File,
		notification.CrashReason.Line, notification.CrashReason.Method,
		notification.CrashReason.Reason)

	var crash HockeyCrash
	var crashLog string
	var err error
	if len(notification.CrashReason.Crashes) > 0 {
		crash = notification.CrashReason.Crashes[0]
		if crash.HasLog {
			if crashLog, err = app.GetCrashLog(crash); err != nil {
				log.Println("Failed to fetch crash logs", err)
				crashLog = ""
			}
		}
	} else {
		crash = HockeyCrash{}
	}

	content := fmt.Sprintf(CONTENT_PATTERN, notification.Url,
		crash.UserString, crashLog)
	return handler.session.FileBug(handler.project, handler.area, title, content)
}
