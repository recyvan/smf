package corecommands

import (
	"context"
	"fmt"
	"io"
	"time"
)

// handleTime 处理time命令
func (cc *CoreCommands) handleTime(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	format := "2006-01-02 15:04:05"
	if len(args) > 0 {
		format = args[0]
	}

	now := time.Now().UTC()
	output := fmt.Sprintf("Current UTC time: %s\n", now.Format(format))
	fmt.Fprint(rw, output)
	return []byte(output), nil
}
