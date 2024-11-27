package quotes

import (
	"embed"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"

	"github.com/cldmstr/seagull/render"
)

var quotes = []string{
	`You will begin to touch heaven, Jonathan, in the moment that you touch perfect speed. And that isn’t flying a thousand miles an hour, or a million, or flying at the speed of light. Because any number is a limit, and perfection doesn’t have limits. Perfect speed, my son, is being there.`,
	`You have the freedom to be yourself, your true self, here and now, and nothing can stand in your way.`,
	`Don’t believe what your eyes are telling you. All they show is limitation. Look with your understanding. Find out what you already know and you will see the way to fly.`,
	`He was not bone and feather but a perfect idea of freedom and flight, limited by nothing at all`,
	`Heaven is not a place, and it is not a time. Heaven is being perfect. - And that isn't flying a thousand miles an hour, or a million, or flying at the speed of light. Because any number is a limit, and perfection doesn't have limits. Perfect speed, my son, is being there.`,
	`He spoke of very simple things- that it is right for a gull to fly, that freedom is the very nature of his being, that whatever stands against that freedom must be set aside, be it ritual or superstition or limitation in any form.

"Set aside," came a voice from the multitude, "even if it be the Law of the Flock?"

"The only true law is that which leads to freedom," Jonathan said. "There is no other.`,
	`To fly as fast as thought, to anywhere that is, you must begin by knowing that you have already arrived.`,
	`You don't love hatred and evil, of course. You have to practice and see the real gull, the good in every one of them, and to help them see it in themselves. That's what I mean by love.`,
	`For most gulls it was not flying that matters, but eating. For this gull, though, it was not eating that mattered, but flight.`,
}

//go:embed views
var views embed.FS

func RegisterRoutes(mux *http.ServeMux, r *render.Renderer) error {
	q := quoteHandler{renderer: r}

	err := r.AddFS("quotes", views, false)
	if err != nil {
		return fmt.Errorf("register quotes templates: %w", err)
	}

	mux.HandleFunc("/quote", q.handleQuote)

	return nil
}

type quoteHandler struct {
	renderer *render.Renderer
}

func (h quoteHandler) handleQuote(w http.ResponseWriter, req *http.Request) {
	slog.Info("Quote base request")

	qNum := rand.Intn(len(quotes))
	err := h.renderer.Render(req.Context(), w, "quotes/quote.tmpl.html", quotes[qNum])
	if err != nil {
		slog.Error("render quote", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
