package corecommands

import (
	"context"
	"fmt"
	"github.com/recyvan/gotsgzengine/internal/command"
	"io"
	"sort"
	"strings"
)

// handleHelp displays help information
func (cc *CoreCommands) handleHelp(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	if len(args) == 0 {
		// Show help usage when no arguments
		helpText := `
Usage: help [command]

Display help information for commands.
    help          - Show this help message
    help <command> - Show detailed help for specific command

Example:
    help list     - Show help for the list command
`
		fmt.Fprint(rw, helpText)
		return []byte(helpText), nil
	}

	// Show specific command help
	cmdName := args[0]
	cmd, exists := cc.registry.Get(cmdName)
	if !exists {
		return nil, fmt.Errorf("command '%s' not found", cmdName)
	}

	cmdHelp := fmt.Sprintf(`
Command: %s
Type: %s
Description: %s
Usage: %s
Background: %v
`, cmd.Name, cmd.Type, cmd.Description, cmd.Usage, cmd.Background)

	fmt.Fprint(rw, cmdHelp)
	return []byte(cmdHelp), nil
}

// showGeneralHelp 显示通用帮助信息
func (cc *CoreCommands) showGeneralHelp(rw io.ReadWriter) ([]byte, error) {
	output := &strings.Builder{}
	fmt.Fprintf(rw, "Available Commands:\n")
	fmt.Fprintf(rw, "Use 'help <command>' for more information about a specific command\n")

	// 收集所有命令类型
	typeMap := make(map[string][]command.Ecommand)
	for _, cmd := range cc.registry.List() {
		cmdType := cmd.Type
		if cmdType == "" {
			cmdType = "other" // 处理未指定类型的命令
		}
		typeMap[cmdType] = append(typeMap[cmdType], cmd)
	}

	// 获取并排序所有类型
	types := make([]string, 0, len(typeMap))
	for t := range typeMap {
		types = append(types, t)
	}
	sort.Strings(types)

	// 按类型显示命令
	for _, cmdType := range types {
		cmds := typeMap[cmdType]
		if len(cmds) > 0 {
			// 将类型首字母大写显示
			displayType := strings.ToUpper(cmdType[:1]) + cmdType[1:]
			fmt.Fprintf(rw, "\n%s Commands:\n", displayType)

			// 对每个类型中的命令按名称排序
			sort.Slice(cmds, func(i, j int) bool {
				return cmds[i].Name < cmds[j].Name
			})

			// 找出最长的命令名，用于对齐
			maxNameLen := 0
			for _, cmd := range cmds {
				if len(cmd.Name) > maxNameLen {
					maxNameLen = len(cmd.Name)
				}
			}

			// 显示命令
			for _, cmd := range cmds {
				// 添加背景任务标记
				bgMark := " "
				if cmd.Background {
					bgMark = "*" // 使用星号标记支持后台运行的命令
				}

				fmt.Fprintf(rw, "  %s %-*s  %s\n",
					bgMark,
					maxNameLen,
					cmd.Name,
					cmd.Description)
			}
		}
	}

	// 添加图例说明
	fmt.Fprintf(rw, "\nLegend:\n")
	fmt.Fprintf(rw, "  * Command can run in background\n")

	fmt.Fprint(rw, output.String())
	return []byte(output.String()), nil
}
