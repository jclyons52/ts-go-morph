package tsmorph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const introspectFixture = `export enum Color { Red = "red", Blue = "blue" }

export type UserId = string;

export type PoolConfig = { host: string; port: number };

export class Pool {
  constructor(config: PoolConfig) {}
}

export interface Config {
  url: string;
  retries?: number;
}

export interface RecordLike extends Record<string, unknown> {}

export abstract class Base {
  abstract name: string;
}

export class Widget {
  private constructor() {}

  static create(): Widget {
    return new Widget();
  }
}

export class Service {
  constructor(private config: Config, count: number = 3) {}
}

export class Derived extends Base {
  name = "derived";
}

export const nums: number[] = [1, 2, 3];

export const maybe: string | null = null;

export function useIt({ config }: { config: Config }) {
  return config.url;
}
`

func introspectProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "introspect.ts")
	if err := os.WriteFile(path, []byte(introspectFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func nodeByName(sf *SourceFile, vsName, declName string) Node {
	for _, vs := range sf.VariableStatements() {
		for _, d := range vs.Declarations() {
			if d.Name() == declName && vs.Text() != "" {
				return d.Node
			}
		}
	}
	_ = vsName
	return Node{}
}

func TestTypeClassObjectArrayEnum(t *testing.T) {
	_, sf := introspectProject(t)

	widget, ok := sf.Class("Widget")
	if !ok {
		t.Fatal("Widget not found")
	}
	wt := widget.Type()
	if !wt.IsClass() {
		t.Fatalf("Widget should be a class, got %q", wt.Text())
	}
	if !wt.IsObject() {
		t.Fatal("Widget should be an object type")
	}

	base, _ := sf.Class("Base")
	if !base.IsAbstract() {
		t.Fatal("Base should be abstract")
	}

	nums := nodeByName(sf, "", "nums")
	if !nums.Type().IsArray() {
		t.Fatalf("nums should be array, got %q", nums.Type().Text())
	}

	en, ok := sf.Enums()[0], true
	if !ok {
		t.Fatal("no enum")
	}
	_ = en
}

func TestTypeNullableAndNonNullable(t *testing.T) {
	_, sf := introspectProject(t)

	maybe := nodeByName(sf, "", "maybe")
	mt := maybe.Type()
	if !mt.IsNullable() {
		t.Fatalf("maybe should be nullable, got %q", mt.Text())
	}
	if !mt.IsUnion() {
		t.Fatalf("maybe should be union, got %q", mt.Text())
	}
	if got := mt.GetNonNullableType().Text(); got != "string" {
		t.Fatalf("non-nullable maybe: %q", got)
	}
}

func TestTypeAliasSymbolAndLiteral(t *testing.T) {
	_, sf := introspectProject(t)

	// The alias declaration's own type resolves to its target ("string").
	userId := sf.TypeAliases()[0]
	if got := userId.Type().Text(); got != "string" {
		t.Fatalf("UserId target type text: %q", got)
	}

	// A parameter typed with a type alias carries the alias symbol (and its
	// text is the alias name).
	pool, ok := sf.Class("Pool")
	if !ok {
		t.Fatal("Pool not found")
	}
	param := pool.GetConstructors()[0].Parameters()[0]
	ty := param.Type()
	if got := ty.Text(); got != "PoolConfig" {
		t.Fatalf("Pool ctor param type text: %q", got)
	}
	sym, ok := ty.AliasSymbol()
	if !ok || sym.Name() != "PoolConfig" {
		t.Fatalf("Pool ctor param alias symbol: %v %q", ok, sym.Name())
	}
	// The underlying (anonymous) symbol should escape to __type.
	if underlying, ok := ty.Symbol(); ok {
		if underlying.Name() != "__type" {
			t.Fatalf("underlying symbol name: %q", underlying.Name())
		}
	}

	en := sf.Enums()[0]
	if len(en.Members()) != 2 {
		t.Fatalf("enum members: %d", len(en.Members()))
	}
}

func TestStringIndexType(t *testing.T) {
	_, sf := introspectProject(t)

	rl, ok := sf.Interface("RecordLike")
	if !ok {
		t.Fatal("RecordLike not found")
	}
	ty := rl.Type()
	idx, ok := ty.StringIndexType()
	if !ok {
		t.Fatal("RecordLike should have a string index type")
	}
	if got := idx.Text(); got != "unknown" {
		t.Fatalf("index type: %q", got)
	}
}

func TestBaseTypes(t *testing.T) {
	_, sf := introspectProject(t)

	derived, ok := sf.Class("Derived")
	if !ok {
		t.Fatal("Derived not found")
	}
	bases := derived.Type().BaseTypes()
	if len(bases) != 1 || bases[0].Text() != "Base" {
		t.Fatalf("Derived base types: %v", bases)
	}
}

func TestConstructorAndStaticFactory(t *testing.T) {
	_, sf := introspectProject(t)

	widget, _ := sf.Class("Widget")
	ctors := widget.GetConstructors()
	if len(ctors) != 1 {
		t.Fatalf("Widget constructors: %d", len(ctors))
	}
	if !ctors[0].HasModifierPrivate() {
		t.Fatal("Widget constructor should be private")
	}

	create, ok := widget.GetStaticMethod("create")
	if !ok {
		t.Fatal("Widget.create not found")
	}
	if got := create.ReturnType().Text(); got != "Widget" {
		t.Fatalf("Widget.create return type: %q", got)
	}

	service, _ := sf.Class("Service")
	svcCtors := service.GetConstructors()
	if len(svcCtors) != 1 {
		t.Fatalf("Service constructors: %d", len(svcCtors))
	}
	params := svcCtors[0].Parameters()
	if len(params) != 2 {
		t.Fatalf("Service ctor params: %d", len(params))
	}
	if !params[0].HasModifierPrivate() {
		t.Fatal("Service config param should be private")
	}
	if !params[1].HasInitializer() {
		t.Fatal("Service count param should have initializer")
	}
}

func TestDestructuredParameter(t *testing.T) {
	_, sf := introspectProject(t)

	useIt, ok := sf.Function("useIt")
	if !ok {
		t.Fatal("useIt not found")
	}
	params := useIt.Parameters()
	if len(params) != 1 {
		t.Fatalf("useIt params: %d", len(params))
	}
	if !params[0].IsDestructured() {
		t.Fatal("useIt param should be destructured")
	}
}

const jsxFixture = `import React from "react";

export function Card() {
  return (
    <div className="card">
      Hello world
      <Button label="Click me" onClick={() => alert("hi")} />
    </div>
  );
}

export const Title = () => <h1>Title text</h1>;

const x = foo("arg1", "arg2");
`

func jsxProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.tsx")
	if err := os.WriteFile(path, []byte(jsxFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func TestJsxTraversal(t *testing.T) {
	_, sf := jsxProject(t)

	texts := sf.DescendantsOfKind(KindJsxText)
	if len(texts) == 0 {
		t.Fatal("expected JSX text nodes")
	}
	foundHello := false
	for _, n := range texts {
		jt, ok := n.AsJsxText()
		if !ok {
			continue
		}
		if strings.TrimSpace(jt.TextContent()) == "Hello world" {
			foundHello = true
		}
	}
	if !foundHello {
		t.Fatal("expected 'Hello world' JSX text")
	}

	attrs := sf.DescendantsOfKind(KindJsxAttribute)
	if len(attrs) == 0 {
		t.Fatal("expected JSX attributes")
	}
	foundLabel := false
	for _, n := range attrs {
		attr, ok := n.AsJsxAttribute()
		if !ok {
			continue
		}
		name, _ := attr.NameNode()
		if name.Text() == "label" {
			init, ok := attr.Initializer()
			if !ok || !init.IsStringLiteral() {
				t.Fatal("label initializer should be a string literal")
			}
			if init.LiteralValue() != "Click me" {
				t.Fatalf("label value: %q", init.LiteralValue())
			}
			foundLabel = true
		}
	}
	if !foundLabel {
		t.Fatal("expected label attribute")
	}
}

func TestExpressionAccessors(t *testing.T) {
	_, sf := jsxProject(t)

	var call Node
	for _, st := range sf.Statements() {
		for _, d := range st.Descendants() {
			if d.IsCallExpression() {
				call = d
			}
		}
	}
	if call.node == nil {
		t.Fatal("expected a call expression")
	}
	expr, ok := call.GetExpression()
	if !ok || !expr.IsIdentifier() || expr.Text() != "foo" {
		t.Fatalf("call callee: %v %q", ok, expr.Text())
	}
	args := call.GetArguments()
	if len(args) != 2 || !args[0].IsStringLiteral() || args[0].LiteralValue() != "arg1" {
		t.Fatalf("call args: %d", len(args))
	}
}

func TestTypeOnlyImportRendering(t *testing.T) {
	p := &Project{}
	w := newCodeWriter(p)
	got := w.importDecl(ImportDeclarationStructure{
		ModuleSpecifier: "foo",
		NamedImports:    []string{"Bar"},
		IsTypeOnly:      true,
	})
	want := `import type { Bar } from "foo";`
	if got != want {
		t.Fatalf("type-only import: got %q, want %q", got, want)
	}
}

func TestBlockInsertStatements(t *testing.T) {
	_, sf := jsxProject(t)

	card, ok := sf.Function("Card")
	if !ok {
		t.Fatal("Card not found")
	}
	body, ok := card.Body()
	if !ok {
		t.Fatal("Card has no body")
	}
	block, ok := body.AsBlock()
	if !ok {
		t.Fatal("Card body should be a block")
	}
	block.InsertStatements(0, `const { t } = useTranslation();`)

	reloaded, _ := sf.Function("Card")
	body2, _ := reloaded.Body()
	stmts := body2.GetStatements()
	if len(stmts) == 0 {
		t.Fatal("expected statements after insert")
	}
	if got := stmts[0].Text(); got != "const { t } = useTranslation();" {
		t.Fatalf("first statement: %q", got)
	}
}

const literalsFixture = `export enum Color { Red = "red", Blue = "blue" }

const str = "hello";
const num = 42;
const yes = true;

export type Handler = (x: number) => string;
export const fn: Handler = (x) => String(x);

export const err: Error = new Error();
`

func literalsProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "literals.ts")
	if err := os.WriteFile(path, []byte(literalsFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func TestLiteralPredicates(t *testing.T) {
	_, sf := literalsProject(t)

	// enum literal (member type)
	en := sf.Enums()[0]
	memberType := en.Members()[0].Type()
	if !memberType.IsEnumLiteral() {
		t.Fatalf("enum member should be enum literal, got %q", memberType.Text())
	}
	if !memberType.IsLiteral() {
		t.Fatal("enum literal should satisfy IsLiteral")
	}

	// string literal
	if str := declType(t, sf, "str"); !str.IsStringLiteral() || !str.IsString() || !str.IsLiteral() {
		t.Fatalf("str should be string literal, got %q", str.Text())
	}
	// number literal
	if num := declType(t, sf, "num"); !num.IsNumberLiteral() || !num.IsNumber() || !num.IsLiteral() {
		t.Fatalf("num should be number literal, got %q", num.Text())
	}
	// boolean literal
	if yes := declType(t, sf, "yes"); !yes.IsBooleanLiteral() || !yes.IsBoolean() || !yes.IsLiteral() {
		t.Fatalf("yes should be boolean literal, got %q", yes.Text())
	}
}

func declType(t *testing.T, sf *SourceFile, name string) Type {
	t.Helper()
	for _, vs := range sf.VariableStatements() {
		for _, d := range vs.Declarations() {
			if d.Name() == name {
				return d.Type()
			}
		}
	}
	t.Fatalf("declaration %q not found", name)
	return Type{}
}

func TestSignatureParameters(t *testing.T) {
	_, sf := literalsProject(t)

	// Use the Handler type alias reference via the fn variable.
	fnVar := declType(t, sf, "fn")
	sigs := fnVar.CallSignatures()
	if len(sigs) != 1 {
		t.Fatalf("fn call signatures: %d", len(sigs))
	}
	params := sigs[0].Parameters()
	if len(params) != 1 {
		t.Fatalf("fn params: %d", len(params))
	}
	if params[0].Name != "x" || !params[0].Type.IsNumber() {
		t.Fatalf("fn param: %q %q", params[0].Name, params[0].Type.Text())
	}
	if !sigs[0].ReturnType().IsString() {
		t.Fatalf("fn return type: %q", sigs[0].ReturnType().Text())
	}
}

func TestBuiltInTypeAndSymbolEqual(t *testing.T) {
	_, sf := literalsProject(t)

	// err: Error is a lib.*.d.ts type.
	errType := declType(t, sf, "err")
	sym, ok := errType.Symbol()
	if !ok {
		t.Fatal("Error type has no symbol")
	}
	decls := sym.Declarations()
	if len(decls) == 0 || !decls[0].SourceFile().IsLibFile() {
		t.Fatal("Error type should be declared in a lib file")
	}

	// Symbol.Equal: the fn variable's symbol equals itself.
	fnSym, ok := declType(t, sf, "fn").Symbol()
	if !ok {
		t.Fatal("fn has no symbol")
	}
	fnSym2, ok := declType(t, sf, "fn").Symbol()
	if !ok || !fnSym.Equal(fnSym2) {
		t.Fatal("identical symbols should compare equal")
	}
}
