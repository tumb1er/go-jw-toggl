package gtk_status

import (
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/tools"
)

const (
	MsgTypeOnline = "online"
)

type GtkStatus struct {
	statusBar *gtk.Statusbar
}

type Status struct {
	ctx string
	msg string
}

func (s *GtkStatus) SetOnlineStatus(online bool) {
	var msg = "Offline"
	if online {
		msg = "Online"
	}
	tools.Schedule(s.setStatus, Status{
		ctx: MsgTypeOnline,
		msg: msg,
	})
}

func (s *GtkStatus) setStatus(status Status) {
	ctxId := s.statusBar.GetContextId(status.ctx)
	s.statusBar.RemoveAll(ctxId)
	s.statusBar.Push(ctxId, status.msg)
}

func New(b *builder.Builder) *GtkStatus {
	statusBar := b.GetStatusbar()
	return &GtkStatus{
		statusBar: statusBar,
	}
}
