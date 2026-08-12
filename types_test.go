package tsmorph

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const typesFixture = `export interface Point {
  x: number;
  y: number;
}

export function makePoint(x: number, y: number): Point {
  return { x, y };
}

export const origin = makePoint(0, 0);

export type Shape = Circle | Square;

export interface Circle { kind: "circle"; radius: number }
export interface Square { kind: "square"; size: number }

export function describe(s: Shape): string {
  return s.kind;
}

export const maybeName: string | null = null;
export const count: number = "oops";
`

func typesProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "types.ts")
	if err := os.WriteFile(path, []byte(typesFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func TestNodeType(t *testing.T) {
	_, sf := typesProject(t)

	// Type of a variable with inferred type.
	var origin Node
	for _, vs := range sf.VariableStatements() {
		for _, d := range vs.Declarations() {
			if d.Name() == "origin" {
				origin = d.Node
			}
		}
	}
	if origin.node == nil {
		t.Fatal("origin declaration not found")
	}
	ty := origin.Type()
	if got := ty.Text(); got != "Point" {
		t.Fatalf("origin type text: got %q, want %q", got, "Point")
	}
	if ty.IsString() || ty.IsNumber() || ty.IsUnion() {
		t.Fatal("Point should not be string/number/union")
	}

	// Properties of the Point type.
	props := ty.Properties()
	names := make([]string, len(props))
	for i, p := range props {
		names[i] = p.Name
	}
	if !slices.Equal(names, []string{"x", "y"}) {
		t.Fatalf("Point properties: %v", names)
	}
	if !props[0].Type.IsNumber() {
		t.Fatalf("x should be number, got %q", props[0].Type.Text())
	}
}

func TestFunctionTypeAndSignatures(t *testing.T) {
	_, sf := typesProject(t)

	fn, ok := sf.Function("describe")
	if !ok {
		t.Fatal("describe not found")
	}
	ty := fn.Type()
	sigs := ty.CallSignatures()
	if len(sigs) != 1 {
		t.Fatalf("call signatures: got %d", len(sigs))
	}
	if got := sigs[0].ReturnType().Text(); got != "string" {
		t.Fatalf("describe return type: %q", got)
	}

	// Union type.
	var maybeName Node
	for _, vs := range sf.VariableStatements() {
		for _, d := range vs.Declarations() {
			if d.Name() == "maybeName" {
				maybeName = d.Node
			}
		}
	}
	ut := maybeName.Type()
	if !ut.IsUnion() {
		t.Fatalf("maybeName should be union, got %q", ut.Text())
	}
	parts := ut.UnionTypes()
	if len(parts) != 2 {
		t.Fatalf("union constituents: got %d", len(parts))
	}
}

func TestInterfaceTypePredicate(t *testing.T) {
	_, sf := typesProject(t)

	iface, ok := sf.Interface("Point")
	if !ok {
		t.Fatal("Point interface not found")
	}
	ty := iface.Type()
	if !ty.IsInterface() {
		t.Fatalf("Point should be an interface type, got %q", ty.Text())
	}
}

func TestNodeSymbol(t *testing.T) {
	_, sf := typesProject(t)

	fn, ok := sf.Function("makePoint")
	if !ok {
		t.Fatal("makePoint not found")
	}
	sym, ok := fn.Symbol()
	if !ok {
		t.Fatal("makePoint has no symbol")
	}
	if sym.Name() != "makePoint" {
		t.Fatalf("symbol name: %q", sym.Name())
	}
	decls := sym.Declarations()
	if len(decls) != 1 {
		t.Fatalf("declarations: got %d", len(decls))
	}
	if decls[0].Kind() != KindFunctionDeclaration {
		t.Fatalf("declaration kind: %v", decls[0].KindName())
	}
	if got := sym.Type().Text(); got != "(x: number, y: number) => Point" {
		t.Fatalf("symbol type: %q", got)
	}
}

func TestPreEmitDiagnostics(t *testing.T) {
	p, sf := typesProject(t)

	diags := p.PreEmitDiagnostics()
	var found *Diagnostic
	for i := range diags {
		if diags[i].Code == 2322 { // Type 'X' is not assignable to type 'Y'
			found = &diags[i]
		}
	}
	if found == nil {
		t.Fatalf("expected TS2322 diagnostic, got %v", diags)
	}
	if found.Category != DiagnosticCategoryError {
		t.Fatalf("category: %v", found.Category)
	}
	df, ok := found.SourceFile()
	if !ok || df.FilePath() != sf.FilePath() {
		t.Fatal("diagnostic should point at types.ts")
	}
	line, _, ok := found.LineAndColumn()
	if !ok || line != 22 { // `export const count: number = "oops";`
		t.Fatalf("diagnostic line: %d", line)
	}
}
