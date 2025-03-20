package plugins

import (
	"context"
	"fmt"
	"github.com/recyvan/smf/internal/command"
	"io"
	"os/user"
	"time"
)

type PluginCommand struct {
}

func NewPluginCommand() *PluginCommand {
	return &PluginCommand{}
}
func (p *PluginCommand) ProvideCommands() []command.Ecommand {
	return []command.Ecommand{
		{
			Name:        "datetime",                                    // 命令名称
			Description: "Show current date/time and user information", // 命令描述
			Usage:       "datetime",                                    // 命令用法
			Type:        "plugin",                                      // 命令类型
			Background:  false,                                         // 是否支持后台运行
			Handler:     p.handleDateTime,                              // 命令处理函数
		},
	}
}

// handleDateTime 处理 datetime 命令
func (p *PluginCommand) handleDateTime(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	// 获取当前时间
	now := time.Now().Local()

	// 获取当前用户
	currentUser, err := user.Current()
	if err != nil {
		currentUser = &user.User{Username: "unknown"}
	}

	// 格式化输出
	output := fmt.Sprintf("Current Date and Time (UTC - YYYY-MM-DD HH:MM:SS formatted): %s\nCurrent User's Login: %s\n",
		now.Format("2006-01-02 15:04:05"),
		currentUser.Username)

	// 写入输出
	fmt.Fprint(rw, output)
	return []byte(output), nil
}
