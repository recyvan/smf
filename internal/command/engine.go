package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// LocalEngine 本地引擎实现
type LocalEngine struct {
	CmdRegistry *Registry
	AutoReg     *AutoRegister
	timeout     time.Duration
}

// NewLocalEngine 创建新的本地引擎实例
func NewLocalEngine() *LocalEngine {
	return &LocalEngine{
		CmdRegistry: NewRegistry(),
		AutoReg:     NewAutoRegister(),
		timeout:     30 * time.Second,
	}
}

// RegisterCommand 注册命令
func (e *LocalEngine) RegisterCommand(cmd Ecommand) {
	e.CmdRegistry.Register(cmd)
}

// Execute 执行命令
func (e *LocalEngine) Execute(rw io.ReadWriter, input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd, args := parseLocalCommand(input)
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}

	ecommand, exists := e.CmdRegistry.Get(cmd)
	if !exists {
		return "", fmt.Errorf("command not found: %s", cmd)
	}

	// 创建本地IO处理器
	//localIO := NewLocalIO()

	result, err := ecommand.Handler(rw, ctx, args)
	if err != nil {
		if cmdErr, ok := err.(*Error); ok {
			errorData := map[string]interface{}{
				"code":    cmdErr.Code,
				"message": cmdErr.Message,
				"details": cmdErr.Details,
			}
			jsonData, _ := json.Marshal(errorData)
			return fmt.Sprintf("ERROR %s", string(jsonData)), nil
		}
		return "", err
	}

	return string(result), nil
}

func parseLocalCommand(input string) (string, []string) {
	parts := strings.Split(strings.TrimSpace(input), " ")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// LocalIO 实现本地IO操作
type LocalIO struct {
	output strings.Builder
}

func NewLocalIO() *LocalIO {
	return &LocalIO{}
}

func (l *LocalIO) Write(p []byte) (n int, err error) {
	return l.output.Write(p)
}

func (l *LocalIO) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (l *LocalIO) GetOutput() string {
	return l.output.String()
}
