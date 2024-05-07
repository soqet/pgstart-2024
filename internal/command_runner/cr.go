package commandrunner

import (
	"context"
	"errors"
	"os/exec"
	"pgstart/internal/database"
	"sync"

	"github.com/rs/zerolog"
)

var (
	ErrCmdIsNotRunning = errors.New("COMMAND IS NOT RUNNING")
)

type Runner struct {
	pending map[uint64]*runningCommand
	pmtx    sync.RWMutex
	db      *database.DB
	logger  zerolog.Logger
}

func New(logger zerolog.Logger, db *database.DB) *Runner {
	db.SetAllEnded(context.Background())
	return &Runner{
		pending: make(map[uint64]*runningCommand),
		db:      db,
		logger:  logger,
	}
}

func (r *Runner) saveToDb(cmd *runningCommand, isEnded bool) error {
	r.logger.Debug().Uint64("cmd id", cmd.id).Msg("saving to db")
	err := r.db.UpdateCmd(context.Background(), database.Command{
		ID:      cmd.id,
		Cmd:     cmd.script,
		IsEnded: isEnded,
		Result:  cmd.Result(),
	})
	return err
}

func (r *Runner) Exec(ctx context.Context, script string) (uint64, error) {
	id, err := r.db.AddCmd(ctx, script)
	if err != nil {
		r.logger.Error().Err(err).Msg("can't add command to db")
		return 0, err
	}
	logger := r.logger.With().Uint64("cmd_id", id).Logger()
	go func() {
		cmdCtx, cancelCmd := context.WithCancel(context.Background())
		defer cancelCmd()
		runningCmd := &runningCommand{
			id:          id,
			script:      script,
			cancel: cancelCmd,
		}
		r.pmtx.Lock()
		r.pending[id] = runningCmd
		r.pmtx.Unlock()
		defer func() {
			// release resources
			r.pmtx.Lock()
			delete(r.pending, runningCmd.id)
			r.pmtx.Unlock()
			err = r.saveToDb(runningCmd, true)
			if err != nil {
				logger.Error().Err(err).Msg("can't save to db completed command")
			}
		}()
		cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", script)
		cmd.Stdout = runningCmd
		err := cmd.Start()
		if err != nil {
			logger.Error().Err(err).Msg("can't start command")
			return
		}
		waitC := make(chan struct{})
		go func() {
			cmd.Wait()
			close(waitC)
		}()
		select {
		case <-cmdCtx.Done():
			logger.Debug().Msg("command killed")
		case <-waitC:
			logger.Debug().Str("result", runningCmd.Result()).Msg("command ended")
		}
	}()
	return id, nil
}

func (r *Runner) GetCmd(ctx context.Context, id uint64) (database.Command, error) {
	r.pmtx.RLock()
	c, ok := r.pending[id]
	r.pmtx.RUnlock()
	if ok {
		return database.Command{
			ID:      c.id,
			Cmd:     c.script,
			IsEnded: false,
			Result:  c.Result(),
		}, nil
	}
	cmd, err := r.db.GetCmdByID(ctx, id)
	if err != nil {
		return database.Command{}, err
	}
	return cmd, nil
}

func (r *Runner) ListCmd(ctx context.Context) ([]database.Command, error) {
	res := []database.Command{}
	r.pmtx.RLock()
	for _, cmd := range r.pending {
		res = append(res, database.Command{
			ID:      cmd.id,
			Cmd:     cmd.script,
			IsEnded: false,
			Result:  cmd.Result(),
		})
	}
	r.pmtx.RUnlock()
	ended, err := r.db.ListEndedCommands(ctx)
	if err != nil {
		return nil, err
	}
	res = append(res, ended...)
	return res, nil
}

func (r *Runner) KillCmd(id uint64) (database.Command, error) {
	r.pmtx.RLock()
	c, ok := r.pending[id]
	r.pmtx.RUnlock()
	if !ok {
		return database.Command{}, ErrCmdIsNotRunning
	}
	dbCmd := database.Command{
		ID:      c.id,
		Cmd:     c.script,
		IsEnded: true,
		Result:  c.Result(),
	}
	c.Kill()
	return dbCmd, nil
}
