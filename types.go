package tsmorph

import (
	"context"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/checker"
)

// withChecker runs fn with the project's type checker, releasing it after.
// The project is built with a single-threaded checker pool, so all Type and
// Symbol values handed out belong to the same checker instance.
func (p *Project) withChecker(fn func(*checker.Checker)) {
	c, release := p.getProgram().GetTypeChecker(context.Background())
	defer release()
	fn(c)
}

// withCheckerT is withChecker for functions returning a value.
func withCheckerT[T any](p *Project, fn func(*checker.Checker) T) T {
	c, release := p.getProgram().GetTypeChecker(context.Background())
	defer release()
	return fn(c)
}

// withCheckerT2 is withCheckerT for functions returning two values.
func withCheckerT2[A, B any](p *Project, fn func(*checker.Checker) (A, B)) (A, B) {
	c, release := p.getProgram().GetTypeChecker(context.Background())
	defer release()
	return fn(c)
}

// Type wraps a compiler type. Like Node, a Type becomes stale once the
// project is modified; do not use it across edits.
type Type struct {
	t       *checker.Type
	project *Project
}

// Text returns a human-readable rendering of the type, e.g. "string" or
// "{ x: number; y: number; }".
func (t Type) Text() string {
	if t.t == nil {
		return ""
	}
	return withCheckerT(t.project, func(c *checker.Checker) string {
		return c.TypeToString(t.t)
	})
}

func (t Type) hasFlags(flags checker.TypeFlags) bool {
	return t.t != nil && t.t.Flags()&flags != 0
}

// IsAny reports whether the type is `any`.
func (t Type) IsAny() bool { return t.hasFlags(checker.TypeFlagsAny) }

// IsUnknown reports whether the type is `unknown`.
func (t Type) IsUnknown() bool { return t.hasFlags(checker.TypeFlagsUnknown) }

// IsString reports whether the type is `string` (including string literals).
func (t Type) IsString() bool {
	return t.hasFlags(checker.TypeFlagsString | checker.TypeFlagsStringLiteral)
}

// IsNumber reports whether the type is `number` (including number literals).
func (t Type) IsNumber() bool {
	return t.hasFlags(checker.TypeFlagsNumber | checker.TypeFlagsNumberLiteral)
}

// IsBoolean reports whether the type is `boolean` (including true/false).
func (t Type) IsBoolean() bool {
	return t.hasFlags(checker.TypeFlagsBoolean | checker.TypeFlagsBooleanLiteral)
}

// IsNull reports whether the type is `null`.
func (t Type) IsNull() bool { return t.hasFlags(checker.TypeFlagsNull) }

// IsUndefined reports whether the type is `undefined`.
func (t Type) IsUndefined() bool { return t.hasFlags(checker.TypeFlagsUndefined) }

// IsNever reports whether the type is `never`.
func (t Type) IsNever() bool { return t.hasFlags(checker.TypeFlagsNever) }

// IsVoid reports whether the type is `void`.
func (t Type) IsVoid() bool { return t.hasFlags(checker.TypeFlagsVoid) }

// IsUnion reports whether the type is a union type (T | U).
func (t Type) IsUnion() bool { return t.hasFlags(checker.TypeFlagsUnion) }

// IsInterface reports whether the type is an object type declared via an
// interface.
func (t Type) IsInterface() bool {
	if !t.hasFlags(checker.TypeFlagsObject) {
		return false
	}
	sym := t.t.Symbol()
	if sym == nil {
		return false
	}
	for _, d := range sym.Declarations {
		if d.Kind == ast.KindInterfaceDeclaration {
			return true
		}
	}
	return false
}

// IsClass reports whether the type is a class type.
func (t Type) IsClass() bool {
	return t.t != nil && t.t.IsClass()
}

// IsObject reports whether the type is an object type (interfaces, classes,
// and anonymous object types all qualify).
func (t Type) IsObject() bool {
	return t.hasFlags(checker.TypeFlagsObject)
}

// IsArray reports whether the type is an array type.
func (t Type) IsArray() bool {
	return withCheckerT(t.project, func(c *checker.Checker) bool {
		return c.IsArrayType(t.t)
	})
}

// IsEnumLiteral reports whether the type is an enum literal type.
func (t Type) IsEnumLiteral() bool {
	return t.t != nil && t.t.IsEnumLiteral()
}

// IsStringLiteral reports whether the type is a string literal type.
func (t Type) IsStringLiteral() bool { return t.hasFlags(checker.TypeFlagsStringLiteral) }

// IsNumberLiteral reports whether the type is a number literal type.
func (t Type) IsNumberLiteral() bool { return t.hasFlags(checker.TypeFlagsNumberLiteral) }

// IsBooleanLiteral reports whether the type is a boolean literal type
// (`true` or `false`).
func (t Type) IsBooleanLiteral() bool { return t.hasFlags(checker.TypeFlagsBooleanLiteral) }

// IsLiteral reports whether the type is a literal type (string, number,
// bigint, or boolean literal).
func (t Type) IsLiteral() bool {
	return t.hasFlags(checker.TypeFlagsStringLiteral |
		checker.TypeFlagsNumberLiteral |
		checker.TypeFlagsBigIntLiteral |
		checker.TypeFlagsBooleanLiteral)
}

// IsNullable reports whether the type can be null or undefined.
func (t Type) IsNullable() bool {
	return withCheckerT(t.project, func(c *checker.Checker) bool {
		return c.IsNullableType(t.t)
	})
}

// GetNonNullableType returns the type with null/undefined removed.
func (t Type) GetNonNullableType() Type {
	return withCheckerT(t.project, func(c *checker.Checker) Type {
		nt := c.GetNonNullableType(t.t)
		return Type{t: nt, project: t.project}
	})
}

// AliasSymbol returns the symbol of the type alias (if the type came through
// an alias), or false. Distinct from Symbol, which returns the underlying
// (aliased) symbol.
func (t Type) AliasSymbol() (Symbol, bool) {
	if t.t == nil || t.t.Alias() == nil || t.t.Alias().Symbol() == nil {
		return Symbol{}, false
	}
	return Symbol{sym: t.t.Alias().Symbol(), project: t.project}, true
}

// BaseTypes returns the base types of the type (the types it `extends` or
// `implements`, resolved through the checker). Empty for non-class/interface
// types.
func (t Type) BaseTypes() []Type {
	return withCheckerT(t.project, func(c *checker.Checker) []Type {
		var out []Type
		for _, bt := range c.GetBaseTypes(t.t) {
			out = append(out, Type{t: bt, project: t.project})
		}
		return out
	})
}

// StringIndexType returns the value type of the type's string index
// signature (as used by `Record<string, ...>`), or false if there is none.
func (t Type) StringIndexType() (Type, bool) {
	return withCheckerT2(t.project, func(c *checker.Checker) (Type, bool) {
		it := c.GetStringIndexType(t.t)
		if it == nil {
			return Type{}, false
		}
		return Type{t: it, project: t.project}, true
	})
}

// TypeArguments returns the type arguments of an instantiated generic type.
func (t Type) TypeArguments() []Type {
	return withCheckerT(t.project, func(c *checker.Checker) []Type {
		var out []Type
		for _, ta := range c.GetTypeArguments(t.t) {
			out = append(out, Type{t: ta, project: t.project})
		}
		return out
	})
}

// UnionTypes returns the constituents of a union type, or nil.
func (t Type) UnionTypes() []Type {
	if !t.IsUnion() {
		return nil
	}
	var out []Type
	for _, ut := range t.t.Types() {
		out = append(out, Type{t: ut, project: t.project})
	}
	return out
}

// Property is a property of an object type.
type Property struct {
	// Name of the property.
	Name string
	// Type of the property.
	Type Type
	// ValueDeclaration is the declaration node for the property, or the zero
	// Node if the property has no declaration.
	ValueDeclaration Node
}

// Properties returns the declared properties of an object type.
func (t Type) Properties() []Property {
	if t.t == nil {
		return nil
	}
	return withCheckerT(t.project, func(c *checker.Checker) []Property {
		var out []Property
		for _, sym := range c.GetPropertiesOfType(t.t) {
			p := Property{
				Name: ast.EscapeAllInternalSymbolNames(sym.Name),
				Type: Type{t: c.GetTypeOfSymbol(sym), project: t.project},
			}
			if len(sym.Declarations) > 0 {
				if n, ok := t.propertyDeclaration(sym.Declarations[0]); ok {
					p.ValueDeclaration = n
				}
			}
			out = append(out, p)
		}
		return out
	})
}

// propertyDeclaration wraps a property declaration node belonging to this
// type's source file (or a sibling file in the same project).
func (t Type) propertyDeclaration(d *ast.Node) (Node, bool) {
	file := ast.GetSourceFileOfNode(d)
	if file == nil {
		return Node{}, false
	}
	sf := &SourceFile{project: t.project, path: file.FileName()}
	return sf.wrap(d)
}

// Signature wraps a call signature of a function type.
type Signature struct {
	sig     *checker.Signature
	project *Project
}

// Text returns a human-readable rendering of the signature.
func (s Signature) Text() string {
	if s.sig == nil {
		return ""
	}
	return withCheckerT(s.project, func(c *checker.Checker) string {
		return c.SignatureToStringEx(s.sig, nil, checker.TypeFormatFlagsNone, nil)
	})
}

// ReturnType returns the signature's return type.
func (s Signature) ReturnType() Type {
	if s.sig == nil {
		return Type{}
	}
	return withCheckerT(s.project, func(c *checker.Checker) Type {
		return Type{t: c.GetReturnTypeOfSignature(s.sig), project: s.project}
	})
}

// Parameter describes a parameter of a signature.
type Parameter struct {
	// Name of the parameter.
	Name string
	// Type of the parameter.
	Type Type
}

// Parameters returns the signature's parameters.
func (s Signature) Parameters() []Parameter {
	if s.sig == nil {
		return nil
	}
	return withCheckerT(s.project, func(c *checker.Checker) []Parameter {
		var out []Parameter
		for _, p := range s.sig.Parameters() {
			out = append(out, Parameter{
				Name: p.Name,
				Type: Type{t: c.GetTypeOfSymbol(p), project: s.project},
			})
		}
		return out
	})
}

// CallSignatures returns the call signatures of a function type.
func (t Type) CallSignatures() []Signature {
	if t.t == nil {
		return nil
	}
	return withCheckerT(t.project, func(c *checker.Checker) []Signature {
		var out []Signature
		for _, sig := range c.GetSignaturesOfType(t.t, checker.SignatureKindCall) {
			out = append(out, Signature{sig: sig, project: t.project})
		}
		return out
	})
}

// Symbol returns the symbol associated with the type, or false.
func (t Type) Symbol() (Symbol, bool) {
	if t.t == nil || t.t.Symbol() == nil {
		return Symbol{}, false
	}
	return Symbol{sym: t.t.Symbol(), project: t.project}, true
}

// Symbol wraps a compiler symbol.
type Symbol struct {
	sym     *ast.Symbol
	project *Project
}

// Name returns the symbol's name. Internal compiler names (which the
// vendored compiler encodes with a \xFE prefix, e.g. the anonymous type
// symbol) are escaped to their `__...` form to match ts-morph.
func (s Symbol) Name() string {
	if s.sym == nil {
		return ""
	}
	return ast.EscapeAllInternalSymbolNames(s.sym.Name)
}

// Equal reports whether two symbols refer to the same compiler symbol.
func (s Symbol) Equal(o Symbol) bool {
	return s.sym != nil && s.sym == o.sym
}

// Declarations returns the nodes declaring this symbol.
func (s Symbol) Declarations() []Node {
	if s.sym == nil {
		return nil
	}
	var out []Node
	for _, d := range s.sym.Declarations {
		file := ast.GetSourceFileOfNode(d)
		if file == nil {
			continue
		}
		sf := &SourceFile{project: s.project, path: file.FileName()}
		if n, ok := sf.wrap(d); ok {
			out = append(out, n)
		}
	}
	return out
}

// Type returns the type of the symbol.
func (s Symbol) Type() Type {
	if s.sym == nil {
		return Type{}
	}
	return withCheckerT(s.project, func(c *checker.Checker) Type {
		return Type{t: c.GetTypeOfSymbol(s.sym), project: s.project}
	})
}

// Type returns the type of the node, computed by the type checker.
func (n Node) Type() Type {
	return withCheckerT(n.sf.project, func(c *checker.Checker) Type {
		return Type{t: c.GetTypeAtLocation(n.node), project: n.sf.project}
	})
}

// Symbol returns the symbol associated with the node, or false. For
// declarations the symbol of the name node is returned.
func (n Node) Symbol() (result Symbol, ok bool) {
	target := n.node
	if name := ast.GetNameOfDeclaration(n.node); name != nil {
		target = name
	}
	n.sf.project.withChecker(func(c *checker.Checker) {
		sym := c.GetSymbolAtLocation(target)
		if sym == nil {
			return
		}
		result, ok = Symbol{sym: sym, project: n.sf.project}, true
	})
	return result, ok
}
