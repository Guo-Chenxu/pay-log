package main

import (
	"embed"

	"github.com/Guo-Chenxu/pay-log/cmd/server"
)

//go:embed all:static
var staticFiles embed.FS

func main() {
	server.Execute(staticFiles)
}
