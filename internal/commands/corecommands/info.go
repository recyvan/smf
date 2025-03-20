package corecommands

import (
	"context"
	"fmt"
	"github.com/recyvan/smf/internal/command"
	"io"
)

// handleInfo 处理info命令
func (cc *CoreCommands) handleInfo(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	info := fmt.Sprintf(`
Engine Information:
------------------
Version:    1.0.0
Build Time: 2025-03-15
Go Version: %s
OS/Arch:    %s/%s

Commands:   %d total
Background: %d commands capable
`, "go1.20", "linux", "amd64", len(cc.registry.List()), countBackgroundCommands(cc.registry))

	fmt.Fprint(rw, info)
	return []byte(info), nil
}
func countBackgroundCommands(registry *command.Registry) int {
	count := 0
	for _, cmd := range registry.List() {
		if cmd.Background {
			count++
		}
	}
	return count
}
