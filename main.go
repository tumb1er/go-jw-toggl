package main

import (
	"go-jw-toggl/jwtoggl"
	"os"
)

func main() {
	app := jwtoggl.NewApp()
	os.Exit(app.Run(os.Args))
}
