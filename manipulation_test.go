package tsmorph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// manipulationOps maps each testdata/manipulation/<case> directory to the
// operations to apply to its before.ts. The result must match after.ts
// byte-for-byte.
var manipulationOps = map[string]func(sf *SourceFile){
	"add-class": func(sf *SourceFile) {
		sf.AddClass(ClassStructure{
			Name:       "Person",
			IsExported: true,
			Properties: []PropertyStructure{{Name: "name", Type: "string"}},
			Constructors: []ConstructorStructure{{
				Parameters: []ParameterStructure{{Name: "name", Type: "string"}},
				Body:       "this.name = name;",
			}},
			Methods: []MethodStructure{{
				Name:       "greet",
				Parameters: []ParameterStructure{{Name: "greeting", Type: "string"}},
				ReturnType: "string",
				Body:       `return greeting + " " + this.name;`,
			}},
		})
	},
	"add-method": func(sf *SourceFile) {
		c, _ := sf.Class("Counter")
		c.AddMethod(MethodStructure{
			Name:       "increment",
			Parameters: []ParameterStructure{{Name: "by", Type: "number"}},
			ReturnType: "void",
			Body:       "this.count += by;",
		})
	},
	"add-interface": func(sf *SourceFile) {
		empty, _ := sf.Interface("Empty")
		empty.AddProperty(PropertyStructure{Name: "id", Type: "number"})
		named, _ := sf.Interface("Named")
		named.AddMethod(MethodStructure{
			Name:       "rename",
			Parameters: []ParameterStructure{{Name: "newName", Type: "string"}},
			ReturnType: "void",
		})
	},
	"add-imports": func(sf *SourceFile) {
		sf.AddImportDeclaration(ImportDeclarationStructure{
			ModuleSpecifier: "./bc",
			DefaultImport:   "def",
			NamedImports:    []string{"b", "c"},
		})
		sf.AddImportDeclaration(ImportDeclarationStructure{
			ModuleSpecifier: "./ns",
			NamespaceImport: "ns",
		})
	},
	"add-params-return-type": func(sf *SourceFile) {
		fn, _ := sf.Function("noop")
		fn.AddParameter(ParameterStructure{Name: "a", Type: "number"})
		fn2, _ := sf.Function("noop")
		fn2.AddParameter(ParameterStructure{Name: "b", Type: "string", IsOptional: true})
		fn3, _ := sf.Function("noop")
		fn3.SetReturnType("void")
	},
	"rename-class": func(sf *SourceFile) {
		c, _ := sf.Class("OldName")
		c.Rename("NewName")
	},
	"remove-node": func(sf *SourceFile) {
		for _, vs := range sf.VariableStatements() {
			for _, d := range vs.Declarations() {
				if d.Name() == "drop" {
					vs.Remove()
					return
				}
			}
		}
	},
	"add-enum-member": func(sf *SourceFile) {
		for _, e := range sf.Enums() {
			if e.Name() == "Color" {
				e.AddMember(EnumMemberStructure{Name: "Green", Initializer: "2"})
			}
		}
	},
	"replace-with-text": func(sf *SourceFile) {
		for _, vs := range sf.VariableStatements() {
			for _, d := range vs.Declarations() {
				if d.Name() == "value" {
					init, _ := d.Initializer()
					init.ReplaceWithText("compute()")
				}
			}
		}
	},
	"add-mixed": func(sf *SourceFile) {
		sf.AddImportDeclaration(ImportDeclarationStructure{
			ModuleSpecifier: "./helper",
			NamedImports:    []string{"helper"},
		})
		sf.AddTypeAlias(TypeAliasStructure{Name: "ID", IsExported: true, Type: "string"})
		sf.AddEnum(EnumStructure{
			Name:       "Status",
			IsExported: true,
			Members:    []EnumMemberStructure{{Name: "Active"}, {Name: "Inactive"}},
		})
		sf.AddVariableStatement(VariableStatementStructure{
			DeclarationKind: "const",
			IsExported:      true,
			Declarations: []VariableDeclarationStructure{{
				Name:        "DEFAULT_ID",
				Type:        "ID",
				Initializer: `"none"`,
			}},
		})
		sf.AddFunction(FunctionStructure{
			Name:       "isActive",
			IsExported: true,
			Parameters: []ParameterStructure{{Name: "status", Type: "Status"}},
			ReturnType: "boolean",
			Body:       "return status === Status.Active;",
		})
	},
}

func TestManipulationGolden(t *testing.T) {
	base := "testdata/manipulation"
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		op, ok := manipulationOps[name]
		if !ok {
			t.Errorf("no operations registered for case %q", name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			before, err := os.ReadFile(filepath.Join(base, name, "before.ts"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(base, name, "after.ts"))
			if err != nil {
				t.Fatal(err)
			}

			p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
			if err != nil {
				t.Fatal(err)
			}
			sf := p.CreateSourceFile("/case.ts", string(before))
			op(sf)

			if got := sf.Text(); got != string(want) {
				t.Errorf("output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
			}
		})
	}
}

func TestManipulationForgetSemantics(t *testing.T) {
	p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.CreateSourceFile("/a.ts", "export class A {\n  x: number;\n}\n")

	c, ok := sf.Class("A")
	if !ok {
		t.Fatal("class A not found")
	}
	if c.IsForgotten() {
		t.Fatal("class should not be forgotten before edits")
	}

	// Mutate the file: the old wrapper must be forgotten.
	sf.AddVariableStatement(VariableStatementStructure{
		Declarations: []VariableDeclarationStructure{{Name: "b"}},
	})
	if !c.IsForgotten() {
		t.Fatal("class should be forgotten after edit")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when using a forgotten node")
		} else if msg, _ := r.(string); !strings.Contains(msg, "forgotten") {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	_ = c.Name() // must panic
}

func TestManipulationCheckerSeesEdits(t *testing.T) {
	p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.CreateSourceFile("/a.ts", "export const x: number = 1;\n")

	// No diagnostics initially.
	if diags := p.PreEmitDiagnostics(); len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Introduce a type error via edit; the rebuilt program must see it.
	for _, vs := range sf.VariableStatements() {
		vs.Declarations() // exercise navigation pre-edit
	}
	sf.AddVariableStatement(VariableStatementStructure{
		Declarations: []VariableDeclarationStructure{{
			Name:        "y",
			Type:        "string",
			Initializer: "42",
		}},
	})

	var found bool
	for _, d := range p.PreEmitDiagnostics() {
		if d.Code == 2322 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected TS2322 after edit")
	}

	// New nodes work and types reflect the new program.
	vs := sf.VariableStatements()[1]
	decl := vs.Declarations()[0]
	if got := decl.Type().Text(); got != "string" {
		t.Fatalf("y type: %q", got)
	}
}
