package tsmorph

import (
	"context"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/diagnostics"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/scanner"
)

// DiagnosticCategory classifies a diagnostic.
type DiagnosticCategory string

const (
	DiagnosticCategoryError      DiagnosticCategory = "error"
	DiagnosticCategoryWarning    DiagnosticCategory = "warning"
	DiagnosticCategorySuggestion DiagnosticCategory = "suggestion"
	DiagnosticCategoryMessage    DiagnosticCategory = "message"
)

// Diagnostic is a compiler message (error, warning, suggestion) optionally
// attached to a position in a source file.
type Diagnostic struct {
	// Message is the formatted message text.
	Message string
	// Code is the TypeScript error code, e.g. 2322.
	Code int32
	// Category classifies the diagnostic.
	Category DiagnosticCategory

	file  *SourceFile
	start int
	end   int
}

// SourceFile returns the file the diagnostic applies to, or false for
// project-level diagnostics.
func (d Diagnostic) SourceFile() (*SourceFile, bool) {
	return d.file, d.file != nil
}

// Start returns the byte offset of the diagnostic in its file, or -1.
func (d Diagnostic) Start() int { return d.start }

// End returns the end byte offset of the diagnostic in its file, or -1.
func (d Diagnostic) End() int { return d.end }

// LineAndColumn returns the 1-based line and column of the diagnostic, or
// false for project-level diagnostics.
func (d Diagnostic) LineAndColumn() (line, column int, ok bool) {
	if d.file == nil {
		return 0, 0, false
	}
	l, off := scanner.GetECMALineAndByteOffsetOfPosition(d.file.astFile(), d.start)
	return l + 1, off + 1, true
}

func categoryOf(c diagnostics.Category) DiagnosticCategory {
	switch c {
	case diagnostics.CategoryError:
		return DiagnosticCategoryError
	case diagnostics.CategoryWarning:
		return DiagnosticCategoryWarning
	case diagnostics.CategorySuggestion:
		return DiagnosticCategorySuggestion
	default:
		return DiagnosticCategoryMessage
	}
}

// PreEmitDiagnostics returns all syntactic, semantic, and project-level
// diagnostics for the project.
func (p *Project) PreEmitDiagnostics() []Diagnostic {
	program := p.getProgram()
	ctx := context.Background()

	var out []Diagnostic
	add := func(ds []*ast.Diagnostic) {
		for _, d := range ds {
			diag := Diagnostic{
				Message:  d.String(),
				Code:     d.Code(),
				Category: categoryOf(d.Category()),
				start:    -1,
				end:      -1,
			}
			if f := d.File(); f != nil && !isLibFile(f.FileName()) {
				diag.file = &SourceFile{project: p, path: f.FileName()}
				diag.start = d.Pos()
				diag.end = d.End()
			}
			out = append(out, diag)
		}
	}

	add(program.GetConfigFileParsingDiagnostics())
	add(program.GetProgramDiagnostics())
	add(program.GetSyntacticDiagnostics(ctx, nil))
	add(program.GetSemanticDiagnostics(ctx, nil))
	return out
}
