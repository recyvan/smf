package corecommands

import (
	"context"
	"fmt"
	"github.com/recyvan/smf/internal/command"
	"io"
	"sort"
	"strings"
)

// List命令 - 列出所有可用命令
func (cc *CoreCommands) handleList(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	commands := cc.registry.List()
	showDetailed := len(args) > 0 && args[0] == "-a"

	// Group commands by type
	typeGroups := make(map[string][]command.Ecommand)
	for _, cmd := range commands {
		typeGroups[cmd.Type] = append(typeGroups[cmd.Type], cmd)
	}

	// Sort types
	types := make([]string, 0, len(typeGroups))
	for t := range typeGroups {
		types = append(types, t)
	}
	sort.Strings(types)

	var output strings.Builder
	output.WriteString("Available Commands:\n")
	output.WriteString("==================\n\n")

	for _, cmdType := range types {
		output.WriteString(fmt.Sprintf("[%s]\n", strings.ToUpper(cmdType)))
		cmds := typeGroups[cmdType]
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name < cmds[j].Name
		})

		if showDetailed {
			// Detailed view
			for _, cmd := range cmds {
				output.WriteString(fmt.Sprintf("  %-12s - %s\n", cmd.Name, cmd.Description))
				output.WriteString(fmt.Sprintf("    Type: %s, Background: %v\n", cmd.Type, cmd.Background))
			}
		} else {
			// Simple view
			cmdNames := make([]string, len(cmds))
			for i, cmd := range cmds {
				cmdNames[i] = cmd.Name
			}
			output.WriteString("  " + strings.Join(cmdNames, ", ") + "\n")
		}
		output.WriteString("\n")
	}

	fmt.Fprint(rw, output.String())
	return []byte(output.String()), nil
}
