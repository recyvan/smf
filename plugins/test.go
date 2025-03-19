package plugins

/*
# 在 plugins目录下执行->建议进行目录执行
go mod init datetime
# 添加主项目依赖
go mod edit -require github.com/recyvan/gotsgzengine@latest
执行
go build -buildmode=plugin -o ./datetime.so datetime.go
确保 datetime.so 在 plugins 目录下
*/
import (
	"context"
	"fmt"
	"github.com/recyvan/gotsgzengine/internal/command"
	"io"
	"os/user"
	"time"
)

// Plugin 导出的插件变量（必须导出这个变量名）
var Plugin plugin

// plugin 实现 command.CommandProvider 接口
type plugin struct{}

// ProvideCommands 返回插件提供的命令列表
func (p plugin) ProvideCommands() []command.Ecommand {
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
func (p plugin) handleDateTime(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
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
