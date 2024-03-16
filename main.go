package main

import (
	"embed"
	"github.com/posty/spine/server"
)

//go:embed frontend/*
var assets embed.FS

func main() {
	server.Start(assets)
}
