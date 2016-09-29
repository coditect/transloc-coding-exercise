package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Database string
	Listen   string
	RootDir  string
}

func Read() (*Config, error) {
	c := new(Config)
	flag.StringVar(&c.Listen, "listen", ":80", "Network address on which to listen")
	flag.StringVar(&c.Database, "database", ":memory:", "Path to SQLite3 database")
	flag.StringVar(&c.RootDir, "rootdir", "public_html", "Root directory for static assets")
	flag.Parse()

	// Make sure that the web root exists and is a directory
	info, err := os.Stat(c.RootDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", c.RootDir)
	}

	return c, nil
}
