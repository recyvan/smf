package command

import (
	"sync"
)

// CommandProvider 命令提供者接口
type CommandProvider interface {
	ProvideCommands() []Ecommand
}

// AutoRegister 自动注册器
type AutoRegister struct {
	providers []CommandProvider
	mu        sync.Mutex
}

func NewAutoRegister() *AutoRegister {
	return &AutoRegister{
		providers: make([]CommandProvider, 0),
	}
}

// AddProvider 添加命令提供者
func (ar *AutoRegister) AddProvider(p CommandProvider) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.providers = append(ar.providers, p)
}

// RegisterAll 注册所有命令
func (ar *AutoRegister) RegisterAll(registry *Registry) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	for _, provider := range ar.providers {
		for _, cmd := range provider.ProvideCommands() {
			registry.Register(cmd)
		}
	}
}

// ProviderCount 获取命令提供者的数量
func (ar *AutoRegister) ProviderCount() int {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return len(ar.providers)
}
