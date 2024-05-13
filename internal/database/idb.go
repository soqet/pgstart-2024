package database

import "context"

type DB interface {
	AddCmd(context.Context, string) (uint64, error)
	GetCmdByID(ctx context.Context, id uint64) (Command, error)
	UpdateCmd(ctx context.Context, cmd Command) error
	SetAllEnded(ctx context.Context) error
	DeleteCmd(ctx context.Context, id uint64) error
	ListCommands(ctx context.Context) ([]Command, error)
	ListEndedCommands(ctx context.Context) ([]Command, error)
}
