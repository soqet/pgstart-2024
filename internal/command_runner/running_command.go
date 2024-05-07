package commandrunner

import (
	"context"
	"strings"
	"sync"
)

type runningCommand struct {
	id        uint64
	script    string
	result    strings.Builder
	resultMtx sync.Mutex
	cancel context.CancelFunc
}

func (c *runningCommand) Write(p []byte) (int, error) {
	c.resultMtx.Lock()
	n, err := c.result.Write(p)
	c.resultMtx.Unlock()
	return n, err
}

func (c *runningCommand) Result() string {
	c.resultMtx.Lock()
	res := c.result.String()
	c.resultMtx.Unlock()
	return res
}

func (c *runningCommand) Kill() {
	c.cancel()
}
