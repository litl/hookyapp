package fogbugz

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Session struct {
	host  string
	token string
}

type Config interface {
	GetEmail() string
	GetPassword() string
	GetHost() string
}

type errorCodeTag struct {
	ErrorCode int    `xml:"code,attr"`
	ErrorDesc string `xml:",chardata"`
}

func (session *Session) FileBug(project string, area string, title string, content string) error {
	values := url.Values{}
	values.Set("cmd", "new")
	values.Set("token", session.token)
	values.Set("sProject", project)
	values.Set("sArea", area)
	values.Set("sTitle", title)
	values.Set("sEvent", content)
	url := &url.URL{"https", "", nil, session.host, "/api.asp", values.Encode(), ""}

	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type Response struct {
		XMLName xml.Name     `xml:"response"`
		Error   errorCodeTag `xml:"error"`
	}

	var r Response
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&r)
	if err != nil {
		return err
	}

	if r.Error.ErrorCode != 0 {
		return errors.New(fmt.Sprintf("Fogbugz error %d: %s", r.Error.ErrorCode, r.Error.ErrorDesc))
	}

	return nil
}

func fecthAuthToken(config Config) (string, error) {
	values := url.Values{}
	values.Set("cmd", "logon")
	values.Set("email", config.GetEmail())
	values.Set("password", config.GetPassword())
	url := &url.URL{"https", "", nil, config.GetHost(), "/api.asp", values.Encode(), ""}

	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	type Response struct {
		XMLName xml.Name     `xml:"response"`
		Token   string       `xml:"token"`
		Error   errorCodeTag `xml:"error"`
	}

	var r Response
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&r)
	if err != nil {
		return "", err
	}

	if r.Error.ErrorCode != 0 {
		return "", errors.New(fmt.Sprintf("Fogbugz error %d: %s", r.Error.ErrorCode, r.Error.ErrorDesc))
	}

	return r.Token, nil
}

func NewSession(config Config) (*Session, error) {
	session := new(Session)

	var err error
	if session.token, err = fecthAuthToken(config); err != nil {
		return nil, err
	}

	session.host = config.GetHost()

	return session, nil
}

func (session *Session) String() string {
	return fmt.Sprintf("FogBugzSession for token %s", session.token)
}
