package corecommands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PyExecOptions Python脚本执行选项
type PyExecOptions struct {
	FilePath    string            // 脚本文件路径
	Args        []string          // 传递给脚本的参数
	Env         map[string]string // 环境变量
	PythonPath  string            // Python解释器路径
	Interactive bool              // 是否交互模式
}

func (cc *CoreCommands) handlePyExec(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	opts, err := parsePyExecArgs(args)
	if err != nil {
		return nil, err
	}

	// 验证文件存在
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("python script not found: %s", opts.FilePath)
	}

	// 构建命令
	cmdArgs := []string{opts.FilePath}
	cmdArgs = append(cmdArgs, opts.Args...)

	cmd := exec.CommandContext(ctx, opts.PythonPath, cmdArgs...)

	// 设置环境变量
	if len(opts.Env) > 0 {
		cmd.Env = os.Environ() // 保留现有环境变量
		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置标准输入输出
	cmd.Stdout = rw
	cmd.Stderr = rw

	if opts.Interactive {
		cmd.Stdin = rw
	}

	// 执行命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start python script: %v", err)
	}

	// 如果是交互模式，等待命令完成
	if opts.Interactive {
		if err := cmd.Wait(); err != nil {
			return nil, fmt.Errorf("python script execution failed: %v", err)
		}
	}

	return []byte(fmt.Sprintf("Python script execution started: %s\n", opts.FilePath)), nil
}

func parsePyExecArgs(args []string) (*PyExecOptions, error) {
	opts := &PyExecOptions{
		PythonPath: "python3",
		Env:        make(map[string]string),
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing script file path")
			}
			opts.FilePath = args[i+1]
			i++
		case "-p", "--python":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing python interpreter path")
			}
			opts.PythonPath = args[i+1]
			i++
		case "-e", "--env":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing environment variable")
			}
			parts := strings.SplitN(args[i+1], "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid environment variable format: %s", args[i+1])
			}
			opts.Env[parts[0]] = parts[1]
			i++
		case "-i", "--interactive":
			opts.Interactive = true
		default:
			// 如果已经找到文件路径，将剩余参数作为脚本参数
			if opts.FilePath != "" {
				opts.Args = append(opts.Args, args[i])
			}
		}
	}

	if opts.FilePath == "" {
		return nil, fmt.Errorf("script file path is required (-f option)")
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(opts.FilePath)
	if err == nil {
		opts.FilePath = absPath
	}

	return opts, nil
}
