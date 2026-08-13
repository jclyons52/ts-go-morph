package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

// Typed accessors for the most commonly used statement, expression, type-node,
// JSX, and module-specifier kinds. These complement the generic accessors on
// Node (GetExpression, GetStatement, GetType, GetArguments, ...) with
// kind-specific child access where a node has more than one.

// sourceText returns the source text for the half-open range [start, end).
func (n Node) sourceText(start, end int) string {
	text := n.sf.Text()
	if start < 0 || end > len(text) || start > end {
		return ""
	}
	return text[start:end]
}

// operatorText returns the source text of the node's first operator token
// child (the vendored compiler does not expose operator tokens as a generic
// accessor). It returns "" when the node has no operator token.
func (n Node) operatorText() string {
	for c := range n.node.IterChildren() {
		if ast.IsToken(c) {
			return n.sourceText(c.Pos(), c.End())
		}
	}
	return ""
}

// --- Statements ---

// IfStatement accessors.

// ThenStatement returns the `then` branch of an if statement.
func (i IfStatement) ThenStatement() (Node, bool) {
	st := i.node.AsIfStatement().ThenStatement
	if st == nil {
		return Node{}, false
	}
	return i.derive(st)
}

// ElseStatement returns the `else` branch of an if statement, or false.
func (i IfStatement) ElseStatement() (Node, bool) {
	st := i.node.AsIfStatement().ElseStatement
	if st == nil {
		return Node{}, false
	}
	return i.derive(st)
}

// ForStatement accessors.

// Initializer returns the for-loop initializer, or false.
func (f ForStatement) Initializer() (Node, bool) {
	init := f.node.AsForStatement().Initializer
	if init == nil {
		return Node{}, false
	}
	return f.derive(init)
}

// Condition returns the for-loop condition, or false.
func (f ForStatement) Condition() (Node, bool) {
	cond := f.node.AsForStatement().Condition
	if cond == nil {
		return Node{}, false
	}
	return f.derive(cond)
}

// Incrementor returns the for-loop incrementor, or false.
func (f ForStatement) Incrementor() (Node, bool) {
	inc := f.node.AsForStatement().Incrementor
	if inc == nil {
		return Node{}, false
	}
	return f.derive(inc)
}

// ForInOrOfStatement accessors.

// Initializer returns the loop variable of a for-in/for-of statement, or false.
func (f ForInOrOfStatement) Initializer() (Node, bool) {
	init := f.node.AsForInOrOfStatement().Initializer
	if init == nil {
		return Node{}, false
	}
	return f.derive(init)
}

// IsForIn reports whether the loop is a `for ... in` statement.
func (f ForInOrOfStatement) IsForIn() bool { return f.Kind() == ast.KindForInStatement }

// IsForOf reports whether the loop is a `for ... of` statement.
func (f ForInOrOfStatement) IsForOf() bool { return f.Kind() == ast.KindForOfStatement }

// TryStatement accessors.

// TryBlock returns the try block.
func (t TryStatement) TryBlock() (Block, bool) {
	block := t.node.AsTryStatement().TryBlock
	if block == nil {
		return Block{}, false
	}
	n, _ := t.derive(block)
	return Block{Node: n}, true
}

// CatchClause returns the catch clause, or false.
func (t TryStatement) CatchClause() (CatchClause, bool) {
	clause := t.node.AsTryStatement().CatchClause
	if clause == nil {
		return CatchClause{}, false
	}
	n, _ := t.derive(clause)
	return CatchClause{Node: n}, true
}

// FinallyBlock returns the finally block, or false.
func (t TryStatement) FinallyBlock() (Block, bool) {
	block := t.node.AsTryStatement().FinallyBlock
	if block == nil {
		return Block{}, false
	}
	n, _ := t.derive(block)
	return Block{Node: n}, true
}

// CatchClause accessors.

// VariableDeclaration returns the catch binding, or false.
func (c CatchClause) VariableDeclaration() (VariableDeclaration, bool) {
	decl := c.node.AsCatchClause().VariableDeclaration
	if decl == nil {
		return VariableDeclaration{}, false
	}
	n, _ := c.derive(decl)
	return VariableDeclaration{Node: n}, true
}

// Block returns the catch block.
func (c CatchClause) Block() (Block, bool) {
	block := c.node.AsCatchClause().Block
	if block == nil {
		return Block{}, false
	}
	n, _ := c.derive(block)
	return Block{Node: n}, true
}

// SwitchStatement accessors.

// CaseBlock returns the switch's case block.
func (s SwitchStatement) CaseBlock() (CaseBlock, bool) {
	cb := s.node.AsSwitchStatement().CaseBlock
	if cb == nil {
		return CaseBlock{}, false
	}
	n, _ := s.derive(cb)
	return CaseBlock{Node: n}, true
}

// CaseBlock accessors.

// Clauses returns the case/default clauses of a switch.
func (c CaseBlock) Clauses() []Node {
	list := c.node.AsCaseBlock().Clauses
	if list == nil {
		return nil
	}
	return c.wrapNodes(list.Nodes)
}

// --- Expressions ---

// BinaryExpression accessors.

// Left returns the left operand.
func (b BinaryExpression) Left() (Node, bool) {
	if left := b.node.AsBinaryExpression().Left; left != nil {
		return b.derive(left)
	}
	return Node{}, false
}

// Right returns the right operand.
func (b BinaryExpression) Right() (Node, bool) {
	if right := b.node.AsBinaryExpression().Right; right != nil {
		return b.derive(right)
	}
	return Node{}, false
}

// OperatorText returns the operator text, e.g. "+" or "===".
func (b BinaryExpression) OperatorText() string { return b.operatorText() }

// ConditionalExpression accessors.

// Condition returns the condition expression.
func (c ConditionalExpression) Condition() (Node, bool) {
	if cond := c.node.AsConditionalExpression().Condition; cond != nil {
		return c.derive(cond)
	}
	return Node{}, false
}

// WhenTrue returns the branch evaluated when the condition is true.
func (c ConditionalExpression) WhenTrue() (Node, bool) {
	if t := c.node.AsConditionalExpression().WhenTrue; t != nil {
		return c.derive(t)
	}
	return Node{}, false
}

// WhenFalse returns the branch evaluated when the condition is false.
func (c ConditionalExpression) WhenFalse() (Node, bool) {
	if f := c.node.AsConditionalExpression().WhenFalse; f != nil {
		return c.derive(f)
	}
	return Node{}, false
}

// PrefixUnaryExpression accessors.

// Operand returns the operand of a prefix unary expression (`!x`, `-x`).
func (u PrefixUnaryExpression) Operand() (Node, bool) {
	if op := u.node.AsPrefixUnaryExpression().Operand; op != nil {
		return u.derive(op)
	}
	return Node{}, false
}

// OperatorText returns the operator text, e.g. "!" or "-".
func (u PrefixUnaryExpression) OperatorText() string { return u.operatorText() }

// PostfixUnaryExpression accessors.

// Operand returns the operand of a postfix unary expression (`x++`, `x--`).
func (u PostfixUnaryExpression) Operand() (Node, bool) {
	if op := u.node.AsPostfixUnaryExpression().Operand; op != nil {
		return u.derive(op)
	}
	return Node{}, false
}

// OperatorText returns the operator text, e.g. "++" or "--".
func (u PostfixUnaryExpression) OperatorText() string { return u.operatorText() }

// TaggedTemplateExpression accessors.

// Tag returns the tag expression of a tagged template, or false.
func (t TaggedTemplateExpression) Tag() (Node, bool) {
	if tag := t.node.AsTaggedTemplateExpression().Tag; tag != nil {
		return t.derive(tag)
	}
	return Node{}, false
}

// Template returns the template literal being tagged.
func (t TaggedTemplateExpression) Template() (Node, bool) {
	if tmpl := t.node.AsTaggedTemplateExpression().Template; tmpl != nil {
		return t.derive(tmpl)
	}
	return Node{}, false
}

// TemplateExpression accessors.

// Head returns the template's head literal.
func (t TemplateExpression) Head() (Node, bool) {
	if head := t.node.AsTemplateExpression().Head; head != nil {
		return t.derive(head)
	}
	return Node{}, false
}

// TemplateSpans returns the template's spans (`...${expr}`).
func (t TemplateExpression) TemplateSpans() []Node {
	list := t.node.AsTemplateExpression().TemplateSpans
	if list == nil {
		return nil
	}
	return t.wrapNodes(list.Nodes)
}

// TemplateSpan accessors.

// Literal returns the template literal chunk following the expression.
func (t TemplateSpan) Literal() (Node, bool) {
	if lit := t.node.AsTemplateSpan().Literal; lit != nil {
		return t.derive(lit)
	}
	return Node{}, false
}

// ElementAccessExpression accessors.

// ArgumentExpression returns the index expression (`obj[key]` -> key).
func (e ElementAccessExpression) ArgumentExpression() (Node, bool) {
	if arg := e.node.AsElementAccessExpression().ArgumentExpression; arg != nil {
		return e.derive(arg)
	}
	return Node{}, false
}

// --- Type nodes ---

// TypeReferenceNode accessors.

// TypeName returns the referenced type name.
func (t TypeReferenceNode) TypeName() (Node, bool) {
	if name := t.node.AsTypeReferenceNode().TypeName; name != nil {
		return t.derive(name)
	}
	return Node{}, false
}

// UnionTypeNode accessors.

// Types returns the union member types.
func (u UnionTypeNode) Types() []Node {
	return u.wrapNodes(u.node.AsUnionTypeNode().Types.Nodes)
}

// IntersectionTypeNode accessors.

// Types returns the intersection member types.
func (i IntersectionTypeNode) Types() []Node {
	return i.wrapNodes(i.node.AsIntersectionTypeNode().Types.Nodes)
}

// ArrayTypeNode accessors.

// ElementType returns the array element type.
func (a ArrayTypeNode) ElementType() (Node, bool) {
	if elem := a.node.AsArrayTypeNode().ElementType; elem != nil {
		return a.derive(elem)
	}
	return Node{}, false
}

// LiteralTypeNode accessors.

// Literal returns the literal of a literal type node (`"a"`, `1`, `true`).
func (l LiteralTypeNode) Literal() (Node, bool) {
	if lit := l.node.AsLiteralTypeNode().Literal; lit != nil {
		return l.derive(lit)
	}
	return Node{}, false
}

// IndexedAccessTypeNode accessors.

// ObjectType returns the object type being indexed.
func (i IndexedAccessTypeNode) ObjectType() (Node, bool) {
	if obj := i.node.AsIndexedAccessTypeNode().ObjectType; obj != nil {
		return i.derive(obj)
	}
	return Node{}, false
}

// IndexType returns the index type.
func (i IndexedAccessTypeNode) IndexType() (Node, bool) {
	if idx := i.node.AsIndexedAccessTypeNode().IndexType; idx != nil {
		return i.derive(idx)
	}
	return Node{}, false
}

// ConditionalTypeNode accessors.

// CheckType returns the checked type of a conditional type.
func (c ConditionalTypeNode) CheckType() (Node, bool) {
	if t := c.node.AsConditionalTypeNode().CheckType; t != nil {
		return c.derive(t)
	}
	return Node{}, false
}

// ExtendsType returns the extends type of a conditional type.
func (c ConditionalTypeNode) ExtendsType() (Node, bool) {
	if t := c.node.AsConditionalTypeNode().ExtendsType; t != nil {
		return c.derive(t)
	}
	return Node{}, false
}

// TrueType returns the true branch of a conditional type.
func (c ConditionalTypeNode) TrueType() (Node, bool) {
	if t := c.node.AsConditionalTypeNode().TrueType; t != nil {
		return c.derive(t)
	}
	return Node{}, false
}

// FalseType returns the false branch of a conditional type.
func (c ConditionalTypeNode) FalseType() (Node, bool) {
	if t := c.node.AsConditionalTypeNode().FalseType; t != nil {
		return c.derive(t)
	}
	return Node{}, false
}

// TypeOperatorNode accessors.

// OperatorText returns the type operator text, e.g. "keyof" or "readonly".
func (t TypeOperatorNode) OperatorText() string { return t.operatorText() }

// TypePredicateNode accessors.

// ParameterName returns the predicate's parameter name (`x is string` -> x).
func (t TypePredicateNode) ParameterName() (Node, bool) {
	if name := t.node.AsTypePredicateNode().ParameterName; name != nil {
		return t.derive(name)
	}
	return Node{}, false
}

// MappedTypeNode accessors.

// TypeParameter returns the mapped type's type parameter.
func (m MappedTypeNode) TypeParameter() (Node, bool) {
	if tp := m.node.AsMappedTypeNode().TypeParameter; tp != nil {
		return m.derive(tp)
	}
	return Node{}, false
}

// NameType returns the mapped type's `as` name re-mapping type, or false.
func (m MappedTypeNode) NameType() (Node, bool) {
	if nt := m.node.AsMappedTypeNode().NameType; nt != nil {
		return m.derive(nt)
	}
	return Node{}, false
}

// ImportTypeNode accessors.

// Argument returns the import type's argument (the module specifier type).
func (i ImportTypeNode) Argument() (Node, bool) {
	if arg := i.node.AsImportTypeNode().Argument; arg != nil {
		return i.derive(arg)
	}
	return Node{}, false
}

// Qualifier returns the import type's qualifier, or false.
func (i ImportTypeNode) Qualifier() (Node, bool) {
	if q := i.node.AsImportTypeNode().Qualifier; q != nil {
		return i.derive(q)
	}
	return Node{}, false
}

// IsTypeOf reports whether this is an `import type(...)` node.
func (i ImportTypeNode) IsTypeOf() bool { return i.node.AsImportTypeNode().IsTypeOf }

// TypeQueryNode accessors.

// ExprName returns the expression name of a `typeof` type query.
func (t TypeQueryNode) ExprName() (Node, bool) {
	if name := t.node.AsTypeQueryNode().ExprName; name != nil {
		return t.derive(name)
	}
	return Node{}, false
}

// --- JSX ---

// JsxOpeningElement accessors.

// TagName returns the element tag name.
func (e JsxOpeningElement) TagName() (Node, bool) {
	if name := e.node.AsJsxOpeningElement().TagName; name != nil {
		return e.derive(name)
	}
	return Node{}, false
}

// Attributes returns the element's attributes, or false.
func (e JsxOpeningElement) Attributes() (JsxAttributes, bool) {
	attrs := e.node.AsJsxOpeningElement().Attributes
	if attrs == nil {
		return JsxAttributes{}, false
	}
	n, _ := e.derive(attrs)
	return JsxAttributes{Node: n}, true
}

// JsxSelfClosingElement accessors.

// TagName returns the element tag name.
func (e JsxSelfClosingElement) TagName() (Node, bool) {
	if name := e.node.AsJsxSelfClosingElement().TagName; name != nil {
		return e.derive(name)
	}
	return Node{}, false
}

// Attributes returns the element's attributes, or false.
func (e JsxSelfClosingElement) Attributes() (JsxAttributes, bool) {
	attrs := e.node.AsJsxSelfClosingElement().Attributes
	if attrs == nil {
		return JsxAttributes{}, false
	}
	n, _ := e.derive(attrs)
	return JsxAttributes{Node: n}, true
}

// JsxClosingElement accessors.

// TagName returns the closing element's tag name.
func (e JsxClosingElement) TagName() (Node, bool) {
	if name := e.node.AsJsxClosingElement().TagName; name != nil {
		return e.derive(name)
	}
	return Node{}, false
}

// JsxFragment accessors.

// OpeningFragment returns the fragment's opening `<>`, or false.
func (f JsxFragment) OpeningFragment() (Node, bool) {
	if frag := f.node.AsJsxFragment().OpeningFragment; frag != nil {
		return f.derive(frag)
	}
	return Node{}, false
}

// ClosingFragment returns the fragment's closing `</>`, or false.
func (f JsxFragment) ClosingFragment() (Node, bool) {
	if frag := f.node.AsJsxFragment().ClosingFragment; frag != nil {
		return f.derive(frag)
	}
	return Node{}, false
}

// JsxNamespacedName accessors.

// Namespace returns the namespace portion of a namespaced JSX name
// (`ns:name` -> ns).
func (n JsxNamespacedName) Namespace() (Node, bool) {
	ns := n.node.AsJsxNamespacedName().Namespace
	if ns == nil {
		return Node{}, false
	}
	return n.derive(ns)
}

// --- Module specifiers ---

// ImportSpecifier accessors.

// IsTypeOnly reports whether the import specifier is a type-only import.
func (s ImportSpecifier) IsTypeOnly() bool { return s.node.AsImportSpecifier().IsTypeOnly }

// ExportSpecifier accessors.

// IsTypeOnly reports whether the export specifier is a type-only export.
func (s ExportSpecifier) IsTypeOnly() bool { return s.node.AsExportSpecifier().IsTypeOnly }

// ImportClause accessors.

// IsTypeOnly reports whether the import clause is a type-only import.
func (c ImportClause) IsTypeOnly() bool { return c.node.AsImportClause().IsTypeOnly() }

// ExportAssignment accessors.

// IsExportEquals reports whether this is `export =` (rather than
// `export default`).
func (e ExportAssignment) IsExportEquals() bool {
	return e.node.AsExportAssignment().IsExportEquals
}

// ImportEqualsDeclaration accessors.

// ModuleReference returns the referenced module, or false.
func (i ImportEqualsDeclaration) ModuleReference() (Node, bool) {
	if ref := i.node.AsImportEqualsDeclaration().ModuleReference; ref != nil {
		return i.derive(ref)
	}
	return Node{}, false
}

// IsTypeOnly reports whether this is an `import type x = ...` declaration.
func (i ImportEqualsDeclaration) IsTypeOnly() bool {
	return i.node.AsImportEqualsDeclaration().IsTypeOnly
}

// --- Literals ---

// Value returns the numeric literal's text.
func (n NumericLiteral) Value() string { return n.node.AsNumericLiteral().Text }

// Value returns the big-int literal's text (including the trailing `n`).
func (b BigIntLiteral) Value() string { return b.node.AsBigIntLiteral().Text }

// Value returns the regular expression literal's text.
func (r RegularExpressionLiteral) Value() string { return r.node.AsRegularExpressionLiteral().Text }

// Value returns the template literal chunk's text.
func (t TemplateHead) Value() string { return t.node.AsTemplateHead().Text }

// Value returns the template literal chunk's text.
func (t TemplateMiddle) Value() string { return t.node.AsTemplateMiddle().Text }

// Value returns the template literal chunk's text.
func (t TemplateTail) Value() string { return t.node.AsTemplateTail().Text }
