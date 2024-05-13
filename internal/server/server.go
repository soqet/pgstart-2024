package server

import (
	cr "pgstart/internal/command_runner"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)


func apiRouter(logger zerolog.Logger, runner *cr.Runner) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/cmd", func(r chi.Router) {
		r.Post("/", handleCreateCmd(logger, runner))
		r.Get("/", handleListCmd(logger, runner))
		r.Get("/{id:[0-9]+}", handleGetCmd(logger, runner))
		r.Post("/{id:[0-9]+}/kill", handleKillCmd(logger, runner))
	})
	return r
}

func NewRouter(logger zerolog.Logger, runner *cr.Runner) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))

	r.Mount("/api", apiRouter(logger, runner))
	return r
}
