package tools

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

func GetIconPath(filename string) string {
	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Could not get cwd:", err)
	}
	iconPath := path.Join(cwd, "glade/resources", filename)
	return iconPath
}
