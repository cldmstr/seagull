package app

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cldmstr/seagull/examples/domains/internal/quotes"
	"github.com/cldmstr/seagull/render"
)

//go:embed views
var views embed.FS

type App struct{}

func (a App) Run() error {

	slog.Info("Starting service")

	renderer, err := render.New("app/index.html", "/")
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	err = renderer.AddFS("app", views, true)
	if err != nil {
		return fmt.Errorf("add app templates: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(renderer.BasePath(), renderer.BasePathHandler())

	err = quotes.RegisterRoutes(mux, renderer)
	if err != nil {
		return fmt.Errorf("register quote routes: %w", err)
	}

	err = http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		return fmt.Errorf("run service: %w", err)
	}

	return nil
}
