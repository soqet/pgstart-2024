package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"pgstart/internal/server"
	"pgstart/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func mustHaveEnv(logger zerolog.Logger, envName string) string {
	env, ok := os.LookupEnv(envName)
	if !ok {
		logger.Panic().Str("env", envName).Msg("Missing env")
	}
	return env
}


func main() {
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(zerolog.DebugLevel)
	logger.Info().Msg("Starting server")
	conn, err := pgx.Connect(context.Background(), mustHaveEnv(logger, "DB_URL"))
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", mustHaveEnv(logger, "PORT")),
		Handler: server.NewRouter(logger, database.New(conn)),
	}
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("")
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger.Info().Msg("Server started")
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("")
	}
	cancel()
	logger.Info().Msg("Server gracefully shutted down")
}
