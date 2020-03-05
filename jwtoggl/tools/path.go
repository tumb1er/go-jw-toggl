package tools

import (
	"log"
	"os"
	"path"
)

func GetIconPath(filename string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get cwd:", err)
	}
	iconPath := path.Join(cwd, "glade/resources", filename)
	return iconPath
}
