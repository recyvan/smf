package corecommands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

func (cc *CoreCommands) handleExec(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error) {
	if len(args) == 0 {
		fmt.Fprint(rw, "missing command")
		return nil, nil
	}
	return executeCommand(rw, ctx, args[0], args[1:]...)
}

func executeCommand(writer io.ReadWriter, ctx context.Context, cmdName string, args ...string) ([]byte, error) {
	// 创建新的上下文，确保每次执行都是独立的
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	// Windows系统命令处理
	if runtime.GOOS == "windows" {
		if cmdName == "cmd" {
			if len(args) < 2 {
				fmt.Fprintf(writer, "invalid command format. Use: exec cmd /c <command>\n")
				return nil, nil
			}
			cmd = exec.CommandContext(cmdCtx, cmdName, args...)
		} else {
			cmd = exec.CommandContext(cmdCtx, cmdName, args...)
		}
	} else {
		cmd = exec.CommandContext(cmdCtx, cmdName, args...)
	}

	// 设置命令执行环境
	cmd.Env = os.Environ()

	// 直接将命令的输出重定向到writer
	cmd.Stdout = writer
	cmd.Stderr = writer

	// 启动命令
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(writer, "Failed to start command: %v\n", err)
		return nil, nil
	}

	// 使用channel等待命令完成
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
		close(done)
	}()

	// 等待命令完成或超时
	select {
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// 命令执行失败但是正常退出
				fmt.Fprintf(writer, "Command failed with exit code: %d\n", exitErr.ExitCode())
				return nil, nil
			}
			// 其他错误
			fmt.Fprintf(writer, "Command failed with error: %v\n", err)
			return nil, nil
		}
		// 成功执行
		fmt.Fprintf(writer, "Command completed successfully\n")
		return nil, nil

	case <-cmdCtx.Done():
		// 命令执行超时，确保清理
		if cmd.Process != nil {
			if runtime.GOOS == "windows" {
				exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(cmd.Process.Pid)).Run()
			} else {
				cmd.Process.Signal(syscall.SIGTERM)
				time.Sleep(time.Second)
				cmd.Process.Kill()
			}
		}
		fmt.Fprintf(writer, "Command timed out after 30 seconds\n")
		return nil, nil
	}
}
