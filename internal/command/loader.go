package command

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

// PluginLoader 插件加载器
type PluginLoader struct {
	pluginDir string
	commands  []Ecommand
}

// NewPluginLoader 创建新的插件加载器
func NewPluginLoader(pluginDir string) *PluginLoader {
	return &PluginLoader{
		pluginDir: pluginDir,
		commands:  make([]Ecommand, 0),
	}
}

// LoadPlugins 加载所有插件
func (pl *PluginLoader) LoadPlugins() error {
	// 确保插件目录存在
	if err := os.MkdirAll(pl.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %v", err)
	}

	// 遍历插件目录
	err := filepath.Walk(pl.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Warning: error accessing path %s: %v\n", path, err)
			return nil
		}
		// 只处理 .so 文件
		if !info.IsDir() && filepath.Ext(path) == ".so" {
			if err := pl.loadPlugin(path); err != nil {
				fmt.Printf("Warning: failed to load plugins %s: %v\n", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking plugins directory: %v", err)
	}

	return nil
}

// loadPlugin 加载单个插件
func (pl *PluginLoader) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugins: %v", err)
	}

	// 查找 Plugin 符号
	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("failed to find Plugin symbol: %v", err)
	}

	// 类型断言
	provider, ok := symbol.(CommandProvider)
	if !ok {
		return fmt.Errorf("invalid plugins type: Plugin symbol must implement CommandProvider interface")
	}

	// 获取并存储命令
	commands := provider.ProvideCommands()
	pl.commands = append(pl.commands, commands...)

	fmt.Printf("Successfully loaded plugins: %s\n", filepath.Base(path))
	return nil
}

// GetCommands 返回所有加载的命令
func (pl *PluginLoader) GetCommands() []Ecommand {
	return pl.commands
}
