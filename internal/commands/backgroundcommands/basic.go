package backgroundcommands

import (
	"context"
	"fmt"
	"github.com/recyvan/gotsgzengine/internal/command"

	"io"
)

// BasicCommands 提供基础命令
type BasicCommands struct {
	tm *TaskManager
	//cw *CommandWrapper
	commands map[string]command.Ecommand
	registry *command.Registry
}

// NewBasicCommands 创建基础命令提供者
func NewBasicCommands(registry *command.Registry) (*BasicCommands, error) {
	tm, err := NewTaskManager(10)

	if err != nil {
		return nil, fmt.Errorf("failed to create task manager: %v", err)
	}
	//cw := NewCommandWrapper(tm)
	//cw.RegisterCommand()

	return &BasicCommands{tm: tm, commands: make(map[string]command.Ecommand), registry: registry}, nil
}

// RegisterCommand 注册命令
// 在 BasicCommands 中修改 RegisterCommand 方法
func (bc *BasicCommands) RegisterCommand() {
	cmds := bc.registry.List()
	for _, cmd := range cmds {
		if cmd.Background {
			bc.commands[cmd.Name] = cmd
			bc.tm.RegisterFunction(cmd.Name, cmd.Handler)
		}
	}
}

func (bc *BasicCommands) handleBg(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	if len(args) < 1 {
		fmt.Fprint(rw, "usage: bg <task_name> [args...]")
		return nil, nil
	}

	bc.RegisterCommand()

	bc.tm.StartTask(rw, args[0], args[1:]...)
	return nil, nil
}

// ProvideCommands 实现 command.CommandProvider 接口
func (bc *BasicCommands) ProvideCommands() []command.Ecommand {
	return []command.Ecommand{
		{
			Name:        "bg",
			Description: "将存在交互等耗时任务的命令放入后台(协程)运行",
			Usage:       "bg <task_name> <task_args>",
			Type:        "system",
			Background:  false,
			Handler:     bc.handleBg,
		},
		{
			Name:        "interact",
			Description: "与后台协程进行交互",
			Usage:       "interact <task_id>",
			Type:        "system",
			Background:  false,
			Handler:     bc.handleInteract,
		},
		{
			Name:        "check",
			Description: "列出所有后台(脚本或函数)协程",
			Usage:       "check",
			Type:        "system",
			Background:  false,
			Handler:     bc.handleList,
		},
		{
			Name:        "kill",
			Description: "杀死指定后台(脚本或函数)协程",
			Usage:       "kill <task_id>",
			Type:        "system",
			Background:  false,
			Handler:     bc.handleKill,
		},
		{
			Name:        "exit",
			Description: "退出程序",
			Usage:       "exit",
			Type:        "system",
			Background:  false,
			Handler:     bc.handleExit,
		},
	}
}

func (bc *BasicCommands) handleInteract(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	if len(args) != 1 {
		fmt.Fprint(rw, "usage: interact <task_id>")
		return nil, nil
	}
	err := bc.tm.InteractTask(rw, args[0])
	return nil, err
}

func (bc *BasicCommands) handleList(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	bc.tm.ListTasks(rw)
	return nil, nil
}

func (bc *BasicCommands) handleKill(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	if len(args) != 1 {
		fmt.Fprint(rw, "usage: kill <task_id>")
		return nil, nil
	}
	err := bc.tm.KillTask(rw, args[0])
	return nil, err
}

func (bc *BasicCommands) handleExit(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	bc.tm.pool.Release()
	fmt.Fprintln(rw, "Bye!")
	return []byte("Bye!"), nil
}
