package command

import (
	"context"
	"io"
	"sync"
)

// Handler 命令处理函数
type Handler func(rw io.ReadWriter, ctx context.Context, args []string) ([]byte, error)
type Ecommand struct {
	Name        string
	Description string
	Usage       string
	Type        string
	//是否运行后台执行
	Background bool
	Handler    Handler
}

type Registry struct {
	sync.RWMutex
	Commands map[string]Ecommand
}

func NewRegistry() *Registry {
	return &Registry{
		Commands: make(map[string]Ecommand),
	}
}

func (r *Registry) Register(cmd Ecommand) {
	r.Lock()
	defer r.Unlock()
	r.Commands[cmd.Name] = cmd
}

func (r *Registry) Get(name string) (Ecommand, bool) {
	r.RLock()
	defer r.RUnlock()
	cmd, exists := r.Commands[name]
	return cmd, exists
}

func (r *Registry) List() []Ecommand {
	r.RLock()
	defer r.RUnlock()
	var cmds []Ecommand
	for _, cmd := range r.Commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}
