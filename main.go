package main

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"os"
	"path"
)

func main() {
	dir := tools.GetAppDir()
	db, err := gorm.Open("sqlite3", path.Join(dir, "jwtoggl.db"))
	if err != nil {
		panic("failed to connect database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	app := jwtoggl.NewApp(db)
	os.Exit(app.Run(os.Args))
}
