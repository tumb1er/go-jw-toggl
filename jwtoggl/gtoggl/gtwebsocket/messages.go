package gtwebsocket

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
)

type MsgPing struct {
	Type string `json:"type"`
}

type OnPingCallback func(ping MsgPing) error

type MsgAction struct {
	Action string `json:"action"`
	Model  string `json:"model"`
}

type MsgTimeEntryAction struct {
	MsgAction
	Data gttimeentry.TimeEntry `json:"data"`
}

type MsgProjectAction struct {
	MsgAction
	Data gtproject.Project `json:"data"`
}

type MsgClientAction struct {
	MsgAction
	Data gtclient.Client `json:"data"`
}

type MsgTagAction struct {
	MsgAction
	Data gttag.Tag `json:"data"`
}

type OnTimeEntryActionCallback func(action string, entry gttimeentry.TimeEntry) error

type OnProjectActionCallback func(action string, project gtproject.Project) error

type OnClientActionCallback func(action string, client gtclient.Client) error

type OnTagActionCallback func(action string, tag gttag.Tag) error
