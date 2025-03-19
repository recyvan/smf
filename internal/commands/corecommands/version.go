package corecommands

import (
	"context"
	"fmt"
	"io"
)

// handleVersion 处理version命令
func (cc *CoreCommands) handleVersion(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	version := `
Engine Version Information:
-------------------------
Version:     1.0.0
Build:       20250315
Commit:      abc123def
Build Date:  2025-03-15
Go Version:  go1.20
`
	fmt.Fprint(rw, version)
	return []byte(version), nil
}
