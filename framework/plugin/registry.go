package plugin

import (
	"fmt"
	"sync"
)

type PluginFactory func() Plugin

var (
	registry = make(map[string]PluginFactory)
	mu       sync.RWMutex
)

func Register(name string, factory PluginFactory) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = factory
}

func Get(name string) (Plugin, error) {
	mu.RLock()
	defer mu.RUnlock()

	factory, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("plugin %q not found", name)
	}

	return factory(), nil
}

func GetAll(names []string) ([]Plugin, error) {
	plugins := make([]Plugin, 0, len(names))

	for _, name := range names {
		plugin, err := Get(name)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
