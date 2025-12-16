package framework

import (
	"fmt"

	"gpma/framework/codegen"
	"gpma/framework/data"
	"gpma/framework/gosl/bind"
	"gpma/framework/gosl/front"
	"gpma/framework/parser"
	"gpma/framework/plugin"
)

type Framework struct {
	plugins []plugin.Plugin
	parser  *parser.Parser
}

func New() *Framework {
	return &Framework{
		plugins: make([]plugin.Plugin, 0),
		parser:  parser.NewParser(),
	}
}

func (f *Framework) AddPlugin(p plugin.Plugin) *Framework {
	f.plugins = append(f.plugins, p)
	return f
}

func (f *Framework) AddPlugins(plugins ...plugin.Plugin) *Framework {
	f.plugins = append(f.plugins, plugins...)
	return f
}

func (f *Framework) AddPluginByName(name string) (*Framework, error) {
	p, err := plugin.Get(name)
	if err != nil {
		return f, err
	}
	f.plugins = append(f.plugins, p)
	return f, nil
}

func (f *Framework) AddPluginsByNames(names ...string) (*Framework, error) {
	for _, name := range names {
		p, err := plugin.Get(name)
		if err != nil {
			return f, err
		}
		f.plugins = append(f.plugins, p)
	}
	return f, nil
}

func (f *Framework) ListAvailablePlugins() []string {
	return plugin.List()
}

type Result struct {
	GeneratedCode string
	Errors        []error
	Modified      bool
	ASTData       *data.ASTData
}

func (f *Framework) ProcessFile(filepath string) (*Result, error) {
	astData, err := f.parser.ParseFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Collecting commnet block from astData
	blocks := front.CollectRawBlocks(astData.FileSet, astData.File, astData.FileSet.File(astData.File.Pos()).Name())
	// Parsing GOSL -> creating SpecModule
	spec, err := front.BuildSpecModule(blocks)

	// binding ast and specifications together
	astData, err = bind.Bind(astData, spec)

	runner := plugin.NewPluginRunner(f.plugins)
	runnerResult, err := runner.Run(astData)
	if err != nil {
		return nil, fmt.Errorf("plugin runner failed: %w", err)
	}

	result := &Result{
		Errors:   runnerResult.Errors,
		Modified: runnerResult.HasModifications,
		ASTData:  astData,
	}

	if runnerResult.HasModifications {
		gen := codegen.NewGenerator(astData)
		code, err := gen.GenerateToString()
		if err != nil {
			return nil, fmt.Errorf("failed to generate code: %w", err)
		}
		result.GeneratedCode = code
	}

	return result, nil
}

func (f *Framework) ProcessSource(source []byte) (*Result, error) {
	astData, err := f.parser.ParseSource(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	runner := plugin.NewPluginRunner(f.plugins)
	runnerResult, err := runner.Run(astData)
	if err != nil {
		return nil, fmt.Errorf("plugin runner failed: %w", err)
	}

	result := &Result{
		Errors:   runnerResult.Errors,
		Modified: runnerResult.HasModifications,
		ASTData:  astData,
	}

	if runnerResult.HasModifications {
		gen := codegen.NewGenerator(astData)
		code, err := gen.GenerateToString()
		if err != nil {
			return nil, fmt.Errorf("failed to generate code: %w", err)
		}
		result.GeneratedCode = code
	}

	return result, nil
}
