package tsmorph

import (
	"strings"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

// JSDoc support, mirroring ts-morph's JSDoc API. The concrete JSDoc node kinds
// (JSDoc, JSDocText, JSDocReturnTag, ...) are generated in kinds_generated.go;
// this file adds the base JSDocTag type, the ParameterTag/PropertyTag pair
// (which share a single AST node in the compiler), and the accessors.

// GetJsDocs returns the JSDoc comments attached to a declaration (leading
// `/** ... */` comments), or nil.
func (n Node) GetJsDocs() []JSDoc {
	n.check()
	var out []JSDoc
	defer func() { _ = recover() }()
	for _, j := range n.node.JSDoc(n.sf.astFile()) {
		if w, ok := n.derive(j); ok {
			out = append(out, JSDoc{Node: w})
		}
	}
	return out
}

// commentText returns the joined text of a JSDoc comment node (JSDocText,
// JSDocLink, JSDocLinkCode, JSDocLinkPlain). The vendored compiler's Node.Text
// handles these kinds specially.
func commentText(n Node) string {
	defer func() { _ = recover() }()
	return n.node.Text()
}

// joinCommentText joins the text of the given comment nodes.
func joinCommentText(nodes []Node) string {
	var sb strings.Builder
	for _, c := range nodes {
		sb.WriteString(commentText(c))
	}
	return strings.TrimSpace(sb.String())
}

// JSDoc accessors.

// GetTags returns the JSDoc tags (`@param`, `@returns`, ...) in the comment.
func (j JSDoc) GetTags() []JSDocTag {
	var out []JSDocTag
	tags := j.node.AsJSDoc().Tags
	if tags == nil {
		return nil
	}
	for _, t := range tags.Nodes {
		if ast.IsJSDocTag(t) {
			out = append(out, JSDocTag{Node{node: t, sf: j.sf, gen: j.gen}})
		}
	}
	return out
}

// GetComment returns the description text preceding the first tag.
func (j JSDoc) GetComment() string {
	return joinCommentText(j.GetComments())
}

// GetInnerText returns the comment text without the surrounding `/**` and
// `*/` markers.
func (j JSDoc) GetInnerText() string {
	text := j.sourceText(j.Pos(), j.End())
	text = strings.TrimPrefix(text, "/**")
	text = strings.TrimSuffix(text, "*/")
	return text
}

// IsMultiLine reports whether the comment spans multiple lines.
func (j JSDoc) IsMultiLine() bool {
	return strings.Contains(j.sourceText(j.Pos(), j.End()), "\n")
}

// JSDocTag is the base wrapper for JSDoc tag nodes. Concrete tags (JSDocReturnTag,
// JSDocParameterTag, ...) can be obtained via their As* downcasts.
type JSDocTag struct{ Node }

// IsJSDocTag reports whether the node is a JSDoc tag.
func (n Node) IsJSDocTag() bool { return ast.IsJSDocTag(n.node) }

// AsJSDocTag downcasts the node to a JSDoc tag, or returns false.
func (n Node) AsJSDocTag() (JSDocTag, bool) {
	n.check()
	if !n.IsJSDocTag() {
		return JSDocTag{}, false
	}
	return JSDocTag{Node: n}, true
}

// TagName returns the tag name without the leading `@` (e.g. "param").
func (t JSDocTag) TagName() string {
	if name, ok := t.TagNameNode(); ok {
		return name.Text()
	}
	return ""
}

// TagNameNode returns the tag name node (the identifier after `@`).
func (t JSDocTag) TagNameNode() (Node, bool) { return t.Node.GetTagName() }

// GetComment returns the text following the tag.
func (t JSDocTag) GetComment() string { return joinCommentText(t.GetComments()) }

// JSDocParameterTag wraps a `@param` tag.
type JSDocParameterTag struct{ JSDocTag }

// IsJSDocParameterTag reports whether the node is a `@param` tag.
func (n Node) IsJSDocParameterTag() bool { return n.node.Kind == ast.KindJSDocParameterTag }

// AsJSDocParameterTag downcasts the node, or returns false.
func (n Node) AsJSDocParameterTag() (JSDocParameterTag, bool) {
	n.check()
	if !n.IsJSDocParameterTag() {
		return JSDocParameterTag{}, false
	}
	return JSDocParameterTag{JSDocTag: JSDocTag{Node: n}}, true
}

// JSDocPropertyTag wraps a `@property` (or `@prop`) tag.
type JSDocPropertyTag struct{ JSDocTag }

// IsJSDocPropertyTag reports whether the node is a `@property` tag.
func (n Node) IsJSDocPropertyTag() bool { return n.node.Kind == ast.KindJSDocPropertyTag }

// AsJSDocPropertyTag downcasts the node, or returns false.
func (n Node) AsJSDocPropertyTag() (JSDocPropertyTag, bool) {
	n.check()
	if !n.IsJSDocPropertyTag() {
		return JSDocPropertyTag{}, false
	}
	return JSDocPropertyTag{JSDocTag: JSDocTag{Node: n}}, true
}

// parameterOrPropertyData returns the shared compiler node for `@param` and
// `@property` tags.
func (n Node) parameterOrPropertyData() *ast.JSDocParameterOrPropertyTag {
	return n.node.AsJSDocParameterOrPropertyTag()
}

// Name returns the parameter/property name, or false.
func (t JSDocParameterTag) Name() (Node, bool) { return t.GetNameNode() }

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocParameterTag) TypeExpression() (Node, bool) {
	if te := t.parameterOrPropertyData().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// IsBracketed reports whether the name is in square brackets (`[x]`).
func (t JSDocParameterTag) IsBracketed() bool { return t.parameterOrPropertyData().IsBracketed }

// Name returns the property name, or false.
func (t JSDocPropertyTag) Name() (Node, bool) { return t.GetNameNode() }

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocPropertyTag) TypeExpression() (Node, bool) {
	if te := t.parameterOrPropertyData().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// IsBracketed reports whether the name is in square brackets (`[x]`).
func (t JSDocPropertyTag) IsBracketed() bool { return t.parameterOrPropertyData().IsBracketed }

// Type-tag accessors.

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocReturnTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocReturnTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocTypeTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocTypeTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocThisTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocThisTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocThrowsTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocThrowsTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocOverloadTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocOverloadTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the `{Type}` expression, or false.
func (t JSDocSatisfiesTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocSatisfiesTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// TypeExpression returns the type expression, or false.
func (t JSDocTypedefTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocTypedefTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// Name returns the typedef name, or false.
func (t JSDocTypedefTag) Name() (Node, bool) { return t.GetNameNode() }

// TypeExpression returns the callback's type expression, or false.
func (t JSDocCallbackTag) TypeExpression() (Node, bool) {
	if te := t.node.AsJSDocCallbackTag().TypeExpression; te != nil {
		return t.derive(te)
	}
	return Node{}, false
}

// Name returns the callback name, or false.
func (t JSDocCallbackTag) Name() (Node, bool) { return t.GetNameNode() }

// ClassName returns the referenced class/interface, or false.
func (t JSDocAugmentsTag) ClassName() (ExpressionWithTypeArguments, bool) {
	cn := t.node.AsJSDocAugmentsTag().ClassName
	if cn == nil {
		return ExpressionWithTypeArguments{}, false
	}
	n, _ := t.derive(cn)
	return ExpressionWithTypeArguments{Node: n}, true
}

// ClassName returns the referenced class/interface, or false.
func (t JSDocImplementsTag) ClassName() (ExpressionWithTypeArguments, bool) {
	cn := t.node.AsJSDocImplementsTag().ClassName
	if cn == nil {
		return ExpressionWithTypeArguments{}, false
	}
	n, _ := t.derive(cn)
	return ExpressionWithTypeArguments{Node: n}, true
}

// NameExpression returns the referenced name, or false.
func (t JSDocSeeTag) NameExpression() (Node, bool) {
	if ne := t.node.AsJSDocSeeTag().NameExpression; ne != nil {
		return t.derive(ne)
	}
	return Node{}, false
}

// TypeParameters returns the template type parameters, or nil.
func (t JSDocTemplateTag) TypeParameters() []Node {
	list := t.node.AsJSDocTemplateTag().TypeParameters
	if list == nil {
		return nil
	}
	return t.wrapNodes(list.Nodes)
}

// Constraint returns the constraint expression, or false.
func (t JSDocTemplateTag) Constraint() (Node, bool) {
	if c := t.node.AsJSDocTemplateTag().Constraint; c != nil {
		return t.derive(c)
	}
	return Node{}, false
}

// ModuleSpecifier returns the import tag's module specifier, or false.
func (t JSDocImportTag) ModuleSpecifier() (Node, bool) {
	if ms := t.node.AsJSDocImportTag().ModuleSpecifier; ms != nil {
		return t.derive(ms)
	}
	return Node{}, false
}

// ImportClause returns the import tag's import clause, or false.
func (t JSDocImportTag) ImportClause() (ImportClause, bool) {
	ic := t.node.AsJSDocImportTag().ImportClause
	if ic == nil {
		return ImportClause{}, false
	}
	n, _ := t.derive(ic)
	return ImportClause{Node: n}, true
}

// --- JSDoc type nodes ---

// Type returns the wrapped type of a JSDoc type node, or false.
func (t JSDocTypeExpression) Type() (Node, bool) { return t.GetType() }

// Type returns the wrapped type, or false.
func (t JSDocNonNullableType) Type() (Node, bool) { return t.GetType() }

// Type returns the wrapped type, or false.
func (t JSDocNullableType) Type() (Node, bool) { return t.GetType() }

// Type returns the wrapped type, or false.
func (t JSDocOptionalType) Type() (Node, bool) { return t.GetType() }

// Type returns the variadic type's wrapped type, or false.
func (t JSDocVariadicType) Type() (Node, bool) {
	if typ := t.node.AsJSDocVariadicType().Type; typ != nil {
		return t.derive(typ)
	}
	return Node{}, false
}

// Name returns the referenced name, or false.
func (t JSDocNameReference) Name() (Node, bool) { return t.GetNameNode() }

// GetPropertyTags returns the `@property` tags of a JSDoc type literal (the
// `{Object}` of a `@typedef`), or nil.
func (t JSDocTypeLiteral) GetPropertyTags() []JSDocPropertyTag {
	var out []JSDocPropertyTag
	for _, p := range t.node.AsJSDocTypeLiteral().JSDocPropertyTags {
		if w, ok := t.derive(p); ok {
			out = append(out, JSDocPropertyTag{JSDocTag: JSDocTag{Node: w}})
		}
	}
	return out
}
