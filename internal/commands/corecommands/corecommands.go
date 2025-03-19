package corecommands

import (
	"github.com/recyvan/gotsgzengine/internal/command"
)

var (
	pyexecstring = `pyexec [options] -f <script_path> [script_args...]
Options:
    -f, --file      Python script file path (required)
    -p, --python    Python interpreter path (default: python3)
    -e, --env       Set environment variables (format: KEY=VALUE)
    -i, --interactive Enable interactive mode
Examples:
    pyexec -f script.py
    pyexec -f script.py arg1 arg2
    pyexec -p /usr/local/bin/python3 -f script.py
    pyexec -e "PYTHONPATH=/custom/path" -f script.py
    bg pyexec -f long_running.py`
)

// CoreCommands 提供引擎核心命令
type CoreCommands struct {
	registry *command.Registry
}

// NewCoreCommands 创建核心命令提供者
func NewCoreCommands(registry *command.Registry) *CoreCommands {
	return &CoreCommands{
		registry: registry,
	}
}

// ProvideCommands 实现 command.CommandProvider 接口
func (cc *CoreCommands) ProvideCommands() []command.Ecommand {
	return []command.Ecommand{
		{
			Name:        "help",
			Description: "Display help information for commands",
			Usage:       "help [command]",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleHelp,
		},
		{
			Name:        "list",
			Description: "List all available commands",
			Usage:       "list [-a]",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleList,
		},
		{
			Name:        "info",
			Description: "Show engine and system information",
			Usage:       "info",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleInfo,
		},
		{
			Name:        "time",
			Description: "Display current time",
			Usage:       "time [format]",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleTime,
		},
		{
			Name:        "echo",
			Description: "Echo input text",
			Usage:       "echo [text...]",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleEcho,
		},
		{
			Name:        "version",
			Description: "Show engine version information",
			Usage:       "version",
			Type:        "system",
			Background:  false,
			Handler:     cc.handleVersion,
		},
		{
			Name:        "exec",
			Description: "Execute system command",
			Usage:       "exec <command> [args...]",
			Type:        "system",
			Handler:     cc.handleExec,
		},

		{
			Name:        "pyexec",
			Description: "Execute Python scripts with various options",
			Usage:       pyexecstring,
			Type:        "system",
			Background:  true, // 支持后台运行
			Handler:     cc.handlePyExec,
		},
	}
}

// 辅助函数
