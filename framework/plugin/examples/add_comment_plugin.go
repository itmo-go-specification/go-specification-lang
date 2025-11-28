package examples

import (
	"go/ast"

	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("add_comment", func() plugin.Plugin {
		return NewAddCommentPlugin()
	})
}

type AddCommentPlugin struct{}

func NewAddCommentPlugin() *AddCommentPlugin {
	return &AddCommentPlugin{}
}

func (p *AddCommentPlugin) Name() string {
	return "AddCommentPlugin"
}

func (p *AddCommentPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	ast.Inspect(astData.File, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Doc == nil {
				funcDecl.Doc = &ast.CommentGroup{
					List: []*ast.Comment{
						{
							Slash: funcDecl.Pos() - 1,
							Text:  "// Auto-generated comment",
						},
					},
				}
				result.WithASTModified()
			}

			info := astData.GetNodeInfo(funcDecl)
			info.Data["has_comment"] = funcDecl.Doc != nil
			result.WithMapModified()
		}
		return true
	})

	return result
}