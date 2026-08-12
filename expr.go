package tsmorph

// Wrappers for expression, statement, and JSX nodes (beyond the declaration
// subset in declarations.go). These are thin value types embedding Node and
// adding accessors for the concrete syntax they represent.

// CallExpression wraps a call expression `f(args)`.
type CallExpression struct{ Node }

// Expression returns the callee (`f`), or false.
func (c CallExpression) Expression() (Node, bool) { return c.GetExpression() }

// Arguments returns the call's arguments.
func (c CallExpression) Arguments() []Node { return c.GetArguments() }

// PropertyAccessExpression wraps a property access `a.b`.
type PropertyAccessExpression struct{ Node }

// Expression returns the object (`a` in `a.b`), or false.
func (p PropertyAccessExpression) Expression() (Node, bool) { return p.GetExpression() }

// Identifier wraps an identifier.
type Identifier struct{ Node }

// StringLiteral wraps a string literal.
type StringLiteral struct{ Node }

// Value returns the literal's text without quotes.
func (l StringLiteral) Value() string { return l.LiteralValue() }

// ArrowFunction wraps an arrow function expression.
type ArrowFunction struct{ Node }

// Block wraps a statement block `{ ... }`.
type Block struct{ Node }

// Statements returns the block's statements.
func (b Block) Statements() []Node { return b.GetStatements() }

// InsertStatements inserts one or more statements at the given index within
// the block (0 <= index <= number of statements), rendering indentation from
// the project's ManipulationSettings. Existing wrappers for the file are
// forgotten after the edit.
func (b Block) InsertStatements(index int, statements ...string) {
	b.InsertStatementsInto(index, statements...)
}

// JsxAttribute wraps a JSX attribute (`foo="bar"` or `foo={expr}`).
type JsxAttribute struct{ Node }

// NameNode returns the attribute's name node.
func (a JsxAttribute) NameNode() (Node, bool) { return a.GetNameNode() }

// Initializer returns the attribute's value expression, or false.
func (a JsxAttribute) Initializer() (Node, bool) { return a.GetInitializer() }

// JsxElement wraps a JSX element with opening/closing tags.
type JsxElement struct{ Node }

// Children returns the element's JSX children (text, expressions, elements).
func (e JsxElement) Children() []Node { return e.GetJsxChildren() }

// JsxText wraps JSX text content.
type JsxText struct{ Node }

// TextContent returns the raw JSX text (the compiler's Node.Text() panics
// for this kind, so it has a dedicated accessor).
func (t JsxText) TextContent() string {
	t.check()
	return t.node.AsJsxText().Text
}
