package main

import (
	"go-jw-toggl/jwtoggl"
	"go-jw-toggl/jwtoggl/tools"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	db, err := gorm.Open(sqlite.Open(path.Join(dir, "jwtoggl.db")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	defer func() {
		if conn, err := db.DB(); err != nil {
			log.Fatal(err)
		} else if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	app := jwtoggl.NewApp(db)
	os.Exit(app.Run(os.Args))
}
