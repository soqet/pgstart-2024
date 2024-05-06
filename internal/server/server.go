package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func apiRouter(logger zerolog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/cmd", func(r chi.Router) {
		r.Post("/", handleCreateCmd(logger))
		r.Get("/", handleListCmd(logger))
		r.Get("/{id:.+}", handleGetCmd(logger))
	})
	return r
}

func NewRouter(logger zerolog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))

	r.Mount("/api", apiRouter(logger))
	return r
}