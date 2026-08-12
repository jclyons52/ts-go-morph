package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

// Typed node wrappers. Each embeds Node and adds kind-specific accessors.
// Obtain them via Node.AsXxx() or the SourceFile enumeration methods.

// ClassDeclaration wraps a `class` declaration.
type ClassDeclaration struct{ Node }

// Methods returns the class's method declarations (excluding constructors).
func (c ClassDeclaration) Methods() []MethodDeclaration {
	var out []MethodDeclaration
	for _, m := range c.node.Members() {
		if m.Kind == ast.KindMethodDeclaration {
			out = append(out, MethodDeclaration{Node{node: m, sf: c.sf, gen: c.gen}})
		}
	}
	return out
}

// Properties returns the class's property declarations.
func (c ClassDeclaration) Properties() []PropertyDeclaration {
	var out []PropertyDeclaration
	for _, m := range c.node.Members() {
		if m.Kind == ast.KindPropertyDeclaration {
			out = append(out, PropertyDeclaration{Node{node: m, sf: c.sf, gen: c.gen}})
		}
	}
	return out
}

// Constructors returns the class's constructor declarations.
func (c ClassDeclaration) Constructors() []Node {
	var out []Node
	for _, m := range c.node.Members() {
		if m.Kind == ast.KindConstructor {
			out = append(out, Node{node: m, sf: c.sf, gen: c.gen})
		}
	}
	return out
}

// heritageTypes returns the texts of heritage clause types (extends/implements)
// for the given clause token.
func (n Node) heritageTypes(token Kind) []string {
	data := n.node.ClassLikeData()
	if data == nil || data.HeritageClauses == nil {
		return nil
	}
	var out []string
	for _, clause := range data.HeritageClauses.Nodes {
		hc := clause.AsHeritageClause()
		if hc.Token != token {
			continue
		}
		for _, t := range hc.Types.Nodes {
			out = append(out, Node{node: t, sf: n.sf, gen: n.gen}.Text())
		}
	}
	return out
}

// Extends returns the text of the class's `extends` expression, or "".
func (c ClassDeclaration) Extends() string {
	if types := c.heritageTypes(ast.KindExtendsKeyword); len(types) > 0 {
		return types[0]
	}
	return ""
}

// Implements returns the texts of the class's `implements` types.
func (c ClassDeclaration) Implements() []string {
	return c.heritageTypes(ast.KindImplementsKeyword)
}

// InterfaceDeclaration wraps an `interface` declaration.
type InterfaceDeclaration struct{ Node }

// Members returns the interface's member elements.
func (i InterfaceDeclaration) Members() []Node {
	var out []Node
	for _, m := range i.node.Members() {
		out = append(out, Node{node: m, sf: i.sf, gen: i.gen})
	}
	return out
}

// Extends returns the texts of the interface's `extends` types.
func (i InterfaceDeclaration) Extends() []string {
	data := i.node.AsInterfaceDeclaration()
	if data.HeritageClauses == nil {
		return nil
	}
	var out []string
	for _, clause := range data.HeritageClauses.Nodes {
		for _, t := range clause.AsHeritageClause().Types.Nodes {
			out = append(out, Node{node: t, sf: i.sf, gen: i.gen}.Text())
		}
	}
	return out
}

// FunctionDeclaration wraps a top-level `function` declaration.
type FunctionDeclaration struct{ Node }

// Parameters returns the function's parameters.
func (f FunctionDeclaration) Parameters() []ParameterDeclaration {
	return parameterNodes(f.Node)
}

// ReturnTypeNode returns the declared return type node, or false for an
// inferred return type.
func (f FunctionDeclaration) ReturnTypeNode() (Node, bool) {
	return f.sf.wrap(f.node.Type())
}

// Body returns the function body block, or false for declarations without a
// body (overloads, ambient declarations).
func (f FunctionDeclaration) Body() (Node, bool) {
	return f.sf.wrap(f.node.Body())
}

// ConstructorDeclaration wraps a class constructor.
type ConstructorDeclaration struct{ Node }

// Parameters returns the constructor's parameters.
func (c ConstructorDeclaration) Parameters() []ParameterDeclaration {
	return parameterNodes(c.Node)
}

// MethodDeclaration wraps a class or object method.
type MethodDeclaration struct{ Node }

// Parameters returns the method's parameters.
func (m MethodDeclaration) Parameters() []ParameterDeclaration {
	return parameterNodes(m.Node)
}

// ReturnTypeNode returns the declared return type node, or false.
func (m MethodDeclaration) ReturnTypeNode() (Node, bool) {
	return m.sf.wrap(m.node.Type())
}

// PropertyDeclaration wraps a class property.
type PropertyDeclaration struct{ Node }

// TypeNode returns the declared type node, or false.
func (p PropertyDeclaration) TypeNode() (Node, bool) {
	return p.sf.wrap(p.node.Type())
}

// Initializer returns the initializer expression, or false.
func (p PropertyDeclaration) Initializer() (Node, bool) {
	return p.sf.wrap(p.node.Initializer())
}

// ParameterDeclaration wraps a function/method/constructor parameter.
type ParameterDeclaration struct{ Node }

// TypeNode returns the declared type node, or false.
func (p ParameterDeclaration) TypeNode() (Node, bool) {
	return p.sf.wrap(p.node.Type())
}

// IsOptional reports whether the parameter has a `?`.
func (p ParameterDeclaration) IsOptional() bool {
	return p.node.QuestionToken() != nil
}

// parameterNodes extracts wrapped parameters from a function-like node.
func parameterNodes(n Node) []ParameterDeclaration {
	var out []ParameterDeclaration
	for _, p := range n.node.Parameters() {
		out = append(out, ParameterDeclaration{Node{node: p, sf: n.sf, gen: n.gen}})
	}
	return out
}

// EnumDeclaration wraps an `enum` declaration.
type EnumDeclaration struct{ Node }

// EnumMember wraps a member of an enum declaration.
type EnumMember struct{ Node }

// Initializer returns the member's initializer expression, or false.
func (e EnumMember) Initializer() (Node, bool) {
	return e.sf.wrap(e.node.Initializer())
}

// Members returns the enum's members.
func (e EnumDeclaration) Members() []EnumMember {
	var out []EnumMember
	for _, m := range e.node.AsEnumDeclaration().Members.Nodes {
		out = append(out, EnumMember{Node{node: m, sf: e.sf, gen: e.gen}})
	}
	return out
}

// TypeAliasDeclaration wraps a `type X = ...` declaration.
type TypeAliasDeclaration struct{ Node }

// TypeNode returns the aliased type node.
func (t TypeAliasDeclaration) TypeNode() (Node, bool) {
	return t.sf.wrap(t.node.AsTypeAliasDeclaration().Type)
}

// VariableStatement wraps a `const`/`let`/`var` statement.
type VariableStatement struct{ Node }

// Declarations returns the statement's variable declarations.
func (v VariableStatement) Declarations() []VariableDeclaration {
	var out []VariableDeclaration
	dl := v.node.AsVariableStatement().DeclarationList
	if dl == nil {
		return nil
	}
	for _, d := range dl.AsVariableDeclarationList().Declarations.Nodes {
		out = append(out, VariableDeclaration{Node{node: d, sf: v.sf, gen: v.gen}})
	}
	return out
}

// DeclarationKind returns "const", "let", or "var".
func (v VariableStatement) DeclarationKind() string {
	dl := v.node.AsVariableStatement().DeclarationList
	if dl == nil {
		return ""
	}
	flags := dl.AsVariableDeclarationList().Flags
	switch {
	case flags&ast.NodeFlagsConst != 0:
		return "const"
	case flags&ast.NodeFlagsLet != 0:
		return "let"
	default:
		return "var"
	}
}

// VariableDeclaration wraps a single declarator in a variable statement.
type VariableDeclaration struct{ Node }

// TypeNode returns the declared type node, or false.
func (v VariableDeclaration) TypeNode() (Node, bool) {
	return v.sf.wrap(v.node.Type())
}

// Initializer returns the initializer expression, or false.
func (v VariableDeclaration) Initializer() (Node, bool) {
	return v.sf.wrap(v.node.Initializer())
}

// ImportDeclaration wraps an `import ... from "..."` statement.
type ImportDeclaration struct{ Node }

// ModuleSpecifier returns the module path string, e.g. "./util".
func (i ImportDeclaration) ModuleSpecifier() string {
	spec := i.node.ModuleSpecifier()
	if spec == nil {
		return ""
	}
	return spec.Text()
}

// DefaultImport returns the default import binding name, or "".
func (i ImportDeclaration) DefaultImport() string {
	clause := i.node.ImportClause()
	if clause == nil {
		return ""
	}
	if name := clause.Name(); name != nil {
		return name.Text()
	}
	return ""
}

// NamedImports returns the named import bindings, e.g. ["a", "b"] for
// `import { a, b } from "x"`. Aliased imports return the local name.
func (i ImportDeclaration) NamedImports() []string {
	clause := i.node.ImportClause()
	if clause == nil {
		return nil
	}
	bindings := clause.AsImportClause().NamedBindings
	if bindings == nil || bindings.Kind != ast.KindNamedImports {
		return nil
	}
	var out []string
	for _, el := range bindings.AsNamedImports().Elements.Nodes {
		out = append(out, el.Name().Text())
	}
	return out
}

// NamespaceImport returns the namespace binding name for
// `import * as ns from "x"`, or "".
func (i ImportDeclaration) NamespaceImport() string {
	clause := i.node.ImportClause()
	if clause == nil {
		return ""
	}
	bindings := clause.AsImportClause().NamedBindings
	if bindings == nil || bindings.Kind != ast.KindNamespaceImport {
		return ""
	}
	return bindings.Name().Text()
}

// ExportDeclaration wraps an `export ... from "..."` statement.
type ExportDeclaration struct{ Node }

// ModuleSpecifier returns the module path for re-exports, or "".
func (e ExportDeclaration) ModuleSpecifier() string {
	spec := e.node.ModuleSpecifier()
	if spec == nil {
		return ""
	}
	return spec.Text()
}

// NamedExports returns the exported names, e.g. ["a", "b"] for
// `export { a, b }`. Aliased exports return the exported name.
func (e ExportDeclaration) NamedExports() []string {
	clause := e.node.AsExportDeclaration().ExportClause
	if clause == nil || clause.Kind != ast.KindNamedExports {
		return nil
	}
	var out []string
	for _, el := range clause.AsNamedExports().Elements.Nodes {
		out = append(out, el.Name().Text())
	}
	return out
}

// IsNamespaceExport reports whether this is `export * from "x"` or
// `export * as ns from "x"`.
func (e ExportDeclaration) IsNamespaceExport() bool {
	clause := e.node.AsExportDeclaration().ExportClause
	if clause == nil {
		// Plain `export * from "x"` has no export clause.
		return e.node.ModuleSpecifier() != nil
	}
	return clause.Kind == ast.KindNamespaceExport
}
