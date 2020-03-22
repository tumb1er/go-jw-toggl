package gtk_status

import (
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/tools"
)

const (
	MsgTypeOnline = "online"
	MsgTypeError  = "error"
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

func (s *GtkStatus) ShowError(msg string) {
	tools.Schedule(s.setStatus, Status{
		ctx: MsgTypeError,
		msg: msg,
	})
}

func (s *GtkStatus) ClearError() {
	tools.Schedule(s.clearStatus, MsgTypeError)
}

func (s *GtkStatus) clearStatus(ctx string) {
	ctxId := s.statusBar.GetContextId(ctx)
	s.statusBar.RemoveAll(ctxId)
}

func New(b *builder.Builder) *GtkStatus {
	statusBar := b.GetStatusbar()
	return &GtkStatus{
		statusBar: statusBar,
	}
}
