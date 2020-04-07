package gtwebsocket

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
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

type OnTimeEntryActionCallback func(action string, entry gttimeentry.TimeEntry) error

type OnProjectActionCallback func(action string, project gtproject.Project) error
