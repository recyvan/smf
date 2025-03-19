package corecommands

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// handleEcho 处理echo命令
func (cc *CoreCommands) handleEcho(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	output := strings.Join(args, " ") + "\n"
	fmt.Fprint(rw, output)
	return []byte(output), nil
}
