package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type HookyAppHandler struct {
	bindAddress string
	bindPort    int
	apps        map[string]*App
}

type CrashHandler interface {
	HandleCrash(app *App, notification HockeyNotification) error
}

type App struct {
	name           string
	hockeyAppId    string
	hockeyApiToken string
	crashHandlers  []CrashHandler
}

func (handler *HookyAppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var err error
	var bytes []byte
	if bytes, err = ioutil.ReadAll(r.Body); err != nil {
		fmt.Println("Failed to read request body", err)
		return
	}

	var notification HockeyNotification
	if err = json.Unmarshal(bytes, &notification); err != nil {
		log.Println("Got errors parsing notification", err)
		return
	}

	app := handler.apps[notification.PublicIdentifier]
	if app == nil {
		log.Println("No app configured for", notification.PublicIdentifier)
		return
	}

	if notification.Type == HOCKEY_NOTIFICATION_CRASH_REASON {
		crashes, err := app.GetCrashes(notification.CrashReason)
		if err != nil {
			log.Println("Failed to fetch crashes for crash reason", err)
			return
		}

		notification.CrashReason.Crashes = crashes

		for _, handler := range app.crashHandlers {
			if err = handler.HandleCrash(app, notification); err != nil {
				log.Println("Error when handling crash", err)
			} else {
				log.Println("Successfully handled crash reason")
			}
		}
	}
}

func main() {
	var err error
	var handler *HookyAppHandler
	if handler, err = ParseConfig("hookyapp.toml"); err != nil {
		log.Fatalln("Failed to initialize from config", err)
		return
	}

	http.Handle("/hockeyapp_webhook", handler)

	log.Printf("Listening on %s:%d\n", handler.bindAddress, handler.bindPort)
	http.ListenAndServe(fmt.Sprintf("%s:%d", handler.bindAddress, handler.bindPort), nil)
}
