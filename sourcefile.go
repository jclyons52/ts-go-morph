package tsmorph

import (
	"fmt"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/scanner"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tspath"
)

// SourceFile is a TypeScript source file in a Project. It is a lightweight
// handle identified by path: after edits, the same handle re-resolves to the
// freshly parsed AST.
type SourceFile struct {
	project *Project
	path    string // normalized absolute path
}

// astFile returns the current underlying AST. Internal use only.
func (s *SourceFile) astFile() *ast.SourceFile {
	key := tspath.ToPath(s.path, s.project.cwd, s.project.fsys.UseCaseSensitiveFileNames())
	return s.project.getProgram().FilesByPath()[key]
}

// canonicalPath returns the map key used for per-file state.
func (s *SourceFile) canonicalPath() string {
	return string(tspath.ToPath(s.path, s.project.cwd, s.project.fsys.UseCaseSensitiveFileNames()))
}

// generation returns the current generation of the file, bumped on every
// edit. Nodes created under an older generation are forgotten.
func (s *SourceFile) generation() int {
	return s.project.generationOf(s.canonicalPath())
}

// FilePath returns the normalized absolute path of the file.
func (s *SourceFile) FilePath() string { return s.path }

// BaseName returns the file name without its directory, e.g. "index.ts".
func (s *SourceFile) BaseName() string { return tspathGetBaseName(s.path) }

// Text returns the full text of the file, including unsaved changes.
func (s *SourceFile) Text() string { return s.astFile().Text() }

// LineAndColumn returns the 1-based line and byte-offset column of pos,
// which must be a byte offset within the file text.
func (s *SourceFile) LineAndColumn(pos int) (line, column int) {
	l, off := scanner.GetECMALineAndByteOffsetOfPosition(s.astFile(), pos)
	return l + 1, off + 1
}

// IsSaved reports whether the file has no unsaved changes.
func (s *SourceFile) IsSaved() bool {
	return !s.project.fsys.hasOverlay(s.path)
}

// Save writes the file's current text to disk if it has unsaved changes.
// It is a no-op for in-memory projects.
func (s *SourceFile) Save() error {
	if s.project.opts.UseInMemoryFileSystem {
		return nil
	}
	if _, err := s.project.fsys.flushFile(s.path); err != nil {
		return fmt.Errorf("tsmorph: save %s: %w", s.path, err)
	}
	s.project.invalidate()
	return nil
}
