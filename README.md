## 本项目是一个基于golang的任意脚本集中管理引擎，可以帮助你快速部署、管理、监控你的脚本。可以对多台服务器进行远程管理或多台客户端进行远程管理，并实时监控脚本运行状态。

## 使用方式

- 本项目支持本地运行和远程运运行。
- 可以通过安装golang环境，下载本项目代码，执行`go mod tidy`安装依赖包，
- 然后执行 `go run localserver/main.go` main.go` 启动本地运行模式，
- 远程模式：生成证书：
- `openssl genrsa -out server.key 2048`
- `openssl req -new -key server.key -out server.csr`
- `openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt`
- 并执行`go run cmd/server/main.go cmd/server/server.go cmd/server/engine_init.go -sc server.crt -sk server.key -p 8080
`运行服务端
- 执行 `go run cmd/client/main.go cmd/client/conn.go -h 127.0.0.1:8080  -u 1234 -p 1234` 运行客户端
- 其中客户端和服务端均支持多端连接，客户端运行执行`change conn.ID`切换连接，可以多个连接共同操作一台服务器。
- 
- 或者编译成可执行文件，直接运行即可(测试阶段！)。
- 用户口令存放位置默认为：token.text,内容格式为：`用户名:密码`

## 已包含功能

- 基本的与系统交互功能：可以执行执行系统命令等
- (后台运行)脚本监控：可以实时监控脚本的运行状态，包括脚本执行状态、脚本执行时间、脚本执行结果等
- 自定义脚本加载，支持任意功能或框架或二进制文件加载到引擎中运行并管理。
- 远程模式支持多端连接，运行多个连接共同操作一台服务器。多个客户端可以同时连接到同一台服务器，并执行命令。
- 以支持python脚本运行,并执行后台交互运行，详见help pyexec命令。

## 自定义脚本加载方式
- 支持编写任意go脚本，放到plugins目录下即可进行加载注册 ，或者可以编译成so文件，然后加载到引擎中运行。详细如下：
```go
package plugins

/*
# 在 plugins目录下执行->建议进行目录执行
go mod init github.com/recyvan/smf/plugins/package_name
# 添加主项目依赖
go mod edit -require github.com/recyvan/smf@latest
本地建议在mod中执行go mod edit -replace命令
执行
 */
go build -buildmode=plugins -o ./name.so name.go
确保 name.so 在 plugins 目录或子目录下
*/
import (
	"context"
	"fmt"
	"github.com/recyvan/smf/internal/command"
	"io"
	"os/user"
	"time"
)

// Plugin 导出的插件变量（必须导出这个变量名）
var Plugin plugin

// plugins 实现 command.CommandProvider 接口
type plugin struct{}

// ProvideCommands 返回插件提供的命令列表
func (p plugin) ProvideCommands() []command.Ecommand {
	return []command.Ecommand{
		{
			Name:        "datetime",                                    // 命令名称
			Description: "Show current date/time and user information", // 命令描述
			Usage:       "datetime",                                    // 命令用法
			Type:        "plugins",                                      // 命令类型
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
go build -buildmode=plugin -o
```
> 若脚本要实现后台运行的功能，请在接口中设置Background为true，可用于创建后台任务。
并在Handler中实现后台运行的逻辑。func (rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error)其中rw为脚本的标准输入输出，ctx为上下文，args为命令行参数。


- 对于二进制文件，在plugins目录下，按照如下方式进行编写：(so文件还在测试阶段，咱不可用)
> 将脚本编写完成后执行 `go build -o test.so -buildmode=plugin test.go` 编译成so文件，并将so文件放入plugins目录下。
> 但该功能仅支持linux系统，windows系统暂不支持编译。
> *注意*，建议下载源码后在plugins下在进行编译，避免依赖包版本问题。

