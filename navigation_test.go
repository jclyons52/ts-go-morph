package tsmorph

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const navigationFixture = `import defaultThing, { a, b as renamed } from "./mod";
import * as ns from "./ns";
import "./side-effect";

export { a } from "./mod";
export * from "./ns";

export interface Shape {
  name: string;
  area(): number;
}

export interface Circle extends Shape, HasRadius {
  radius: number;
}

export class Base {
  protected id: number = 1;
}

export class CircleImpl extends Base implements Shape {
  radius: number;
  private label: string = "c";

  constructor(radius: number) {
    super();
    this.radius = radius;
  }

  area(): number {
    return Math.PI * this.radius * this.radius;
  }

  describe(prefix: string, loud?: boolean): string {
    return prefix + "circle";
  }
}

export enum Color {
  Red,
  Green = 2,
  Blue = 4,
}

export type Point = { x: number; y: number };

export const origin: Point = { x: 0, y: 0 };
let mutable = 5;

export function distance(x1: number, y1: number): number {
  return Math.sqrt(x1 * x1 + y1 * y1);
}

export default function main(): void {}
`

func navigationProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.ts")
	if err := os.WriteFile(path, []byte(navigationFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	sf := p.SourceFile(path)
	if sf == nil {
		t.Fatal("fixture.ts not found in project")
	}
	return p, sf
}

func TestNavigationTopLevel(t *testing.T) {
	_, sf := navigationProject(t)

	if got := len(sf.Classes()); got != 2 {
		t.Fatalf("Classes: got %d, want 2", got)
	}
	if got := len(sf.Interfaces()); got != 2 {
		t.Fatalf("Interfaces: got %d, want 2", got)
	}
	if got := len(sf.Functions()); got != 2 {
		t.Fatalf("Functions: got %d, want 2 (incl. default)", got)
	}
	if got := len(sf.Enums()); got != 1 {
		t.Fatalf("Enums: got %d, want 1", got)
	}
	if got := len(sf.TypeAliases()); got != 1 {
		t.Fatalf("TypeAliases: got %d, want 1", got)
	}
	if got := len(sf.VariableStatements()); got != 2 {
		t.Fatalf("VariableStatements: got %d, want 2", got)
	}
	if got := len(sf.ImportDeclarations()); got != 3 {
		t.Fatalf("ImportDeclarations: got %d, want 3", got)
	}
	if got := len(sf.ExportDeclarations()); got != 2 {
		t.Fatalf("ExportDeclarations: got %d, want 2", got)
	}
}

func TestClassDeclarationAccessors(t *testing.T) {
	_, sf := navigationProject(t)

	c, ok := sf.Class("CircleImpl")
	if !ok {
		t.Fatal("class CircleImpl not found")
	}
	if !c.IsExported() {
		t.Fatal("CircleImpl should be exported")
	}
	if got := c.Extends(); got != "Base" {
		t.Fatalf("Extends: got %q", got)
	}
	if got := c.Implements(); !slices.Equal(got, []string{"Shape"}) {
		t.Fatalf("Implements: got %v", got)
	}

	methods := c.Methods()
	if len(methods) != 2 {
		t.Fatalf("Methods: got %d, want 2", len(methods))
	}
	if methods[0].Name() != "area" || methods[1].Name() != "describe" {
		t.Fatalf("method names: %q, %q", methods[0].Name(), methods[1].Name())
	}

	rt, ok := methods[0].ReturnTypeNode()
	if !ok || rt.Text() != "number" {
		t.Fatalf("area return type: %v %q", ok, rt.Text())
	}

	params := methods[1].Parameters()
	if len(params) != 2 {
		t.Fatalf("describe params: got %d", len(params))
	}
	if params[0].Name() != "prefix" || params[1].Name() != "loud" {
		t.Fatalf("param names: %q, %q", params[0].Name(), params[1].Name())
	}
	if params[0].IsOptional() || !params[1].IsOptional() {
		t.Fatal("IsOptional wrong for describe params")
	}
	pt, ok := params[0].TypeNode()
	if !ok || pt.Text() != "string" {
		t.Fatalf("prefix type: %v %q", ok, pt.Text())
	}

	props := c.Properties()
	if len(props) != 2 {
		t.Fatalf("Properties: got %d, want 2", len(props))
	}
	if props[0].Name() != "radius" || props[1].Name() != "label" {
		t.Fatalf("property names: %q, %q", props[0].Name(), props[1].Name())
	}
	init, ok := props[1].Initializer()
	if !ok || init.Text() != `"c"` {
		t.Fatalf("label initializer: %v %q", ok, init.Text())
	}

	if got := len(c.Constructors()); got != 1 {
		t.Fatalf("Constructors: got %d", got)
	}

	// Text round-trip of the whole class declaration. Modifiers (export)
	// are part of the node, matching TypeScript AST semantics.
	text := c.Text()
	if len(text) == 0 || text[:12] != "export class" {
		t.Fatalf("class text should start with 'export class', got prefix %q", text[:min(12, len(text))])
	}
}

func TestInterfaceAndEnumAndAlias(t *testing.T) {
	_, sf := navigationProject(t)

	iface, ok := sf.Interface("Circle")
	if !ok {
		t.Fatal("interface Circle not found")
	}
	if got := iface.Extends(); !slices.Equal(got, []string{"Shape", "HasRadius"}) {
		t.Fatalf("Extends: got %v", got)
	}
	if got := len(iface.Members()); got != 1 {
		t.Fatalf("Members: got %d, want 1", got)
	}

	enums := sf.Enums()
	color := enums[0]
	if color.Name() != "Color" {
		t.Fatalf("enum name: %q", color.Name())
	}
	members := color.Members()
	if len(members) != 3 {
		t.Fatalf("enum members: got %d", len(members))
	}
	if members[1].Name() != "Green" {
		t.Fatalf("member name: %q", members[1].Name())
	}
	init, ok := members[1].Initializer()
	if !ok || init.Text() != "2" {
		t.Fatalf("Green initializer: %v %q", ok, init.Text())
	}

	aliases := sf.TypeAliases()
	if aliases[0].Name() != "Point" {
		t.Fatalf("alias name: %q", aliases[0].Name())
	}
	tn, ok := aliases[0].TypeNode()
	if !ok || tn.Text() != "{ x: number; y: number }" {
		t.Fatalf("alias type: %v %q", ok, tn.Text())
	}
}

func TestVariablesAndFunctions(t *testing.T) {
	_, sf := navigationProject(t)

	vs := sf.VariableStatements()
	if vs[0].DeclarationKind() != "const" || vs[1].DeclarationKind() != "let" {
		t.Fatalf("declaration kinds: %q, %q", vs[0].DeclarationKind(), vs[1].DeclarationKind())
	}
	decls := vs[0].Declarations()
	if len(decls) != 1 || decls[0].Name() != "origin" {
		t.Fatalf("declarations: %v", decls)
	}
	tn, ok := decls[0].TypeNode()
	if !ok || tn.Text() != "Point" {
		t.Fatalf("origin type: %v %q", ok, tn.Text())
	}

	fn, ok := sf.Function("distance")
	if !ok {
		t.Fatal("function distance not found")
	}
	if got := len(fn.Parameters()); got != 2 {
		t.Fatalf("params: got %d", got)
	}
	if _, ok := fn.Body(); !ok {
		t.Fatal("distance should have a body")
	}

	// The default-exported function has a name too.
	var defFn FunctionDeclaration
	for _, f := range sf.Functions() {
		if f.IsDefaultExport() {
			defFn = f
		}
	}
	if defFn.Name() != "main" {
		t.Fatalf("default function: %q", defFn.Name())
	}
}

func TestImportsAndExports(t *testing.T) {
	_, sf := navigationProject(t)

	imports := sf.ImportDeclarations()
	if imports[0].ModuleSpecifier() != "./mod" {
		t.Fatalf("specifier: %q", imports[0].ModuleSpecifier())
	}
	if imports[0].DefaultImport() != "defaultThing" {
		t.Fatalf("default import: %q", imports[0].DefaultImport())
	}
	if got := imports[0].NamedImports(); !slices.Equal(got, []string{"a", "renamed"}) {
		t.Fatalf("named imports: %v", got)
	}
	if imports[1].NamespaceImport() != "ns" {
		t.Fatalf("namespace import: %q", imports[1].NamespaceImport())
	}
	if imports[2].ModuleSpecifier() != "./side-effect" || imports[2].DefaultImport() != "" {
		t.Fatal("side-effect import parsed wrong")
	}

	exports := sf.ExportDeclarations()
	if got := exports[0].NamedExports(); !slices.Equal(got, []string{"a"}) {
		t.Fatalf("named exports: %v", got)
	}
	if exports[0].ModuleSpecifier() != "./mod" {
		t.Fatalf("re-export specifier: %q", exports[0].ModuleSpecifier())
	}
	if !exports[1].IsNamespaceExport() {
		t.Fatal("expected namespace export")
	}
}

func TestTreeNavigation(t *testing.T) {
	_, sf := navigationProject(t)

	// DescendantsOfKind finds nested nodes too.
	ids := sf.DescendantsOfKind(KindIdentifier)
	if len(ids) == 0 {
		t.Fatal("expected identifier descendants")
	}

	// Find the `area` method via the class, then walk back up to the class.
	c, _ := sf.Class("CircleImpl")
	area := c.Methods()[0]
	anc, ok := area.FirstAncestorByKind(KindClassDeclaration)
	if !ok || anc.Name() != "CircleImpl" {
		t.Fatalf("ancestor: %v %q", ok, anc.Name())
	}

	parent, ok := area.Parent()
	if !ok || parent.Kind() != KindClassDeclaration {
		t.Fatalf("parent kind: %v", parent.KindName())
	}

	// Positions and text.
	line, col := area.StartLineAndColumn()
	if line != 30 { // "area(): number {" line in fixture
		t.Fatalf("area line: got %d, want 30", line)
	}
	if col <= 0 {
		t.Fatalf("area column: got %d", col)
	}

	// FirstDescendantByKind from root.
	if n, ok := sf.FirstDescendantByKind(KindEnumDeclaration); !ok || n.Name() != "Color" {
		t.Fatalf("first enum: %v %q", ok, n.Name())
	}
}
