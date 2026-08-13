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
	KindClassDeclaration              = ast.KindClassDeclaration
	KindInterfaceDeclaration          = ast.KindInterfaceDeclaration
	KindFunctionDeclaration           = ast.KindFunctionDeclaration
	KindMethodDeclaration             = ast.KindMethodDeclaration
	KindPropertyDeclaration           = ast.KindPropertyDeclaration
	KindConstructor                   = ast.KindConstructor
	KindParameter                     = ast.KindParameter
	KindEnumDeclaration               = ast.KindEnumDeclaration
	KindEnumMember                    = ast.KindEnumMember
	KindTypeAliasDeclaration          = ast.KindTypeAliasDeclaration
	KindVariableStatement             = ast.KindVariableStatement
	KindVariableDeclaration           = ast.KindVariableDeclaration
	KindImportDeclaration             = ast.KindImportDeclaration
	KindExportDeclaration             = ast.KindExportDeclaration
	KindExportAssignment              = ast.KindExportAssignment
	KindIdentifier                    = ast.KindIdentifier
	KindStringLiteral                 = ast.KindStringLiteral
	KindNoSubstitutionTemplateLiteral = ast.KindNoSubstitutionTemplateLiteral
	KindCallExpression                = ast.KindCallExpression
	KindPropertyAccessExpression      = ast.KindPropertyAccessExpression
	KindArrowFunction                 = ast.KindArrowFunction
	KindFunctionExpression            = ast.KindFunctionExpression
	KindBlock                         = ast.KindBlock
	KindPropertySignature             = ast.KindPropertySignature
	KindObjectBindingPattern          = ast.KindObjectBindingPattern
	KindJsxText                       = ast.KindJsxText
	KindJsxAttribute                  = ast.KindJsxAttribute
	KindJsxElement                    = ast.KindJsxElement
	KindJsxSelfClosingElement         = ast.KindJsxSelfClosingElement
	KindJsxFragment                   = ast.KindJsxFragment
	KindJsxExpression                 = ast.KindJsxExpression
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

// IsZero reports whether the node is the zero value (no underlying AST node).
func (n Node) IsZero() bool { return n.node == nil }

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

// HasModifier reports whether the node has the given syntactic modifier
// flags (any of them set). Use the ModifierFlags* constants in the vendored
// ast package (e.g. ast.ModifierFlagsPrivate).
func (n Node) HasModifier(flags ast.ModifierFlags) bool {
	n.check()
	return n.hasModifiers(flags)
}

// HasModifierPrivate reports whether the node has a `private` modifier.
func (n Node) HasModifierPrivate() bool { return n.hasModifiers(ast.ModifierFlagsPrivate) }

// HasModifierProtected reports whether the node has a `protected` modifier.
func (n Node) HasModifierProtected() bool { return n.hasModifiers(ast.ModifierFlagsProtected) }

// HasModifierStatic reports whether the node has a `static` modifier.
func (n Node) HasModifierStatic() bool { return n.hasModifiers(ast.ModifierFlagsStatic) }

// IsAbstract reports whether the node has an `abstract` modifier.
func (n Node) IsAbstract() bool { return n.hasModifiers(ast.ModifierFlagsAbstract) }

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

// --- Generic expression / statement / JSX accessors ---
//
// These narrow the constructor to arbitrary expression and statement kinds
// (beyond the declaration subset wrapped above). They are thin wrappers
// around the compiler's accessors; the returned Node is a sibling wrapper in
// the same generation.

// IsCallExpression reports whether the node is a call expression.
func (n Node) IsCallExpression() bool { return ast.IsCallExpression(n.node) }

// IsPropertyAccessExpression reports whether the node is a property access
// (`a.b`).
func (n Node) IsPropertyAccessExpression() bool {
	return n.node.Kind == ast.KindPropertyAccessExpression
}

// IsIdentifier reports whether the node is an identifier.
func (n Node) IsIdentifier() bool { return n.node.Kind == ast.KindIdentifier }

// IsStringLiteral reports whether the node is a string literal.
func (n Node) IsStringLiteral() bool { return n.node.Kind == ast.KindStringLiteral }

// IsArrowFunction reports whether the node is an arrow function.
func (n Node) IsArrowFunction() bool { return n.node.Kind == ast.KindArrowFunction }

// IsBlock reports whether the node is a statement block (`{ ... }`).
func (n Node) IsBlock() bool { return n.node.Kind == ast.KindBlock }

// IsFunctionLike reports whether the node is a function-like declaration or
// expression (function, method, constructor, arrow, etc.).
func (n Node) IsFunctionLike() bool { return ast.IsFunctionLike(n.node) }

// IsVariableDeclaration reports whether the node is a variable declaration.
func (n Node) IsVariableDeclaration() bool { return n.node.Kind == ast.KindVariableDeclaration }

// IsObjectBindingPattern reports whether the node is an object binding
// pattern (`{ a, b }` in a destructuring parameter).
func (n Node) IsObjectBindingPattern() bool {
	return n.node.Kind == ast.KindObjectBindingPattern
}

// IsJsxText reports whether the node is JSX text content.
func (n Node) IsJsxText() bool { return n.node.Kind == ast.KindJsxText }

// IsJsxAttribute reports whether the node is a JSX attribute.
func (n Node) IsJsxAttribute() bool { return n.node.Kind == ast.KindJsxAttribute }

// IsJsxElement reports whether the node is a JSX element (opening+closing).
func (n Node) IsJsxElement() bool { return n.node.Kind == ast.KindJsxElement }

// AsCallExpression downcasts the node, or returns false.
func (n Node) AsCallExpression() (CallExpression, bool) {
	n.check()
	if !n.IsCallExpression() {
		return CallExpression{}, false
	}
	return CallExpression{Node: n}, true
}

// AsPropertyAccessExpression downcasts the node, or returns false.
func (n Node) AsPropertyAccessExpression() (PropertyAccessExpression, bool) {
	n.check()
	if !n.IsPropertyAccessExpression() {
		return PropertyAccessExpression{}, false
	}
	return PropertyAccessExpression{Node: n}, true
}

// AsIdentifier downcasts the node, or returns false.
func (n Node) AsIdentifier() (Identifier, bool) {
	n.check()
	if !n.IsIdentifier() {
		return Identifier{}, false
	}
	return Identifier{Node: n}, true
}

// AsStringLiteral downcasts the node, or returns false.
func (n Node) AsStringLiteral() (StringLiteral, bool) {
	n.check()
	if !n.IsStringLiteral() {
		return StringLiteral{}, false
	}
	return StringLiteral{Node: n}, true
}

// AsArrowFunction downcasts the node, or returns false.
func (n Node) AsArrowFunction() (ArrowFunction, bool) {
	n.check()
	if !n.IsArrowFunction() {
		return ArrowFunction{}, false
	}
	return ArrowFunction{Node: n}, true
}

// AsBlock downcasts the node, or returns false.
func (n Node) AsBlock() (Block, bool) {
	n.check()
	if !n.IsBlock() {
		return Block{}, false
	}
	return Block{Node: n}, true
}

// AsVariableDeclaration downcasts the node, or returns false.
func (n Node) AsVariableDeclaration() (VariableDeclaration, bool) {
	n.check()
	if !n.IsVariableDeclaration() {
		return VariableDeclaration{}, false
	}
	return VariableDeclaration{Node: n}, true
}

// AsJsxAttribute downcasts the node, or returns false.
func (n Node) AsJsxAttribute() (JsxAttribute, bool) {
	n.check()
	if !n.IsJsxAttribute() {
		return JsxAttribute{}, false
	}
	return JsxAttribute{Node: n}, true
}

// AsJsxElement downcasts the node, or returns false.
func (n Node) AsJsxElement() (JsxElement, bool) {
	n.check()
	if !n.IsJsxElement() {
		return JsxElement{}, false
	}
	return JsxElement{Node: n}, true
}

// AsJsxText downcasts the node, or returns false.
func (n Node) AsJsxText() (JsxText, bool) {
	n.check()
	if !n.IsJsxText() {
		return JsxText{}, false
	}
	return JsxText{Node: n}, true
}

// LiteralValue returns the unquoted text of a string literal or no-substitution
// template literal, or "" if the node is not such a literal.
func (n Node) LiteralValue() string {
	n.check()
	if n.IsStringLiteral() {
		return n.node.AsStringLiteral().Text
	}
	if n.node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return n.node.Text()
	}
	return ""
}

// GetExpression returns the object expression of a call, property access, or
// JSX expression node, or returns false. For a call `f(a)`, this is `f`.
func (n Node) GetExpression() (Node, bool) {
	n.check()
	return n.child(n.node.Expression)
}

// GetArguments returns the arguments of a call expression, or nil.
func (n Node) GetArguments() []Node {
	n.check()
	return n.children(n.node.Arguments)
}

// GetNameNode returns the name node (for attributes, property access, etc.),
// or false.
func (n Node) GetNameNode() (Node, bool) {
	n.check()
	if name := n.node.Name(); name != nil {
		return n.derive(name)
	}
	return Node{}, false
}

// GetInitializer returns the initializer of the node, or false.
func (n Node) GetInitializer() (Node, bool) {
	n.check()
	if init := n.node.Initializer(); init != nil {
		return n.derive(init)
	}
	return Node{}, false
}

// GetBody returns the body (block or expression) of a function-like node, or
// false.
func (n Node) GetBody() (Node, bool) {
	n.check()
	if body := n.node.Body(); body != nil {
		return n.derive(body)
	}
	return Node{}, false
}

// GetStatements returns the statements of a block or source file, or nil.
func (n Node) GetStatements() []Node {
	n.check()
	return n.children(n.node.Statements)
}

// GetJsxChildren returns the children of a JSX element or fragment, or nil.
func (n Node) GetJsxChildren() []Node {
	n.check()
	var out []Node
	if children := n.node.Children(); children != nil {
		for _, c := range children.Nodes {
			if w, ok := n.derive(c); ok {
				out = append(out, w)
			}
		}
	}
	return out
}

// lineOf returns the 0-based line of a byte offset within the file.
func (n Node) lineOf(pos int) int {
	return scanner.GetECMALineOfPosition(n.sf.astFile(), pos)
}

// wrapNodes wraps a slice of AST child nodes, skipping nil entries.
func (n Node) wrapNodes(nodes []*ast.Node) []Node {
	var out []Node
	for _, c := range nodes {
		if c == nil {
			continue
		}
		out = append(out, Node{node: c, sf: n.sf, gen: n.gen})
	}
	return out
}

// child invokes f, which extracts a child node, and wraps the result. The
// vendored compiler panics when a node kind does not support the requested
// child; that panic is translated into a false result so accessors are safe to
// call on any node.
func (n Node) child(f func() *ast.Node) (Node, bool) {
	defer func() { _ = recover() }()
	if c := f(); c != nil {
		return n.derive(c)
	}
	return Node{}, false
}

// children invokes f, which extracts a child-node list, and wraps the result,
// translating the vendored compiler's kind-mismatch panic into nil.
func (n Node) children(f func() []*ast.Node) []Node {
	defer func() { _ = recover() }()
	return n.wrapNodes(f())
}

// GetType returns the declared type node (for typed declarations, parameters,
// and type-annotated expressions), or false when there is none.
func (n Node) GetType() (Node, bool) {
	n.check()
	return n.child(n.node.Type)
}

// GetTypeArguments returns the type arguments of a call/new expression or type
// reference, or nil.
func (n Node) GetTypeArguments() []Node {
	n.check()
	return n.children(n.node.TypeArguments)
}

// GetTypeParameters returns the type parameters of a declaration, or nil.
func (n Node) GetTypeParameters() []Node {
	n.check()
	return n.children(n.node.TypeParameters)
}

// GetMembers returns the members of a class, interface, enum, or type literal,
// or nil.
func (n Node) GetMembers() []Node {
	n.check()
	return n.children(n.node.Members)
}

// GetProperties returns the properties of an object literal or type literal,
// or nil.
func (n Node) GetProperties() []Node {
	n.check()
	return n.children(n.node.Properties)
}

// GetElements returns the elements of an array literal, or nil.
func (n Node) GetElements() []Node {
	n.check()
	return n.children(n.node.Elements)
}

// GetDecorators returns the decorators of a node, or nil.
func (n Node) GetDecorators() []Node {
	n.check()
	return n.children(n.node.Decorators)
}

// GetComments returns the comment nodes of a JSDoc comment or JSDoc tag, or
// nil.
func (n Node) GetComments() []Node {
	n.check()
	return n.children(n.node.Comments)
}

// GetModifierNodes returns the modifier nodes of a node, or nil.
func (n Node) GetModifierNodes() []Node {
	n.check()
	return n.children(n.node.ModifierNodes)
}

// GetTagName returns the tag name of a JSX element or JSDoc tag, or false.
func (n Node) GetTagName() (Node, bool) {
	n.check()
	return n.child(n.node.TagName)
}

// GetLabel returns the label of a labeled statement, or false.
func (n Node) GetLabel() (Node, bool) {
	n.check()
	return n.child(n.node.Label)
}

// GetAttributes returns the import attributes (or assertions) node, or false.
func (n Node) GetAttributes() (Node, bool) {
	n.check()
	return n.child(n.node.Attributes)
}

// GetStatement returns the single embedded statement of a control-flow node
// (if/for/while/do/with/labeled), or false.
func (n Node) GetStatement() (Node, bool) {
	n.check()
	return n.child(n.node.Statement)
}

// GetPostfixToken returns the postfix operator token (`++`, `--`, `!`), or
// false.
func (n Node) GetPostfixToken() (Node, bool) {
	n.check()
	return n.child(n.node.PostfixToken)
}

// GetTypeExpression returns the type expression of a type assertion, satisfies,
// or typeof node, or false.
func (n Node) GetTypeExpression() (Node, bool) {
	n.check()
	return n.child(n.node.TypeExpression)
}

// HasQuestionToken reports whether the node carries a `?` token (optional
// property/parameter/method, conditional, etc.).
func (n Node) HasQuestionToken() bool {
	n.check()
	return n.node.QuestionToken() != nil
}

// HasQuestionDotToken reports whether the node carries a `?.` optional-chaining
// token.
func (n Node) HasQuestionDotToken() bool {
	n.check()
	_, ok := n.child(n.node.QuestionDotToken)
	return ok
}
