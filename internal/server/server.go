package server

import (
	"pgstart/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)


func apiRouter(logger zerolog.Logger, db *database.DB) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/cmd", func(r chi.Router) {
		r.Post("/", handleCreateCmd(logger, db))
		r.Get("/", handleListCmd(logger, db))
		r.Get("/{id:[0-9]+}", handleGetCmd(logger, db))
	})
	return r
}

func NewRouter(logger zerolog.Logger, db *database.DB) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))

	r.Mount("/api", apiRouter(logger, db))
	return r
}