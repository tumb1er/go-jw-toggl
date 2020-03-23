package status

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/status/gtk_status"
	"log"
)

type Status struct {
	status    *gtk_status.GtkStatus
	online    bool
	connected bool
}

func New(b *builder.Builder) *Status {
	return &Status{status: gtk_status.New(b)}
}

func (s *Status) SetOnlineStatus(online bool) bool {
	if s.online == online {
		return false
	}
	log.Println("Online status:", online)
	s.online = online
	if online {
		s.connected = true
	}
	s.status.SetOnlineStatus(online)
	return !s.connected
}

func (s *Status) ShowError(msg string) {
	log.Println("Error:", msg)
	s.status.ShowError(msg)
}

func (s *Status) ClearError() {
	s.status.ClearError()
}
