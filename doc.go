// Package tsmorph is a Go port of ts-morph (https://github.com/dsherret/ts-morph)
// built on the native Go port of the TypeScript compiler
// (microsoft/typescript-go, vendored under third_party/typescript-go).
//
// # Quick start
//
//	p, err := tsmorph.NewProject(tsmorph.ProjectOptions{
//		TsConfigFilePath: "tsconfig.json",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, sf := range p.SourceFiles() {
//		for _, c := range sf.Classes() {
//			fmt.Println(c.Name())
//		}
//	}
//
// # Navigation
//
// SourceFile exposes top-level declarations (Classes, Interfaces, Functions,
// Enums, TypeAliases, VariableStatements, ImportDeclarations,
// ExportDeclarations). Node supports arbitrary traversal via Children,
// Descendants, DescendantsOfKind, and FirstAncestorByKind, plus typed
// downcasts (AsClassDeclaration, ...).
//
// # Types
//
// Node.Type() and Node.Symbol() query the type checker;
// Project.PreEmitDiagnostics reports compiler diagnostics.
//
// # Manipulation
//
// Mutations (AddClass, AddMethod, Rename, Remove, ReplaceWithText, ...)
// apply immediately: the file's text is updated in memory and re-parsed,
// and previously obtained Node wrappers for that file are forgotten (using
// one panics). Re-fetch nodes from the SourceFile after each edit. Edits
// reach disk only via Project.Save or SourceFile.Save.
package tsmorph
