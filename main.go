package main

import (
	"embed"
	"github.com/sharify-labs/spine/server"
)

//go:embed frontend/*
var assets embed.FS
var version string

func main() {
	server.Start(assets, version)
}
