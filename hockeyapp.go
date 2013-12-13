package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HockeyNotificationType string

const (
	HOCKEY_NOTIFICATION_CRASH_REASON HockeyNotificationType = "crash_reason"
	HOCKEY_NOTIFICATION_APP_VERSION  HockeyNotificationType = "app_version"
)

type HockeyNotification struct {
	PublicIdentifier string                 `json:"public_identifier"`
	Type             HockeyNotificationType `json:"type"`
	SentAt           time.Time              `json:"sent_at"`
	Url              string                 `json:"url"`
	CrashReason      HockeyCrashReason      `json:"crash_reason"`
	AppVersion       HockeyAppVersion       `json:"app_version"`
}

type HockeyCrashReason struct {
	Id                 int       `json:"id"`
	AppId              int       `json:"app_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Status             int       `json:"status"`
	Reason             string    `json:"reason"`
	LastCrashAt        time.Time `json:"last_crash_at"`
	Fixed              bool      `json:"fixed"`
	AppVersionId       int       `json:"app_version_id"`
	BundleVersion      string    `json:"bundle_version"`
	BundleShortVersion string    `json:"bundle_short_version"`
	NumberOfCrashes    int       `json:"number_of_crashes"`
	Method             string    `json:"method"`
	File               string    `json:"file"`
	Class              string    `json:"class"`
	Line               string    `json:"line"`
	Crashes            []HockeyCrash
}

type HockeyCrash struct {
	Id                 int       `json:"id"`
	AppId              int       `json:"app_id"`
	BundleVersion      string    `json:"bundle_version"`
	BundleShortVersion string    `json:"bundle_short_version"`
	ContactString      string    `json:"contact_string"`
	CrashReasonId      int       `json:"crash_reason_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	HasDescription     bool      `json:"has_description"`
	HasLog             bool      `json:"has_log"`
	Model              string    `json:"model"`
	Oem                string    `json:"oem"`
	OsVersion          string    `json:"os_version"`
	UserString         string    `json:"user_string"`
}

type HockeyAppVersion struct {
	Version          string        `json:"version"`
	ShortVersion     string        `json:"shortversion"`
	Title            string        `json:"title"`
	Timestamp        int           `json:"timestamp"`
	AppSize          int           `json:"appsize"`
	Notes            template.HTML `json:"notes"`
	Mandatory        bool          `json:"mandatory"`
	MinimumOsVersion string        `json:"minimum_os_version"`
	DeviceFamily     *string       `json:"device_family"`
	Id               int           `json:"id"`
	AppId            int           `json:"app_id"`
	RestrictedToTags bool          `json:"restricted_to_tags"`
	Status           int           `json:"status"`
	Tags             *[]string     `json:"tags"`
}

const HOST = "rink.hockeyapp.net"
const API_BASE = "/api/2"

func (app *App) apiGetCall(call string, args *url.Values) (*http.Response, error) {
	client := &http.Client{}

	path := fmt.Sprintf("%s/apps/%s%s", API_BASE, app.hockeyAppId, call)
	url := &url.URL{"https", "", nil, HOST, path, args.Encode(), ""}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-HockeyAppToken", app.hockeyApiToken)

	return client.Do(req)
}

func (app *App) apiPostCall(call string, formArgs *url.Values) (*http.Response, error) {
	client := &http.Client{}

	path := fmt.Sprintf("%s/apps/%s%s", API_BASE, app.hockeyAppId, call)
	url := &url.URL{"https", "", nil, HOST, path, "", ""}

	req, err := http.NewRequest("POST", url.String(), strings.NewReader(formArgs.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-HockeyAppToken", app.hockeyApiToken)

	return client.Do(req)
}

func (app *App) SetBugTrackerUrl(reason HockeyCrashReason, trackerUrl string) error {
	var err error
	call := fmt.Sprintf("/crash_reasons/%d", reason.Id)

	args := &url.Values{}
	args.Set("ticket_url", trackerUrl)

	var resp *http.Response
	if resp, err = app.apiPostCall(call, args); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("HockeyApp error: %s", resp.Status))
	}

	return nil
}

func (app *App) GetCrashes(reason HockeyCrashReason) ([]HockeyCrash, error) {
	var err error
	call := fmt.Sprintf("/crash_reasons/%d", reason.Id)

	var resp *http.Response
	if resp, err = app.apiGetCall(call, &url.Values{}); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bytes []byte
	if bytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	var responseJson struct {
		Crashes []HockeyCrash `json:"crashes"`
	}

	if err := json.Unmarshal(bytes, &responseJson); err != nil {
		return nil, err
	}

	return responseJson.Crashes, nil
}

func (app *App) GetCrashLog(crash HockeyCrash) (string, error) {
	call := fmt.Sprintf("/crashes/%d", crash.Id)

	args := &url.Values{}
	args.Set("format", "log")

	var err error
	var resp *http.Response
	if resp, err = app.apiGetCall(call, args); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var bytes []byte
	if bytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	}

	return string(bytes), nil
}
