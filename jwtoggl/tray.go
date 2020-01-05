package jwtoggl

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
	"path"
)

const trayId = "com.github.tumb1er.go-jw-toggl.notification"

func NewIndicator(menu *gtk.Menu, menuItem *gtk.MenuItem) *appindicator.Indicator {
	indicator := appindicator.New(trayId, "", appindicator.CategoryApplicationStatus)
	indicator.SetLabel(" #taskId", "")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get cwd:", err)
	}
	indicator.SetIcon(path.Join(cwd, "toggl.png"))
	indicator.SetStatus(appindicator.StatusActive)
	indicator.SetMenu(menu)
	indicator.SetSecondaryActivateTarget(menuItem)
	return indicator
}
