package tsmorph

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/core"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/parser"
)

// Smoke test proving the vendored typescript-go tree is importable and the
// parser works end to end.
func TestVendoredParserSmoke(t *testing.T) {
	text := "const x: number = 1;"
	sf := parser.ParseSourceFile(
		ast.SourceFileParseOptions{FileName: "/test.ts"},
		text,
		core.ScriptKindTS,
	)
	if sf == nil {
		t.Fatal("expected non-nil SourceFile")
	}
	if sf.Statements == nil || len(sf.Statements.Nodes) != 1 {
		t.Fatalf("expected 1 statement, got %v", sf.Statements)
	}
	stmt := sf.Statements.Nodes[0]
	if stmt.Kind != ast.KindVariableStatement {
		t.Fatalf("expected VariableStatement, got %v", stmt.Kind)
	}
	// Node.Text() only supports identifiers/literals in typescript-go; slice
	// the source text via node positions instead.
	loc := stmt.Loc
	if got := text[loc.Pos():loc.End()]; got != "const x: number = 1;" {
		t.Fatalf("unexpected statement text: %q", got)
	}
}
