package commandrunner

import (
	"context"
	"errors"
	"os"
	"pgstart/internal/database"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var _ database.DB = (*MockDB)(nil)

type MockDB struct {
	mtx sync.RWMutex
	rows map[uint64]database.Command
	curIdx uint64
}

func NewMDB() database.DB {
	return &MockDB{
		rows: make(map[uint64]database.Command),
	}
}

func (db *MockDB) AddCmd(_ context.Context, cmd string) (uint64, error) {
	db.mtx.Lock()
	id := db.curIdx
	db.rows[id] = database.Command{
		ID: db.curIdx,
		Cmd: cmd,
	}
	db.curIdx++
	db.mtx.Unlock()
	return id, nil
}

func (db *MockDB) GetCmdByID(_ context.Context, id uint64) (database.Command, error) {
	db.mtx.RLock()
	c, ok := db.rows[id]
	db.mtx.RUnlock()
	if !ok {
		return c, errors.New("No command with given id")
	}
	return c, nil
}

func (db *MockDB) UpdateCmd(_ context.Context, cmd database.Command) error {
	db.mtx.Lock()
	db.rows[cmd.ID] = cmd
	db.mtx.Unlock()
	return nil
}

func (db *MockDB) SetAllEnded(_ context.Context) error {
	db.mtx.Lock()
	for k, v := range db.rows {
		v.IsEnded = true
		db.rows[k] = v
	}
	db.mtx.Unlock()
	return nil
}

func (db *MockDB) DeleteCmd(_ context.Context, id uint64) error {
	db.mtx.Lock()
	delete(db.rows, id)
	db.mtx.Unlock()
	return nil
}

func (db *MockDB) ListCommands(ctx context.Context) ([]database.Command, error) {
	var res []database.Command 
	db.mtx.RLock()
	for _, v := range db.rows {
		res = append(res, v)
	}
	db.mtx.RLock()
	return res, nil
}

func (db *MockDB) ListEndedCommands(ctx context.Context) ([]database.Command, error) {
	var res []database.Command 
	db.mtx.RLock()
	for _, v := range db.rows {
		if !v.IsEnded {
			continue
		}
		res = append(res, v)
	}
	db.mtx.RLock()
	return res, nil
}

func TestBasic(t *testing.T) {
	db := NewMDB()
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(zerolog.Disabled)
	r := New(logger, db)
	ctx := context.Background()
	id, err := r.Exec(ctx, "echo 123")
	if err != nil {
		t.Fatalf("can't exec command: %s", err)
	}
	time.Sleep(time.Second)
	c, err := r.GetCmd(ctx, id)
	if err != nil {
		t.Fatalf("can't get command: %s", err)
	}
	if c.Result != "123\n" {
		t.Fatalf("wrong result of command. expected: 123; got: %s", c.Result)
	}
}

func Test2Commands(t *testing.T) {
	db := NewMDB()
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(zerolog.Disabled)
	r := New(logger, db)
	ctx := context.Background()
	id, err := r.Exec(ctx, "echo 123; echo 456")
	if err != nil {
		t.Fatalf("can't exec command: %s", err)
	}
	time.Sleep(time.Second)
	c, err := r.GetCmd(ctx, id)
	if err != nil {
		t.Fatalf("can't get command: %s", err)
	}
	if c.Result != "123\n456\n" {
		t.Fatalf("wrong result of command. expected: 123\n456\n; got: %s", c.Result)
	}
}

func TestLongCommand(t *testing.T) {
	db := NewMDB()
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(zerolog.Disabled)
	r := New(logger, db)
	ctx := context.Background()
	id, err := r.Exec(ctx, "echo 123; sleep 3; echo 456")
	if err != nil {
		t.Fatalf("can't exec command: %s", err)
	}
	time.Sleep(time.Second)
	c, err := r.GetCmd(ctx, id)
	if err != nil {
		t.Fatalf("can't get command: %s", err)
	}
	if c.Result != "123\n" {
		t.Fatalf("wrong result of command. expected: 123\n; got: %s", c.Result)
	}
	time.Sleep(time.Second * 3)
	c, err = r.GetCmd(ctx, id)
	if err != nil {
		t.Fatalf("can't get command: %s", err)
	}
	if c.Result != "123\n456\n" {
		t.Fatalf("wrong result of command. expected: 123\n456\n; got: %s", c.Result)
	}
}

func TestManyCommands(t *testing.T) {
	count := uint64(100)
	db := NewMDB()
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.TimeOnly}).Level(zerolog.Disabled)
	r := New(logger, db)
	ctx := context.Background()
	for i := uint64(0); i < count; i++ {
		id, err := r.Exec(ctx, "echo 123; sleep 5; echo 456")
		if err != nil {
			t.Fatalf("can't exec command: %s", err)
		}
		if id != i {
			t.Fatalf("wrong result of command. expected: %d; got: %d", id, i)
		}
	}
	time.Sleep(time.Second * 6)
	for i := uint64(0); i < count; i++ {
		c, err := r.GetCmd(ctx, i)
		if err != nil {
			t.Fatalf("can't get command: %s", err)
		}
		if c.Result != "123\n456\n" {
			t.Fatalf("wrong result of command (id: %d). expected: 123\n456\n; got: %s", i, c.Result)
		}
	}
}
