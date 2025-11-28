package plugin

import (
	"gpma/framework/data"
)

type PluginResult struct {
	ASTModified  bool
	MapModified  bool
	Error        error
}

func NewPluginResult() *PluginResult {
	return &PluginResult{
		ASTModified: false,
		MapModified: false,
		Error:       nil,
	}
}

func (r *PluginResult) WithASTModified() *PluginResult {
	r.ASTModified = true
	return r
}

func (r *PluginResult) WithMapModified() *PluginResult {
	r.MapModified = true
	return r
}

func (r *PluginResult) WithError(err error) *PluginResult {
	r.Error = err
	return r
}

func (r *PluginResult) HasModifications() bool {
	return r.ASTModified || r.MapModified
}

type Plugin interface {
	Name() string
	Process(astData *data.ASTData) *PluginResult
}