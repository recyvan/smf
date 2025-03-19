package customcommands

import (
	"context"
	"fmt"
	"github.com/recyvan/gotsgzengine/internal/command"
	"io"
)

type CustomCommand struct {
	// 自定义命令的属性
}

func NewCustomCommands() *CustomCommand {
	return &CustomCommand{}
}

func (cc *CustomCommand) ProvideCommands() []command.Ecommand {
	// 创建命令
	return []command.Ecommand{
		{
			Name:        "test",
			Description: "This is a test command",
			Usage:       "test",
			Type:        "customcommands",
			Background:  true, // 支持后台运行
			Handler:     cc.test,
		},
	}
}

// test 命令实现
func (cc *CustomCommand) test(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	fmt.Fprintln(rw, "hello world ")
	fmt.Fprintln(rw, "Please enter your name: ")

	var name string
	fmt.Fscanln(rw, &name)
	fmt.Fprintln(rw, "Hello World, "+name)
	return []byte("Test command executed successfully!"), nil
}
