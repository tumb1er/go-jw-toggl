package tray

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tumb1er/go-notifier/notifier"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
)

const (
	menuShow    = "menu_show"
	menuExit    = "menu_exit"
	menuSync    = "menu_sync"
	trayId      = "go-jw-toggl.notification"
	trayMenuId  = "tray_menu"
	TogglIconBW = "toggl_bw.png"
	TogglIcon   = "toggl.png"
)

type Tray struct {
	menu      *gtk.Menu
	menuShow  *gtk.MenuItem
	menuSync  *gtk.MenuItem
	menuExit  *gtk.MenuItem
	indicator *appindicator.Indicator
	popup     *notifier.Notifier
}

func New(b *builder.Builder) (*Tray, error) {
	icon := tools.GetIconPath(TogglIconBW)
	t := &Tray{
		indicator: appindicator.New(trayId, icon, appindicator.CategoryApplicationStatus),
	}
	var err error
	if t.menu, err = b.GetMenu(trayMenuId); err != nil {
		return nil, err
	}
	if t.menuShow, err = b.GetMenuItem(menuShow); err != nil {
		return nil, err
	}
	if t.menuSync, err = b.GetMenuItem(menuSync); err != nil {
		return nil, err
	}
	if t.menuExit, err = b.GetMenuItem(menuExit); err != nil {
		return nil, err
	}
	if t.popup, err = notifier.NewNotifier(icon); err != nil {
		return nil, err
	}
	t.SetCurrentEntry(nil)
	t.indicator.SetStatus(appindicator.StatusActive)
	t.indicator.SetMenu(t.menu)
	t.indicator.SetSecondaryActivateTarget(t.menuShow)
	return t, nil
}

func (t *Tray) ConnectSignals(
	onNew func(),
	onSync func(),
	onExit func(),
) error {
	if _, err := t.menuShow.Connect("activate", onNew); err != nil {
		return err
	}
	if _, err := t.menuExit.Connect("activate", onExit); err != nil {
		return err
	}
	if _, err := t.menuSync.Connect("activate", onSync); err != nil {
		return err
	}
	return nil
}

func (t *Tray) SetCurrentEntry(entry *gttimeentry.TimeEntry) {
	if entry == nil {
		t.indicator.SetLabel("Work!", "")
		icon := tools.GetIconPath(TogglIconBW)
		t.indicator.SetIcon(icon)
		if err := t.popup.SetIcon(icon); err != nil {
			log.Fatal("Popup error:", err)
		}
		t.menuShow.SetLabel("New...")
	} else {
		t.indicator.SetLabel(tools.Shorten(entry.Description), "")
		icon := tools.GetIconPath(TogglIcon)
		t.indicator.SetIcon(icon)
		t.menuShow.SetLabel(entry.Description)
		if err := t.popup.SetIcon(icon); err != nil {
			log.Fatal("Popup error:", err)
		}
	}
}

func (t *Tray) ShowNotification(title, body string) error {
	log.Printf("Notify: title=%s body=%s\n", title, body)
	return t.popup.Update("", title, body)
}
