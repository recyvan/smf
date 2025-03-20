package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/recyvan/smf/internal/command"
	pc "github.com/recyvan/smf/internal/command/plugin"
	"github.com/recyvan/smf/internal/commands/backgroundcommands"
	"github.com/recyvan/smf/internal/commands/corecommands"
	"github.com/recyvan/smf/internal/commands/customcommands"
	"github.com/recyvan/smf/plugins"
	"io"
	"net"
	"os"
	"strings"
)

type ReadWriter struct {
	Reader io.Reader
	Writer io.Writer
}

// PluginCommandProvider 插件命令提供者
type PluginCommandProvider struct {
	commands []command.Ecommand
}

func (p *PluginCommandProvider) ProvideCommands() []command.Ecommand {
	return p.commands
}
func enginInit() *command.LocalEngine {

	engine_1 := command.NewLocalEngine()

	// 创建TaskManager
	_, err := backgroundcommands.NewTaskManager(10)
	if err != nil {
		fmt.Printf("Error creating task manager: %v\n", err)
		os.Exit(1)
	}
	//// 创建并添加基础命令提供者
	basicCommands, err := backgroundcommands.NewBasicCommands(engine_1.CmdRegistry)
	// 创建并添加自定义命令提供者
	customCommands := customcommands.NewCustomCommands()
	// 创建并添加核心命令提供者
	coreCommands := corecommands.NewCoreCommands(engine_1.CmdRegistry)
	// 注册并包装命令
	//注册未编译插件的命令
	pluginCommands := plugins.NewPluginCommand()

	// 添加提供者到自动注册器
	engine_1.AutoReg.AddProvider(basicCommands)
	engine_1.AutoReg.AddProvider(customCommands)
	engine_1.AutoReg.AddProvider(coreCommands)
	engine_1.AutoReg.AddProvider(pluginCommands)
	// 加载插件
	pluginDir := "./plugins"
	pluginLoader := pc.NewPluginLoader(pluginDir)

	if err := pluginLoader.LoadPlugins(); err != nil {
		fmt.Printf("Warning: error loading plugins: %v\n", err)
	}
	// 创建插件命令提供者
	pluginProvider := &PluginCommandProvider{commands: pluginLoader.GetCommands()}
	engine_1.AutoReg.AddProvider(pluginProvider)
	// 注册所有命令
	engine_1.AutoReg.RegisterAll(engine_1.CmdRegistry)
	return engine_1
}
func Run(engine *command.LocalEngine, conn net.Conn) {
	// 进行本地io重定向
	rw := NewReadWriter(conn, conn)
	fmt.Println("Engine Core v1.0.0 (2025-03-15)")
	fmt.Println("Type 'help' for available commands")
	fmt.Fprintln(rw, "Engine Core v1.0.0 (2025-03-15)")
	fmt.Fprintln(rw, "Type 'help' for available commands")
	defer conn.Close()
	scanner := bufio.NewScanner(rw.Reader)
	for {
		fmt.Fprint(rw, ">")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}

		cmd, exists := engine.CmdRegistry.Get(cmdName)
		if exists {
			_, err := cmd.Handler(rw, context.Background(), args)
			if err != nil {
				fmt.Fprintf(rw, "Error: %v\n", err)
			}
			if cmdName == "exit" {
				return
			}
		} else {
			fmt.Fprintf(rw, "Unknown command: %s\n", cmdName)
			fmt.Fprintln(rw, "Type 'help' for available commands")
		}
	}
}
func NewReadWriter(reader io.Reader, writer io.Writer) *ReadWriter {
	return &ReadWriter{
		Reader: reader,
		Writer: writer,
	}
}

// 实现 io.Reader 接口
func (rw *ReadWriter) Read(p []byte) (n int, err error) {
	return rw.Reader.Read(p)
}

// 实现 io.Writer 接口
func (rw *ReadWriter) Write(p []byte) (n int, err error) {
	return rw.Writer.Write(p)
}
