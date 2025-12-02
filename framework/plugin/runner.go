package plugin

import (
	"fmt"
	"gpma/framework/data"
)

type RunnerResult struct {
	HasModifications bool
	Errors           []error
}

type PluginRunner struct {
	plugins []Plugin
}

func NewPluginRunner(plugins []Plugin) *PluginRunner {
	return &PluginRunner{
		plugins: plugins,
	}
}

func (r *PluginRunner) Run(astData *data.ASTData) (*RunnerResult, error) {
	result := &RunnerResult{
		HasModifications: false,
		Errors:           make([]error, 0),
	}

	for _, plugin := range r.plugins {
		snapshot, err := astData.CreateSnapshot()
		if err != nil {
			return nil, fmt.Errorf("failed to create snapshot before plugin %s: %w", plugin.Name(), err)
		}

		pluginResult := plugin.Process(astData)

		if pluginResult.Error != nil {
			err := astData.RestoreFromSnapshot(snapshot)
			if err != nil {
				return nil, fmt.Errorf("failed to restore snapshot after plugin %s error: %w", plugin.Name(), err)
			}
			result.Errors = append(result.Errors, fmt.Errorf("plugin %s failed: %w", plugin.Name(), pluginResult.Error))
			continue
		}

		if pluginResult.HasModifications() {
			result.HasModifications = true
		}
	}

	return result, nil
}
