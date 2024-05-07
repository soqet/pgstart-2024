package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	commandrunner "pgstart/internal/command_runner"
	"pgstart/internal/database"
	"pgstart/internal/server"

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
	var logLevel zerolog.Level
	level, _ := os.LookupEnv("LOG_LEVEL")
	switch level {
	case "disabled":
		logLevel = zerolog.Disabled
	case "info":
		logLevel = zerolog.InfoLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(logLevel)
	logger.Info().Msg("Starting server")
	conn, err := pgx.Connect(context.Background(), mustHaveEnv(logger, "DB_URL"))
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	runner := commandrunner.New(logger, database.New(conn))
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", mustHaveEnv(logger, "PORT")),
		Handler: server.NewRouter(logger, runner),
	}
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("")
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	logger.Info().Msg("Ready to accept requests")
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("")
	}
	cancel()
	logger.Info().Msg("Server gracefully shutted down")
}
