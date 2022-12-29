package main

import (
	"embed"
	"io/fs"
	"webapp/cmd"

	log "github.com/maerics/golog"
)

//go:embed public/*
var publicAssets embed.FS

func main() {
	cmd.PublicAssets = log.Must1(fs.Sub(publicAssets, "public"))
	cmd.Run()
}
