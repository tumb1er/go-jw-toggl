package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go-jw-toggl/jwtoggl"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"os"
	"path"
)

//go:generate go generate go-jw-toggl/jwtoggl/store go-jw-toggl/jwtoggl/state

func main() {
	dir := path.Join(tools.GetAppDir(), "db")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open("sqlite3", path.Join(dir, "jwtoggl.db"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	app := jwtoggl.NewApp(db)
	os.Exit(app.Run(os.Args))
}
