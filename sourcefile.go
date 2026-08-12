package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/scanner"
)

// SourceFile is a TypeScript source file in a Project.
type SourceFile struct {
	project *Project
	file    *ast.SourceFile
}

// astFile returns the underlying AST. Internal use only.
func (s *SourceFile) astFile() *ast.SourceFile { return s.file }

// FilePath returns the normalized absolute path of the file.
func (s *SourceFile) FilePath() string { return s.file.FileName() }

// BaseName returns the file name without its directory, e.g. "index.ts".
func (s *SourceFile) BaseName() string { return tspathGetBaseName(s.file.FileName()) }

// Text returns the full text of the file.
func (s *SourceFile) Text() string { return s.file.Text() }

// LineAndColumn returns the 1-based line and byte-offset column of pos,
// which must be a byte offset within the file text.
func (s *SourceFile) LineAndColumn(pos int) (line, column int) {
	l, off := scanner.GetECMALineAndByteOffsetOfPosition(s.file, pos)
	return l + 1, off + 1
}
