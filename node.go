package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/scanner"
)

// Kind identifies the syntactic kind of a Node. It is an alias of the
// vendored compiler's ast.Kind, so the full set of Kind constants in that
// package can be used interchangeably.
type Kind = ast.Kind

// Commonly used kinds, re-exported for convenience.
const (
	KindClassDeclaration     = ast.KindClassDeclaration
	KindInterfaceDeclaration = ast.KindInterfaceDeclaration
	KindFunctionDeclaration  = ast.KindFunctionDeclaration
	KindMethodDeclaration    = ast.KindMethodDeclaration
	KindPropertyDeclaration  = ast.KindPropertyDeclaration
	KindConstructor          = ast.KindConstructor
	KindParameter            = ast.KindParameter
	KindEnumDeclaration      = ast.KindEnumDeclaration
	KindEnumMember           = ast.KindEnumMember
	KindTypeAliasDeclaration = ast.KindTypeAliasDeclaration
	KindVariableStatement    = ast.KindVariableStatement
	KindVariableDeclaration  = ast.KindVariableDeclaration
	KindImportDeclaration    = ast.KindImportDeclaration
	KindExportDeclaration    = ast.KindExportDeclaration
	KindExportAssignment     = ast.KindExportAssignment
	KindIdentifier           = ast.KindIdentifier
)

// Node wraps a TypeScript AST node. It is a thin, comparable value type;
// the zero value is not usable.
//
// A Node is only valid while its source file is unmodified: after any edit
// to the file the node is forgotten (IsForgotten returns true) and calling
// any of its methods panics, matching ts-morph's behaviour. Re-fetch nodes
// from the SourceFile after edits.
type Node struct {
	node *ast.Node
	sf   *SourceFile
	gen  int
}

// wrap creates a Node for astNode, recording the file's current generation.
// Returns (Node, false) when astNode is nil.
func (s *SourceFile) wrap(astNode *ast.Node) (Node, bool) {
	if astNode == nil {
		return Node{}, false
	}
	return Node{node: astNode, sf: s, gen: s.generation()}, true
}

// derive creates a child Node from an existing one (same generation).
func (n Node) derive(astNode *ast.Node) (Node, bool) {
	if astNode == nil {
		return Node{}, false
	}
	return Node{node: astNode, sf: n.sf, gen: n.gen}, true
}

// IsForgotten reports whether the node is stale because its source file was
// modified after the node was obtained.
func (n Node) IsForgotten() bool {
	return n.sf == nil || n.gen != n.sf.generation()
}

// check panics if the node is forgotten.
func (n Node) check() {
	if n.IsForgotten() {
		panic("tsmorph: attempt to use a forgotten node (the source file was modified after this node was obtained); re-fetch it from the SourceFile")
	}
}

// ASTNode returns the underlying compiler AST node. Intended for advanced
// use; the compiler packages are internal and may change between releases.
func (n Node) ASTNode() *ast.Node { n.check(); return n.node }

// SourceFile returns the source file containing this node.
func (n Node) SourceFile() *SourceFile { return n.sf }

// Kind returns the syntactic kind of the node.
func (n Node) Kind() Kind { n.check(); return n.node.Kind }

// KindName returns the human-readable kind, e.g. "ClassDeclaration".
func (n Node) KindName() string { return n.node.KindString() }

// Pos returns the byte offset of the start of the node, including any
// leading trivia (whitespace/comments).
func (n Node) Pos() int { n.check(); return n.node.Pos() }

// Start returns the byte offset of the first token of the node, skipping
// leading trivia.
func (n Node) Start() int {
	n.check()
	return scanner.SkipTrivia(n.sf.Text(), n.node.Pos())
}

// End returns the byte offset of the end of the node.
func (n Node) End() int { n.check(); return n.node.End() }

// Text returns the source text of the node, excluding surrounding trivia.
func (n Node) Text() string {
	n.check()
	text := n.sf.Text()
	start, end := n.Start(), n.node.End()
	if start < 0 || end > len(text) || start > end {
		return ""
	}
	return text[start:end]
}

// StartLineAndColumn returns the 1-based line and byte column of Start.
func (n Node) StartLineAndColumn() (line, column int) {
	return n.sf.LineAndColumn(n.Start())
}

// Parent returns the parent node, or false if this is the root.
func (n Node) Parent() (Node, bool) {
	n.check()
	return n.derive(n.node.Parent)
}

// Children returns the direct children of the node.
func (n Node) Children() []Node {
	n.check()
	var out []Node
	for c := range n.node.IterChildren() {
		out = append(out, Node{node: c, sf: n.sf, gen: n.gen})
	}
	return out
}

// Descendants returns all descendants of the node in depth-first order.
func (n Node) Descendants() []Node {
	n.check()
	var out []Node
	var walk func(an *ast.Node)
	walk = func(an *ast.Node) {
		for c := range an.IterChildren() {
			out = append(out, Node{node: c, sf: n.sf, gen: n.gen})
			walk(c)
		}
	}
	walk(n.node)
	return out
}

// DescendantsOfKind returns all descendants (depth-first) with the given kind.
func (n Node) DescendantsOfKind(kind Kind) []Node {
	n.check()
	var out []Node
	var walk func(an *ast.Node)
	walk = func(an *ast.Node) {
		for c := range an.IterChildren() {
			if c.Kind == kind {
				out = append(out, Node{node: c, sf: n.sf, gen: n.gen})
			}
			walk(c)
		}
	}
	walk(n.node)
	return out
}

// FirstDescendantByKind returns the first descendant (depth-first) with the
// given kind, or false.
func (n Node) FirstDescendantByKind(kind Kind) (Node, bool) {
	var found *ast.Node
	var walk func(an *ast.Node) bool
	walk = func(an *ast.Node) bool {
		for c := range an.IterChildren() {
			if c.Kind == kind {
				found = c
				return true
			}
			if walk(c) {
				return true
			}
		}
		return false
	}
	n.check()
	if walk(n.node) {
		return Node{node: found, sf: n.sf, gen: n.gen}, true
	}
	return Node{}, false
}

// FirstAncestorByKind returns the closest ancestor with the given kind,
// or false.
func (n Node) FirstAncestorByKind(kind Kind) (Node, bool) {
	n.check()
	if a := ast.FindAncestorKind(n.node, kind); a != nil {
		return Node{node: a, sf: n.sf, gen: n.gen}, true
	}
	return Node{}, false
}

// nameNode returns the node's name node if it has one.
func (n Node) nameNode() (Node, bool) {
	name := ast.GetNameOfDeclaration(n.node)
	if name == nil {
		return Node{}, false
	}
	return Node{node: name, sf: n.sf, gen: n.gen}, true
}

// Name returns the text of the node's name, or "" if it has none.
func (n Node) Name() string {
	if name, ok := n.nameNode(); ok {
		return name.Text()
	}
	return ""
}

// modifiers reports whether the node has the given syntactic modifiers.
func (n Node) hasModifiers(flags ast.ModifierFlags) bool {
	return ast.HasSyntacticModifier(n.node, flags)
}

// IsExported reports whether the node has an `export` modifier (or is a
// default export).
func (n Node) IsExported() bool {
	return n.hasModifiers(ast.ModifierFlagsExport)
}

// IsDefaultExport reports whether the node has a `default` modifier.
func (n Node) IsDefaultExport() bool {
	return n.hasModifiers(ast.ModifierFlagsDefault)
}

// --- Kind predicates ---

func (n Node) IsClassDeclaration() bool     { return ast.IsClassDeclaration(n.node) }
func (n Node) IsInterfaceDeclaration() bool { return ast.IsInterfaceDeclaration(n.node) }
func (n Node) IsFunctionDeclaration() bool  { return ast.IsFunctionDeclaration(n.node) }
func (n Node) IsMethodDeclaration() bool    { return ast.IsMethodDeclaration(n.node) }
func (n Node) IsPropertyDeclaration() bool  { return ast.IsPropertyDeclaration(n.node) }
func (n Node) IsEnumDeclaration() bool      { return ast.IsEnumDeclaration(n.node) }
func (n Node) IsTypeAliasDeclaration() bool { return ast.IsTypeAliasDeclaration(n.node) }
func (n Node) IsVariableStatement() bool    { return ast.IsVariableStatement(n.node) }
func (n Node) IsImportDeclaration() bool    { return ast.IsImportDeclaration(n.node) }
func (n Node) IsExportDeclaration() bool    { return ast.IsExportDeclaration(n.node) }

// --- Downcasts ---

// AsClassDeclaration downcasts the node, or returns false.
func (n Node) AsClassDeclaration() (ClassDeclaration, bool) {
	if !n.IsClassDeclaration() {
		return ClassDeclaration{}, false
	}
	return ClassDeclaration{Node: n}, true
}

// AsInterfaceDeclaration downcasts the node, or returns false.
func (n Node) AsInterfaceDeclaration() (InterfaceDeclaration, bool) {
	if !n.IsInterfaceDeclaration() {
		return InterfaceDeclaration{}, false
	}
	return InterfaceDeclaration{Node: n}, true
}

// AsFunctionDeclaration downcasts the node, or returns false.
func (n Node) AsFunctionDeclaration() (FunctionDeclaration, bool) {
	if !n.IsFunctionDeclaration() {
		return FunctionDeclaration{}, false
	}
	return FunctionDeclaration{Node: n}, true
}

// AsMethodDeclaration downcasts the node, or returns false.
func (n Node) AsMethodDeclaration() (MethodDeclaration, bool) {
	if !n.IsMethodDeclaration() {
		return MethodDeclaration{}, false
	}
	return MethodDeclaration{Node: n}, true
}

// AsPropertyDeclaration downcasts the node, or returns false.
func (n Node) AsPropertyDeclaration() (PropertyDeclaration, bool) {
	if !n.IsPropertyDeclaration() {
		return PropertyDeclaration{}, false
	}
	return PropertyDeclaration{Node: n}, true
}

// AsEnumDeclaration downcasts the node, or returns false.
func (n Node) AsEnumDeclaration() (EnumDeclaration, bool) {
	if !n.IsEnumDeclaration() {
		return EnumDeclaration{}, false
	}
	return EnumDeclaration{Node: n}, true
}

// AsTypeAliasDeclaration downcasts the node, or returns false.
func (n Node) AsTypeAliasDeclaration() (TypeAliasDeclaration, bool) {
	if !n.IsTypeAliasDeclaration() {
		return TypeAliasDeclaration{}, false
	}
	return TypeAliasDeclaration{Node: n}, true
}

// AsVariableStatement downcasts the node, or returns false.
func (n Node) AsVariableStatement() (VariableStatement, bool) {
	if !n.IsVariableStatement() {
		return VariableStatement{}, false
	}
	return VariableStatement{Node: n}, true
}

// AsImportDeclaration downcasts the node, or returns false.
func (n Node) AsImportDeclaration() (ImportDeclaration, bool) {
	if !n.IsImportDeclaration() {
		return ImportDeclaration{}, false
	}
	return ImportDeclaration{Node: n}, true
}

// AsConstructorDeclaration downcasts the node, or returns false.
func (n Node) AsConstructorDeclaration() (ConstructorDeclaration, bool) {
	if n.Kind() != ast.KindConstructor {
		return ConstructorDeclaration{}, false
	}
	return ConstructorDeclaration{Node: n}, true
}

// AsExportDeclaration downcasts the node, or returns false.
func (n Node) AsExportDeclaration() (ExportDeclaration, bool) {
	if !n.IsExportDeclaration() {
		return ExportDeclaration{}, false
	}
	return ExportDeclaration{Node: n}, true
}

// lineOf returns the 0-based line of a byte offset within the file.
func (n Node) lineOf(pos int) int {
	return scanner.GetECMALineOfPosition(n.sf.astFile(), pos)
}
