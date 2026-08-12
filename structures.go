package tsmorph

import (
	"strings"
)

// Structure types describe code to be generated. They are plain data;
// rendering honours the project's ManipulationSettings.

// ParameterStructure describes a parameter.
type ParameterStructure struct {
	Name        string
	Type        string // optional
	Initializer string // optional
	IsOptional  bool
	IsReadonly  bool
}

// MethodStructure describes a method (in a class) or method signature (in an
// interface; Body is then ignored).
type MethodStructure struct {
	Name       string
	Parameters []ParameterStructure
	ReturnType string // optional; inferred when empty
	Body       string // optional; defaults to an empty body
	IsStatic   bool
	IsAsync    bool
}

// PropertyStructure describes a property.
type PropertyStructure struct {
	Name        string
	Type        string // optional
	Initializer string // optional
	IsStatic    bool
	IsReadonly  bool
	IsOptional  bool
}

// ConstructorStructure describes a class constructor.
type ConstructorStructure struct {
	Parameters []ParameterStructure
	Body       string
}

// ClassStructure describes a class declaration.
type ClassStructure struct {
	Name            string
	IsExported      bool
	IsDefaultExport bool
	Extends         string   // optional
	Implements      []string // optional
	Methods         []MethodStructure
	Properties      []PropertyStructure
	Constructors    []ConstructorStructure
}

// InterfaceStructure describes an interface declaration.
type InterfaceStructure struct {
	Name       string
	IsExported bool
	Extends    []string // optional
	Methods    []MethodStructure
	Properties []PropertyStructure
}

// FunctionStructure describes a function declaration.
type FunctionStructure struct {
	Name            string
	IsExported      bool
	IsDefaultExport bool
	IsAsync         bool
	Parameters      []ParameterStructure
	ReturnType      string // optional
	Body            string // optional
}

// EnumMemberStructure describes an enum member.
type EnumMemberStructure struct {
	Name        string
	Initializer string // optional
}

// EnumStructure describes an enum declaration.
type EnumStructure struct {
	Name       string
	IsExported bool
	Members    []EnumMemberStructure
}

// TypeAliasStructure describes a type alias declaration.
type TypeAliasStructure struct {
	Name       string
	IsExported bool
	Type       string
}

// VariableDeclarationStructure describes a single declarator.
type VariableDeclarationStructure struct {
	Name        string
	Type        string // optional
	Initializer string // optional
}

// VariableStatementStructure describes a variable statement.
type VariableStatementStructure struct {
	// DeclarationKind is "const" (default), "let", or "var".
	DeclarationKind string
	IsExported      bool
	Declarations    []VariableDeclarationStructure
}

// ImportDeclarationStructure describes an import declaration.
type ImportDeclarationStructure struct {
	ModuleSpecifier string
	DefaultImport   string   // optional
	NamespaceImport string   // optional; e.g. "ns" for `import * as ns`
	NamedImports    []string // optional
	// IsTypeOnly renders `import type { ... } from "..."`.
	IsTypeOnly bool
}

// ExportDeclarationStructure describes an `export ... from` declaration.
type ExportDeclarationStructure struct {
	ModuleSpecifier   string   // optional
	NamedExports      []string // optional
	IsNamespaceExport bool     // `export * from "x"`
}

// codeWriter renders structures to text honouring manipulation settings.
type codeWriter struct {
	settings ManipulationSettings
	sb       strings.Builder
}

func newCodeWriter(p *Project) *codeWriter {
	return &codeWriter{settings: p.ManipulationSettings()}
}

func (w *codeWriter) write(s string) { w.sb.WriteString(s) }
func (w *codeWriter) nl()            { w.sb.WriteString(string(w.settings.NewLineKind)) }
func (w *codeWriter) String() string { return w.sb.String() }
func (w *codeWriter) quote(s string) string {
	return string(w.settings.QuoteKind) + s + string(w.settings.QuoteKind)
}

func (w *codeWriter) exportKeyword(isExported bool) {
	if isExported {
		w.write("export ")
	}
}

func (w *codeWriter) params(params []ParameterStructure) {
	w.write("(")
	for i, p := range params {
		if i > 0 {
			w.write(", ")
		}
		if p.IsReadonly {
			w.write("readonly ")
		}
		w.write(p.Name)
		if p.IsOptional {
			w.write("?")
		}
		if p.Type != "" {
			w.write(": " + p.Type)
		}
		if p.Initializer != "" {
			w.write(" = " + p.Initializer)
		}
	}
	w.write(")")
}

func (w *codeWriter) returnType(rt string) {
	if rt != "" {
		w.write(": " + rt)
	}
}

// bracedBody writes " { ... }" with body lines indented one level.
// body may be empty, producing "{ }" for one-line empties.
func (w *codeWriter) bracedBody(body string, indent string) {
	inner := indent + w.settings.IndentationText
	if strings.TrimSpace(body) == "" {
		w.write(" {}")
		return
	}
	w.write(" {")
	w.nl()
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			w.nl()
			continue
		}
		w.write(inner + line)
		w.nl()
	}
	w.write(indent + "}")
}

func (w *codeWriter) methodSignature(m MethodStructure) {
	if m.IsStatic {
		w.write("static ")
	}
	if m.IsAsync {
		w.write("async ")
	}
	w.write(m.Name)
	w.params(m.Parameters)
	w.returnType(m.ReturnType)
}

func (w *codeWriter) property(p PropertyStructure) {
	if p.IsStatic {
		w.write("static ")
	}
	if p.IsReadonly {
		w.write("readonly ")
	}
	w.write(p.Name)
	if p.IsOptional {
		w.write("?")
	}
	if p.Type != "" {
		w.write(": " + p.Type)
	}
	if p.Initializer != "" {
		w.write(" = " + p.Initializer)
	}
}

// --- Top-level structure renderers ---

func (w *codeWriter) classDecl(c ClassStructure) string {
	w.exportKeyword(c.IsExported)
	if c.IsDefaultExport {
		w.write("default ")
	}
	w.write("class " + c.Name)
	if c.Extends != "" {
		w.write(" extends " + c.Extends)
	}
	if len(c.Implements) > 0 {
		w.write(" implements " + strings.Join(c.Implements, ", "))
	}
	inner := w.settings.IndentationText
	if len(c.Constructors) == 0 && len(c.Properties) == 0 && len(c.Methods) == 0 {
		w.write(" {}")
		return w.String()
	}
	w.write(" {")
	w.nl()
	first := true
	sep := func() {
		if !first {
			w.nl()
		}
		first = false
	}
	for _, p := range c.Properties {
		sep()
		w.write(inner)
		w.property(p)
		w.write(";")
		w.nl()
	}
	for _, ctor := range c.Constructors {
		sep()
		w.write(inner + "constructor")
		w.params(ctor.Parameters)
		w.bracedBody(ctor.Body, inner)
		w.nl()
	}
	for _, m := range c.Methods {
		sep()
		w.write(inner)
		w.methodSignature(m)
		w.bracedBody(m.Body, inner)
		w.nl()
	}
	w.write("}")
	return w.String()
}

func (w *codeWriter) interfaceDecl(i InterfaceStructure) string {
	w.exportKeyword(i.IsExported)
	w.write("interface " + i.Name)
	if len(i.Extends) > 0 {
		w.write(" extends " + strings.Join(i.Extends, ", "))
	}
	if len(i.Methods) == 0 && len(i.Properties) == 0 {
		w.write(" {}")
		return w.String()
	}
	inner := w.settings.IndentationText
	w.write(" {")
	w.nl()
	for _, p := range i.Properties {
		w.write(inner)
		w.property(p)
		w.write(";")
		w.nl()
	}
	for _, m := range i.Methods {
		w.write(inner)
		w.methodSignature(m)
		w.write(";")
		w.nl()
	}
	w.write("}")
	return w.String()
}

func (w *codeWriter) functionDecl(f FunctionStructure) string {
	w.exportKeyword(f.IsExported)
	if f.IsDefaultExport {
		w.write("default ")
	}
	if f.IsAsync {
		w.write("async ")
	}
	w.write("function " + f.Name)
	w.params(f.Parameters)
	w.returnType(f.ReturnType)
	w.bracedBody(f.Body, "")
	return w.String()
}

func (w *codeWriter) enumDecl(e EnumStructure) string {
	w.exportKeyword(e.IsExported)
	w.write("enum " + e.Name + " {")
	w.nl()
	inner := w.settings.IndentationText
	for i, m := range e.Members {
		w.write(inner + m.Name)
		if m.Initializer != "" {
			w.write(" = " + m.Initializer)
		}
		if i < len(e.Members)-1 || w.settings.UseTrailingCommas {
			w.write(",")
		}
		w.nl()
	}
	w.write("}")
	return w.String()
}

func (w *codeWriter) typeAliasDecl(t TypeAliasStructure) string {
	w.exportKeyword(t.IsExported)
	w.write("type " + t.Name + " = " + t.Type + ";")
	return w.String()
}

func (w *codeWriter) variableStatement(v VariableStatementStructure) string {
	w.exportKeyword(v.IsExported)
	kind := v.DeclarationKind
	if kind == "" {
		kind = "const"
	}
	w.write(kind + " ")
	for i, d := range v.Declarations {
		if i > 0 {
			w.write(", ")
		}
		w.write(d.Name)
		if d.Type != "" {
			w.write(": " + d.Type)
		}
		if d.Initializer != "" {
			w.write(" = " + d.Initializer)
		}
	}
	w.write(";")
	return w.String()
}

func (w *codeWriter) importDecl(i ImportDeclarationStructure) string {
	w.write("import ")
	if i.IsTypeOnly {
		w.write("type ")
	}
	var parts []string
	if i.DefaultImport != "" {
		parts = append(parts, i.DefaultImport)
	}
	if i.NamespaceImport != "" {
		parts = append(parts, "* as "+i.NamespaceImport)
	}
	if len(i.NamedImports) > 0 {
		parts = append(parts, "{ "+strings.Join(i.NamedImports, ", ")+" }")
	}
	if len(parts) > 0 {
		w.write(strings.Join(parts, ", ") + " from ")
	}
	w.write(w.quote(i.ModuleSpecifier) + ";")
	return w.String()
}

func (w *codeWriter) exportDecl(e ExportDeclarationStructure) string {
	w.write("export ")
	if e.IsNamespaceExport {
		w.write("*")
	} else {
		w.write("{ " + strings.Join(e.NamedExports, ", ") + " }")
	}
	if e.ModuleSpecifier != "" {
		w.write(" from " + w.quote(e.ModuleSpecifier))
	}
	w.write(";")
	return w.String()
}
