package backgroundcommands

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/recyvan/gotsgzengine/internal/command"
)

// CommandWrapper 包裹器，用于包装命令和任务管理器
type CommandWrapper struct {
	CmdRegistry *command.Registry
	TaskManager *TaskManager
	mu          sync.Mutex
}

// NewCommandWrapper 创建一个新的命令包裹器
func NewCommandWrapper(registry *command.Registry, tm *TaskManager) *CommandWrapper {
	return &CommandWrapper{
		CmdRegistry: registry,
		TaskManager: tm,
	}
}

// RegisterCommand 注册命令
func (cw *CommandWrapper) RegisterCommand(cmd command.Ecommand) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.CmdRegistry.Register(cmd)
}

// GetCommand 获取命令
func (cw *CommandWrapper) GetCommand(name string) (command.Ecommand, bool) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	for _, cmd := range cw.CmdRegistry.List() {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return command.Ecommand{}, false
}

// ListCommands 列出所有命令
func (cw *CommandWrapper) ListCommands() []command.Ecommand {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.CmdRegistry.List()
}

// RunCommand 运行命令
func (cw *CommandWrapper) RunCommand(rw io.ReadWriter, ctx context.Context, name string, args []string) ([]byte, error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cmd, exists := cw.CmdRegistry.Get(name)
	if !exists {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return cmd.Handler(rw, ctx, args)
}
