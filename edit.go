package tsmorph

import (
	"strings"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

// This file implements the mutating API. Every public method computes text
// edits from the current AST and applies them immediately (matching
// ts-morph's eager model): the file is re-parsed and previously obtained
// Node wrappers for the file are forgotten. Methods that return a node
// re-fetch it from the freshly parsed tree.

// --- SourceFile level insertions ---

// insertStatementText inserts rendered text at the given statement index.
// An index >= the statement count appends at the end of the file.
func (s *SourceFile) insertStatementText(index int, text string) {
	stmts := s.Statements()
	nl := string(s.project.ManipulationSettings().NewLineKind)

	if index >= len(stmts) {
		fileText := s.Text()
		insertPos := len(fileText)
		prefix := ""
		if strings.TrimSpace(fileText) != "" {
			if strings.HasSuffix(fileText, nl) {
				prefix = nl // blank line between declarations
			} else {
				prefix = nl + nl
			}
		}
		s.applyEdits([]textEdit{{start: insertPos, end: insertPos, newText: prefix + text + nl}})
		return
	}

	pos := stmts[index].Start()
	s.applyEdits([]textEdit{{start: pos, end: pos, newText: text + nl + nl}})
}

// InsertClass inserts a class declaration at the given statement index.
func (s *SourceFile) InsertClass(index int, structure ClassStructure) ClassDeclaration {
	s.insertStatementText(index, newCodeWriter(s.project).classDecl(structure))
	c, _ := s.Class(structure.Name)
	return c
}

// AddClass appends a class declaration to the end of the file.
func (s *SourceFile) AddClass(structure ClassStructure) ClassDeclaration {
	return s.InsertClass(len(s.Statements()), structure)
}

// InsertInterface inserts an interface declaration at the given statement index.
func (s *SourceFile) InsertInterface(index int, structure InterfaceStructure) InterfaceDeclaration {
	s.insertStatementText(index, newCodeWriter(s.project).interfaceDecl(structure))
	i, _ := s.Interface(structure.Name)
	return i
}

// AddInterface appends an interface declaration to the end of the file.
func (s *SourceFile) AddInterface(structure InterfaceStructure) InterfaceDeclaration {
	return s.InsertInterface(len(s.Statements()), structure)
}

// InsertFunction inserts a function declaration at the given statement index.
func (s *SourceFile) InsertFunction(index int, structure FunctionStructure) FunctionDeclaration {
	s.insertStatementText(index, newCodeWriter(s.project).functionDecl(structure))
	f, _ := s.Function(structure.Name)
	return f
}

// AddFunction appends a function declaration to the end of the file.
func (s *SourceFile) AddFunction(structure FunctionStructure) FunctionDeclaration {
	return s.InsertFunction(len(s.Statements()), structure)
}

// AddEnum appends an enum declaration to the end of the file.
func (s *SourceFile) AddEnum(structure EnumStructure) EnumDeclaration {
	s.insertStatementText(len(s.Statements()), newCodeWriter(s.project).enumDecl(structure))
	for _, e := range s.Enums() {
		if e.Name() == structure.Name {
			return e
		}
	}
	return EnumDeclaration{}
}

// AddTypeAlias appends a type alias declaration to the end of the file.
func (s *SourceFile) AddTypeAlias(structure TypeAliasStructure) TypeAliasDeclaration {
	s.insertStatementText(len(s.Statements()), newCodeWriter(s.project).typeAliasDecl(structure))
	for _, t := range s.TypeAliases() {
		if t.Name() == structure.Name {
			return t
		}
	}
	return TypeAliasDeclaration{}
}

// AddVariableStatement appends a variable statement to the end of the file.
func (s *SourceFile) AddVariableStatement(structure VariableStatementStructure) VariableStatement {
	s.insertStatementText(len(s.Statements()), newCodeWriter(s.project).variableStatement(structure))
	vs := s.VariableStatements()
	return vs[len(vs)-1]
}

// AddImportDeclaration adds an import declaration after the last existing
// import, or at the top of the file.
func (s *SourceFile) AddImportDeclaration(structure ImportDeclarationStructure) ImportDeclaration {
	text := newCodeWriter(s.project).importDecl(structure)
	nl := string(s.project.ManipulationSettings().NewLineKind)

	imports := s.ImportDeclarations()
	if len(imports) > 0 {
		last := imports[len(imports)-1]
		s.applyEdits([]textEdit{{start: last.End(), end: last.End(), newText: nl + text}})
	} else {
		s.applyEdits([]textEdit{{start: 0, end: 0, newText: text + nl}})
	}

	for _, imp := range s.ImportDeclarations() {
		if strings.TrimSpace(imp.Text()) == text {
			return imp
		}
	}
	return ImportDeclaration{}
}

// AddExportDeclaration appends an `export ... from` declaration to the end
// of the file.
func (s *SourceFile) AddExportDeclaration(structure ExportDeclarationStructure) ExportDeclaration {
	s.insertStatementText(len(s.Statements()), newCodeWriter(s.project).exportDecl(structure))
	eds := s.ExportDeclarations()
	return eds[len(eds)-1]
}

// --- Member insertion helpers ---

// memberIndent returns the indentation to use for members of the node.
func (n Node) memberIndent() string {
	text := n.sf.Text()
	return indentOfLine(text, n.Start()) + n.sf.project.ManipulationSettings().IndentationText
}

// baseIndent returns the indentation of the line the node starts on.
func (n Node) baseIndent() string {
	return indentOfLine(n.sf.Text(), n.Start())
}

// openBracePos returns the offset just past the first '{' of the node.
func (n Node) openBracePos() int {
	text := n.sf.Text()
	for i := n.Start(); i < n.End(); i++ {
		if text[i] == '{' {
			return i + 1
		}
	}
	return -1
}

// insertMember inserts rendered member text at the given member index of a
// braced node (class, interface, enum). The first line of memberText is
// indented by insertMember; continuation lines must already carry full
// indentation (use renderMethod/renderProperty with memberIndent()).
func (n Node) insertMember(index int, memberText string) {
	nl := string(n.sf.project.ManipulationSettings().NewLineKind)
	members := n.node.Members()
	indent := n.memberIndent()

	if len(members) == 0 {
		open := n.openBracePos()
		if open < 0 {
			panic("tsmorph: could not find opening brace of " + n.KindName())
		}
		n.sf.applyEdits([]textEdit{{start: open, end: open, newText: nl + indent + memberText + nl + n.baseIndent()}})
		return
	}

	if index < len(members) {
		member := Node{node: members[index], sf: n.sf, gen: n.gen}
		pos := member.Start()
		n.sf.applyEdits([]textEdit{{start: pos, end: pos, newText: memberText + nl + indent}})
		return
	}

	last := Node{node: members[len(members)-1], sf: n.sf, gen: n.gen}
	// Separate the new member from the previous one with a blank line.
	n.sf.applyEdits([]textEdit{{start: last.End(), end: last.End(), newText: nl + nl + indent + memberText}})
}

// renderMethod renders a method for insertion. indent is the indentation of
// the method's first line (used to indent body lines and the closing brace).
func renderMethod(p *Project, m MethodStructure, indent string) string {
	w := newCodeWriter(p)
	w.methodSignature(m)
	w.bracedBody(m.Body, indent)
	return w.String()
}

// renderProperty renders a single-line property.
func renderProperty(p *Project, prop PropertyStructure) string {
	w := newCodeWriter(p)
	w.property(prop)
	w.write(";")
	return w.String()
}

// --- Class member edits ---

// InsertMethod inserts a method into the class at the given member index.
func (c ClassDeclaration) InsertMethod(index int, structure MethodStructure) MethodDeclaration {
	c.check()
	className := c.Name()
	c.insertMember(index, renderMethod(c.sf.project, structure, c.memberIndent()))
	cls, _ := c.sf.Class(className)
	for _, m := range cls.Methods() {
		if m.Name() == structure.Name {
			return m
		}
	}
	return MethodDeclaration{}
}

// AddMethod appends a method to the class.
func (c ClassDeclaration) AddMethod(structure MethodStructure) MethodDeclaration {
	c.check()
	return c.InsertMethod(len(c.node.Members()), structure)
}

// InsertProperty inserts a property into the class at the given member index.
func (c ClassDeclaration) InsertProperty(index int, structure PropertyStructure) PropertyDeclaration {
	c.check()
	className := c.Name()
	c.insertMember(index, renderProperty(c.sf.project, structure))
	cls, _ := c.sf.Class(className)
	for _, prop := range cls.Properties() {
		if prop.Name() == structure.Name {
			return prop
		}
	}
	return PropertyDeclaration{}
}

// AddProperty appends a property to the class.
func (c ClassDeclaration) AddProperty(structure PropertyStructure) PropertyDeclaration {
	c.check()
	return c.InsertProperty(len(c.node.Members()), structure)
}

// AddConstructor inserts a constructor as the first member of the class.
func (c ClassDeclaration) AddConstructor(structure ConstructorStructure) ConstructorDeclaration {
	c.check()
	className := c.Name()
	indent := c.memberIndent()
	w := newCodeWriter(c.sf.project)
	w.write("constructor")
	w.params(structure.Parameters)
	w.bracedBody(structure.Body, indent)
	c.insertMember(0, w.String())
	cls, _ := c.sf.Class(className)
	ctors := cls.Constructors()
	if len(ctors) == 0 {
		return ConstructorDeclaration{}
	}
	cd, _ := ctors[0].AsConstructorDeclaration()
	return cd
}

// --- Interface member edits ---

// AddMethod appends a method signature to the interface.
func (i InterfaceDeclaration) AddMethod(structure MethodStructure) Node {
	i.check()
	name := i.Name()
	w := newCodeWriter(i.sf.project)
	w.methodSignature(structure)
	w.write(";")
	i.insertMember(len(i.node.Members()), w.String())
	iface, _ := i.sf.Interface(name)
	members := iface.Members()
	return members[len(members)-1]
}

// AddProperty appends a property signature to the interface.
func (i InterfaceDeclaration) AddProperty(structure PropertyStructure) Node {
	i.check()
	name := i.Name()
	i.insertMember(len(i.node.Members()), renderProperty(i.sf.project, structure))
	iface, _ := i.sf.Interface(name)
	members := iface.Members()
	return members[len(members)-1]
}

// --- Enum member edits ---

// AddMember appends a member to the enum.
func (e EnumDeclaration) AddMember(structure EnumMemberStructure) EnumMember {
	e.check()
	enumName := e.Name()
	text := structure.Name
	if structure.Initializer != "" {
		text += " = " + structure.Initializer
	}
	members := e.node.Members()
	if len(members) > 0 {
		last := Node{node: members[len(members)-1], sf: e.sf, gen: e.gen}
		indent := e.memberIndent()
		nl := string(e.sf.project.ManipulationSettings().NewLineKind)
		srcText := e.sf.Text()
		// A separating comma (if present) sits just past the member's End.
		if last.End() < len(srcText) && srcText[last.End()] == ',' {
			pos := last.End() + 1
			e.sf.applyEdits([]textEdit{{start: pos, end: pos, newText: nl + indent + text + ","}})
		} else {
			e.sf.applyEdits([]textEdit{{start: last.End(), end: last.End(), newText: "," + nl + indent + text + ","}})
		}
		return findEnumMember(e.sf, enumName, structure.Name)
	}
	e.insertMember(0, text+",")
	return findEnumMember(e.sf, enumName, structure.Name)
}

func findEnumMember(sf *SourceFile, enumName, memberName string) EnumMember {
	for _, en := range sf.Enums() {
		if en.Name() != enumName {
			continue
		}
		for _, m := range en.Members() {
			if m.Name() == memberName {
				return m
			}
		}
	}
	return EnumMember{}
}

// --- Parameter and return type edits ---

// openParenPos returns the offset of the '(' opening the parameter list.
func (n Node) openParenPos() int {
	n.check()
	text := n.sf.Text()
	depth := 0
	for i := n.Start(); i < n.End(); i++ {
		switch text[i] {
		case '<':
			depth++
		case '>':
			if depth > 0 {
				depth--
			}
		case '(':
			if depth == 0 {
				return i
			}
		}
	}
	panic("tsmorph: could not find parameter list of " + n.KindName())
}

// closeParenPos returns the offset of the ')' closing the parameter list.
func (n Node) closeParenPos() int {
	open := n.openParenPos()
	text := n.sf.Text()
	depth := 0
	for i := open; i < n.End(); i++ {
		switch text[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	panic("tsmorph: could not find closing paren of " + n.KindName())
}

// renderParam renders a single parameter.
func renderParam(p ParameterStructure) string {
	var sb strings.Builder
	if p.IsReadonly {
		sb.WriteString("readonly ")
	}
	sb.WriteString(p.Name)
	if p.IsOptional {
		sb.WriteString("?")
	}
	if p.Type != "" {
		sb.WriteString(": " + p.Type)
	}
	if p.Initializer != "" {
		sb.WriteString(" = " + p.Initializer)
	}
	return sb.String()
}

// addParameter appends a parameter to a function-like node.
func (n Node) addParameter(p ParameterStructure) {
	params := n.node.Parameters()
	text := renderParam(p)
	if len(params) == 0 {
		open := n.openParenPos()
		n.sf.applyEdits([]textEdit{{start: open + 1, end: open + 1, newText: text}})
		return
	}
	last := Node{node: params[len(params)-1], sf: n.sf, gen: n.gen}
	n.sf.applyEdits([]textEdit{{start: last.End(), end: last.End(), newText: ", " + text}})
}

// setReturnType sets or replaces the return type of a function-like node.
func (n Node) setReturnType(typeText string) {
	if rt := n.node.Type(); rt != nil {
		rtNode := Node{node: rt, sf: n.sf, gen: n.gen}
		n.sf.replaceRange(rtNode.Start(), rtNode.End(), typeText)
		return
	}
	close := n.closeParenPos()
	n.sf.applyEdits([]textEdit{{start: close + 1, end: close + 1, newText: ": " + typeText}})
}

// AddParameter appends a parameter to the function.
func (f FunctionDeclaration) AddParameter(p ParameterStructure) ParameterDeclaration {
	f.check()
	name := f.Name()
	f.addParameter(p)
	fn, _ := f.sf.Function(name)
	return lastParamOf(fn.Parameters())
}

// AddParameter appends a parameter to the method.
func (m MethodDeclaration) AddParameter(p ParameterStructure) ParameterDeclaration {
	m.check()
	meth := m.refetchByName()
	m.addParameter(p)
	return lastParamOf(meth.refetchByName().Parameters())
}

// AddParameter appends a parameter to the constructor.
func (c ConstructorDeclaration) AddParameter(p ParameterStructure) ParameterDeclaration {
	c.check()
	ctor := c.refetch()
	ctor.addParameter(p)
	return lastParamOf(c.refetch().Parameters())
}

// SetReturnType sets the declared return type of the function.
func (f FunctionDeclaration) SetReturnType(typeText string) {
	f.check()
	f.setReturnType(typeText)
}

// SetReturnType sets the declared return type of the method.
func (m MethodDeclaration) SetReturnType(typeText string) {
	m.check()
	m.setReturnType(typeText)
}

// refetchByName re-obtains the method from the freshly parsed tree by
// locating its containing class and name.
func (m MethodDeclaration) refetchByName() MethodDeclaration {
	name := m.Name()
	class, ok := m.FirstAncestorByKind(ast.KindClassDeclaration)
	if !ok {
		return m
	}
	className := class.Name()
	cls, ok := m.sf.Class(className)
	if !ok {
		return m
	}
	for _, meth := range cls.Methods() {
		if meth.Name() == name {
			return meth
		}
	}
	return m
}

// refetch re-obtains the constructor from the freshly parsed tree.
func (c ConstructorDeclaration) refetch() ConstructorDeclaration {
	class, ok := c.FirstAncestorByKind(ast.KindClassDeclaration)
	if !ok {
		return c
	}
	cls, ok := c.sf.Class(class.Name())
	if !ok {
		return c
	}
	ctors := cls.Constructors()
	if len(ctors) == 0 {
		return c
	}
	cd, _ := ctors[0].AsConstructorDeclaration()
	return cd
}

func lastParamOf(params []ParameterDeclaration) ParameterDeclaration {
	if len(params) == 0 {
		return ParameterDeclaration{}
	}
	return params[len(params)-1]
}

// --- Generic node edits ---

// ReplaceWithText replaces the node's text range with new text.
func (n Node) ReplaceWithText(text string) {
	n.check()
	n.sf.replaceRange(n.Start(), n.End(), text)
}

// Remove deletes the node from the source file. If the node occupies whole
// lines, the lines are removed entirely (including the trailing newline).
func (n Node) Remove() {
	n.check()
	text := n.sf.Text()
	start, end := n.Start(), n.End()

	ls := lineStart(text, start)
	if isBlank(text[ls:start]) {
		start = ls
	}
	le := lineEnd(text, end)
	if le < len(text) && isBlank(text[end:le]) {
		end = le + 1 // include the newline
	}
	n.sf.replaceRange(start, end, "")
}

// Rename changes the name of a declaration. It only renames the declaration
// site identifier; references elsewhere are NOT updated (unlike ts-morph's
// cross-file rename).
func (n Node) Rename(newName string) {
	n.check()
	name, ok := n.nameNode()
	if !ok {
		panic("tsmorph: node of kind " + n.KindName() + " has no name to rename")
	}
	n.sf.replaceRange(name.Start(), name.End(), newName)
}
