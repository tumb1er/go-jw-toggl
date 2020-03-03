package jwtoggl

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
	"path"
)

const (
	trayId      = "com.github.tumb1er.go-jw-toggl.notification"
	TogglIconBW = "toggl_bw.png"
	TogglIcon   = "toggl.png"
)

func NewIndicator(menu *gtk.Menu, menuItem *gtk.MenuItem) *appindicator.Indicator {
	indicator := appindicator.New(trayId, "", appindicator.CategoryApplicationStatus)
	indicator.SetLabel(" Work!", "")
	iconPath := getIconPath(TogglIconBW)
	indicator.SetIcon(iconPath)
	indicator.SetStatus(appindicator.StatusActive)
	indicator.SetMenu(menu)
	indicator.SetSecondaryActivateTarget(menuItem)
	return indicator
}

func getIconPath(filename string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get cwd:", err)
	}
	iconPath := path.Join(cwd, "glade/resources", filename)
	return iconPath
}

type CurrentEntryInfo struct {
	Title    string
	Body     string
	Label    string
	Status   appindicator.Status
	Icon     string
	Menu     string
	External bool
}

func (j *jApp) UpdateCurrentEntry(info CurrentEntryInfo) {

	j.indicator.SetLabel(info.Label, "")
	j.indicator.SetStatus(info.Status)
	j.indicator.SetIcon(getIconPath(info.Icon))

	j.menuNew.SetLabel(info.Menu)
	if err := j.popup.SetIcon(getIconPath(info.Icon)); err != nil {
		log.Fatal("Popup error:", err)
	} else if info.External {
		j.showNotification(info.Title, info.Body)
	}

}

func (j *jApp) showNotification(title, body string) {
	log.Printf("Notify: title=%s body=%s\n", title, body)
	if err := j.popup.Update("", title, body); err != nil {
		log.Println("Can't show popup:", err)
	}
}
