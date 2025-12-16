package front

import (
	"encoding/json"
	"fmt"
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"

	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// JSONNode
type JSONNode struct {
	Kind     string     `json:"kind"`               // "rule" or "terminal" or "other"
	Rule     string     `json:"rule,omitempty"`     // rule name
	Type     string     `json:"type"`               // Go-type
	Text     string     `json:"text,omitempty"`     // node text
	Span     *JSONSpan  `json:"span,omitempty"`     // span
	Children []JSONNode `json:"children,omitempty"` // children
}

type JSONSpan struct {
	Start JSONPos `json:"start"`
	Stop  JSONPos `json:"stop"`
}

type JSONPos struct {
	Line       int `json:"line"`
	Col        int `json:"col"`
	TokenIndex int `json:"tokenIndex"`
}

// DumpTreeAsJSON
func DumpTreeAsJSON(parser *acsl.ACSLParser, tokens antlr.TokenStream, tree antlr.Tree) (string, error) {
	root := buildJSONNode(parser, tokens, tree)

	b, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func buildJSONNode(p antlr.Parser, toks antlr.TokenStream, n antlr.Tree) JSONNode {
	switch x := n.(type) {

	case antlr.ParserRuleContext:
		rule := ruleName(p, x.GetRuleIndex())

		start := x.GetStart()
		stop := x.GetStop()

		node := JSONNode{
			Kind: "rule",
			Rule: rule,
			Type: fmt.Sprintf("%T", x),
			Text: normalizeWS(nodeText(toks, x)),
		}

		if start != nil && stop != nil {
			node.Span = &JSONSpan{
				Start: JSONPos{Line: start.GetLine(), Col: start.GetColumn(), TokenIndex: start.GetTokenIndex()},
				Stop:  JSONPos{Line: stop.GetLine(), Col: stop.GetColumn(), TokenIndex: stop.GetTokenIndex()},
			}
		}

		for i := 0; i < n.GetChildCount(); i++ {
			ch := n.GetChild(i)
			node.Children = append(node.Children, buildJSONNode(p, toks, ch))
		}
		return node

	case antlr.TerminalNode:
		tok := x.GetSymbol()
		node := JSONNode{
			Kind: "terminal",
			Type: fmt.Sprintf("%T", x),
			Text: tok.GetText(),
		}
		if tok != nil {
			node.Span = &JSONSpan{
				Start: JSONPos{Line: tok.GetLine(), Col: tok.GetColumn(), TokenIndex: tok.GetTokenIndex()},
				Stop:  JSONPos{Line: tok.GetLine(), Col: tok.GetColumn() + len([]rune(tok.GetText())), TokenIndex: tok.GetTokenIndex()},
			}
		}
		return node

	default:
		node := JSONNode{
			Kind: "other",
			Type: fmt.Sprintf("%T", n),
			Text: "",
		}
		for i := 0; i < n.GetChildCount(); i++ {
			ch := n.GetChild(i)
			node.Children = append(node.Children, buildJSONNode(p, toks, ch))
		}
		return node
	}
}

func ruleName(p antlr.Parser, ruleIndex int) string {
	names := p.GetRuleNames()
	if ruleIndex >= 0 && ruleIndex < len(names) {
		return names[ruleIndex]
	}
	return "<?>"
}

func nodeText(toks antlr.TokenStream, ctx antlr.ParserRuleContext) string {
	if ctx.GetStart() == nil || ctx.GetStop() == nil {
		return ctx.GetText()
	}
	a := ctx.GetStart().GetTokenIndex()
	b := ctx.GetStop().GetTokenIndex()
	if a < 0 || b < 0 || b < a {
		return ctx.GetText()
	}
	return toks.GetTextFromInterval(antlr.NewInterval(a, b))
}

func normalizeWS(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}

func safeFileName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unnamed"
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	s = re.ReplaceAllString(s, "_")
	s = strings.Trim(s, "._-")
	if s == "" {
		return "unnamed"
	}
	return s
}

func writeJSONForBlock(outDir string, block RawSpecBlock, jsonStr string, idx int) (string, error) {
	base := "block"

	name := safeFileName(base) + "_" + safeFileName(strconv.Itoa(idx)) + ".json"
	path := filepath.Join(outDir, name)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(jsonStr), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

