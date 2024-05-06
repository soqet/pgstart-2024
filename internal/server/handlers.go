package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"pgstart/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func handleCreateCmd(logger zerolog.Logger, db *database.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger = logger.With().Str("endpoint", "create cmd").Logger()
		logger.Debug().Msg("got request")
		req := CreateCmdRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if errors.As(err, new(*json.SyntaxError)) || errors.As(err, new(*json.UnmarshalTypeError)) {
			logger.Debug().Err(err).Msg("invalid request")
			resp := ResponseSchema{
				Errors: &[]ErrorSchema{{
					Code: ParseError,
					Desc: "Invalid body",
				}},
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		id, err := db.AddCmd(r.Context(), req.Script)
		if err != nil {
			logger.Error().Err(err).Msg("invalid request")
			resp := ResponseSchema{
				Errors: &[]ErrorSchema{{
					Code: InternalError,
					Desc: "Something went wrong",
				}},
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := ResponseSchema{
			Data: CreateCmdResponse{
				ID: id,
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

func handleListCmd(logger zerolog.Logger, db *database.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger = logger.With().Str("endpoint", "list cmd").Logger()
		logger.Debug().Msg("got request")
		dbCommands, err := db.ListCommands(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("invalid request")
			resp := ResponseSchema{
				Errors: &[]ErrorSchema{{
					Code: InternalError,
					Desc: "Something went wrong",
				}},
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
		respCommands := []CommandSchema{}
		for _, c := range dbCommands {
			respCommands = append(respCommands, CommandSchema{
				ID: c.ID,
				Script: c.Cmd,
				IsEnded: c.IsEnded,
				Result: c.Result,
			})
		}
		resp := ResponseSchema{
			Data: respCommands,
		}
		json.NewEncoder(w).Encode(resp)
	})
}

func handleGetCmd(logger zerolog.Logger, db *database.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger = logger.With().Str("endpoint", "get cmd").Logger()
		logger.Debug().Msg("got request")
		cmdID := chi.URLParam(r, "id")
		ID, err := strconv.ParseUint(cmdID, 10, 64)
		if err != nil {
			logger.Warn().Err(err).Str("parameter", "id").Msg("invalid path parameter. change routes")
			logger.Debug().Err(err).Msg("invalid request")
			resp := ResponseSchema{
				Errors: &[]ErrorSchema{{
					Code: ParseError,
					Desc: "Invalid id",
				}},
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		dbCmd, err := db.GetCmdByID(r.Context(), ID)
		if err != nil {
			logger.Error().Err(err).Msg("invalid request")
			resp := ResponseSchema{
				Errors: &[]ErrorSchema{{
					Code: InternalError,
					Desc: "Something went wrong",
				}},
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := ResponseSchema{
			Data: GetCmdResponse{
				CommandSchema: CommandSchema{
					ID: dbCmd.ID,
					Script: dbCmd.Cmd,
					IsEnded: dbCmd.IsEnded,
					Result: dbCmd.Result,
				},

			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}
