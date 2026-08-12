package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

// Navigation methods on SourceFile.

// RootNode returns the file's root as a Node, enabling arbitrary traversal.
func (s *SourceFile) RootNode() Node {
	n, _ := s.wrap(s.astFile().AsNode())
	return n
}

// Statements returns the top-level statements of the file.
func (s *SourceFile) Statements() []Node {
	var out []Node
	for _, st := range s.astFile().Statements.Nodes {
		if n, ok := s.wrap(st); ok {
			out = append(out, n)
		}
	}
	return out
}

// topLevel returns wrapped top-level statements of the given kinds.
func (s *SourceFile) topLevel(kinds ...Kind) []Node {
	var out []Node
	for _, st := range s.astFile().Statements.Nodes {
		for _, k := range kinds {
			if st.Kind == k {
				if n, ok := s.wrap(st); ok {
					out = append(out, n)
				}
				break
			}
		}
	}
	return out
}

// Classes returns the top-level class declarations.
func (s *SourceFile) Classes() []ClassDeclaration {
	var out []ClassDeclaration
	for _, n := range s.topLevel(ast.KindClassDeclaration) {
		c, _ := n.AsClassDeclaration()
		out = append(out, c)
	}
	return out
}

// Class returns the top-level class with the given name, or false.
func (s *SourceFile) Class(name string) (ClassDeclaration, bool) {
	for _, c := range s.Classes() {
		if c.Name() == name {
			return c, true
		}
	}
	return ClassDeclaration{}, false
}

// Interfaces returns the top-level interface declarations.
func (s *SourceFile) Interfaces() []InterfaceDeclaration {
	var out []InterfaceDeclaration
	for _, n := range s.topLevel(ast.KindInterfaceDeclaration) {
		i, _ := n.AsInterfaceDeclaration()
		out = append(out, i)
	}
	return out
}

// Interface returns the top-level interface with the given name, or false.
func (s *SourceFile) Interface(name string) (InterfaceDeclaration, bool) {
	for _, i := range s.Interfaces() {
		if i.Name() == name {
			return i, true
		}
	}
	return InterfaceDeclaration{}, false
}

// Functions returns the top-level function declarations.
func (s *SourceFile) Functions() []FunctionDeclaration {
	var out []FunctionDeclaration
	for _, n := range s.topLevel(ast.KindFunctionDeclaration) {
		f, _ := n.AsFunctionDeclaration()
		out = append(out, f)
	}
	return out
}

// Function returns the top-level function with the given name, or false.
func (s *SourceFile) Function(name string) (FunctionDeclaration, bool) {
	for _, f := range s.Functions() {
		if f.Name() == name {
			return f, true
		}
	}
	return FunctionDeclaration{}, false
}

// Enums returns the top-level enum declarations.
func (s *SourceFile) Enums() []EnumDeclaration {
	var out []EnumDeclaration
	for _, n := range s.topLevel(ast.KindEnumDeclaration) {
		e, _ := n.AsEnumDeclaration()
		out = append(out, e)
	}
	return out
}

// TypeAliases returns the top-level type alias declarations.
func (s *SourceFile) TypeAliases() []TypeAliasDeclaration {
	var out []TypeAliasDeclaration
	for _, n := range s.topLevel(ast.KindTypeAliasDeclaration) {
		t, _ := n.AsTypeAliasDeclaration()
		out = append(out, t)
	}
	return out
}

// VariableStatements returns the top-level variable statements.
func (s *SourceFile) VariableStatements() []VariableStatement {
	var out []VariableStatement
	for _, n := range s.topLevel(ast.KindVariableStatement) {
		v, _ := n.AsVariableStatement()
		out = append(out, v)
	}
	return out
}

// ImportDeclarations returns the file's import declarations.
func (s *SourceFile) ImportDeclarations() []ImportDeclaration {
	var out []ImportDeclaration
	for _, n := range s.topLevel(ast.KindImportDeclaration) {
		i, _ := n.AsImportDeclaration()
		out = append(out, i)
	}
	return out
}

// ExportDeclarations returns the file's `export ... from` declarations.
func (s *SourceFile) ExportDeclarations() []ExportDeclaration {
	var out []ExportDeclaration
	for _, n := range s.topLevel(ast.KindExportDeclaration) {
		e, _ := n.AsExportDeclaration()
		out = append(out, e)
	}
	return out
}

// DescendantsOfKind returns all descendants of the file with the given kind.
func (s *SourceFile) DescendantsOfKind(kind Kind) []Node {
	return s.RootNode().DescendantsOfKind(kind)
}

// FirstDescendantByKind returns the first descendant of the file with the
// given kind, or false.
func (s *SourceFile) FirstDescendantByKind(kind Kind) (Node, bool) {
	return s.RootNode().FirstDescendantByKind(kind)
}
