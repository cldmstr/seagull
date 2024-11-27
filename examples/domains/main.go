package main

import (
	"log/slog"
	"os"

	"github.com/cldmstr/seagull/examples/domains/internal/app"
)

func main() {
	a := app.App{}

	err := a.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
