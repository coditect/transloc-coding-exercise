package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/coditect/transloc-coding-exercise/config"
	"github.com/coditect/transloc-coding-exercise/rest"
	"github.com/coditect/transloc-coding-exercise/sqlite"
	"github.com/NYTimes/gziphandler"
)

func main() {

	c, err := config.Read()
	if err != nil {
		fmt.Println("Configuration error:", err)
		os.Exit(1)
	}

	db, err := sqlite.New(c.Database)
	if err != nil {
		fmt.Println("Unable to initialize database:", err)
		os.Exit(1)
	}

	server := rest.NewServer(db, c.RootDir)
	err = http.ListenAndServe(c.Listen, gziphandler.GzipHandler(server))
	if err != nil {
		fmt.Println("HTTP server error:", err)
		os.Exit(1)
	}
}
