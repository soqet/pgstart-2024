package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var _ DB = (*dbConn)(nil)

type dbConn struct {
	conn *pgx.Conn
}

func New(db *pgx.Conn) DB {
	return &dbConn{conn: db}
}

type Command struct {
	ID      uint64
	Cmd     string
	IsEnded bool
	Result  string
}

func (db *dbConn) AddCmd(ctx context.Context, cmd string) (uint64, error) {
	const query = "INSERT INTO commands (cmd) VALUES ($1) RETURNING id"
	r := db.conn.QueryRow(ctx, query, cmd)
	var id uint64
	err := r.Scan(&id)
	return id, err
}

func (db *dbConn) GetCmdByID(ctx context.Context, id uint64) (Command, error) {
	const query = "SELECT * FROM commands WHERE id = $1 LIMIT 1"
	r := db.conn.QueryRow(ctx, query, id)
	var c Command
	err := r.Scan(&c.ID, &c.Cmd, &c.IsEnded, &c.Result)
	return c, err
}

const updateCmdQuery = `
INSERT INTO commands VALUES ($1, $2, $3, $4) 
ON CONFLICT (id) DO UPDATE 
SET cmd = excluded.cmd, is_ended = excluded.is_ended, result = excluded.result
`

func (db *dbConn) UpdateCmd(ctx context.Context, cmd Command) error {
	_, err := db.conn.Exec(ctx, updateCmdQuery, cmd.ID, cmd.Cmd, cmd.IsEnded, cmd.Result)
	return err
}

func (db *dbConn) SetAllEnded(ctx context.Context) error {
	const query = "UPDATE commands SET is_ended = true WHERE is_ended = false"
	_, err := db.conn.Exec(ctx, query)
	return err
}

func (db *dbConn) DeleteCmd(ctx context.Context, id uint64) error {
	const query = "DELETE FROM commands WHERE id = $1"
	_, err := db.conn.Exec(ctx, query, id)
	return err
}

func (db *dbConn) listCommands(ctx context.Context, query string) ([]Command, error) {
	rows, err := db.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Command
	for rows.Next() {
		var c Command
		if err := rows.Scan(&c.ID, &c.Cmd, &c.IsEnded, &c.Result); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (db *dbConn) ListCommands(ctx context.Context) ([]Command, error) {
	const query = "SELECT * FROM commands ORDER BY id"
	return db.listCommands(ctx, query)
}

func (db *dbConn) ListEndedCommands(ctx context.Context) ([]Command, error) {
	const query = "SELECT * FROM commands WHERE is_ended = true ORDER BY id"
	return db.listCommands(ctx, query)
}
