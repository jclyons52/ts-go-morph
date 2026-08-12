// Package tsmorph is a Go port of ts-morph (https://github.com/dsherret/ts-morph)
// built on the native Go port of the TypeScript compiler
// (microsoft/typescript-go, vendored under third_party/typescript-go).
//
// It provides a simple wrapper around the compiler API for setting up,
// navigating, and manipulating TypeScript source files.
package tsmorph

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/bundled"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/compiler"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/core"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tsoptions"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tspath"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/vfs"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/vfs/osvfs"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/vfs/vfstest"
)

// ProjectOptions configures a Project.
type ProjectOptions struct {
	// TsConfigFilePath loads the project from a tsconfig.json. May be
	// relative to the current working directory.
	TsConfigFilePath string

	// RootFilePaths is an explicit list of root files, used when no tsconfig
	// is provided.
	RootFilePaths []string

	// CompilerOptions overrides the compiler options from the tsconfig, or
	// provides them when no tsconfig is used. May be nil.
	CompilerOptions *core.CompilerOptions

	// UseInMemoryFileSystem starts the project with an empty in-memory file
	// system instead of the real one. Paths should be absolute, e.g.
	// "/src/index.ts". Save() is a no-op in this mode.
	UseInMemoryFileSystem bool
}

// Project is a collection of source files and the compiler state binding
// them together. A Project is not safe for concurrent use.
type Project struct {
	opts ProjectOptions

	cwd  string
	fsys *overlayFS
	host compiler.CompilerHost

	// config is the parsed tsconfig (nil when RootFilePaths are used).
	config *tsoptions.ParsedCommandLine
	// extraRoots are root files added on top of the tsconfig's file list.
	extraRoots []string

	mu      sync.Mutex
	program *compiler.Program // lazily built; nil when invalidated
}

// NewProject creates a Project from the given options.
func NewProject(opts ProjectOptions) (*Project, error) {
	var base vfs.FS
	cwd := "/"
	if opts.UseInMemoryFileSystem {
		base = vfstest.FromMap(map[string]string{}, true)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("tsmorph: get working directory: %w", err)
		}
		cwd = tspath.NormalizePath(wd)
		base = osvfs.FS()
	}
	// WrapFS redirects embedded lib.*.d.ts paths to the embedded FS.
	fsys := newOverlayFS(bundled.WrapFS(base))

	p := &Project{opts: opts, cwd: cwd, fsys: fsys}
	p.host = compiler.NewCompilerHost(cwd, fsys, bundled.LibPath(), nil, nil)

	if opts.TsConfigFilePath != "" {
		configPath := p.absPath(opts.TsConfigFilePath)
		config, errs := tsoptions.GetParsedCommandLineOfConfigFile(configPath, opts.CompilerOptions, nil, p.host, nil)
		if len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, e := range errs {
				msgs[i] = e.String()
			}
			return nil, fmt.Errorf("tsmorph: errors parsing %s:\n%s", configPath, strings.Join(msgs, "\n"))
		}
		p.config = config
	} else {
		roots := make([]string, len(opts.RootFilePaths))
		for i, r := range opts.RootFilePaths {
			roots[i] = p.absPath(r)
		}
		p.extraRoots = roots
	}

	return p, nil
}

// absPath converts path to a normalized absolute path against the project's
// working directory.
func (p *Project) absPath(path string) string {
	return tspath.GetNormalizedAbsolutePath(path, p.cwd)
}

// compilerOptions returns the effective compiler options.
func (p *Project) compilerOptions() *core.CompilerOptions {
	if p.config != nil {
		return p.config.CompilerOptions()
	}
	if p.opts.CompilerOptions != nil {
		return p.opts.CompilerOptions
	}
	return &core.CompilerOptions{}
}

// rootFileNames returns the tsconfig's files plus any added roots.
func (p *Project) rootFileNames() []string {
	var names []string
	if p.config != nil {
		names = p.config.FileNames()
	}
	return append(names, p.extraRoots...)
}

// comparePathsOptions builds path comparison options for this project.
func (p *Project) comparePathsOptions() tspath.ComparePathsOptions {
	return tspath.ComparePathsOptions{
		CurrentDirectory:          p.cwd,
		UseCaseSensitiveFileNames: p.fsys.UseCaseSensitiveFileNames(),
	}
}

// getProgram returns the current compiler.Program, building it on first use
// or after invalidation.
func (p *Project) getProgram() *compiler.Program {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.program != nil {
		return p.program
	}
	config := tsoptions.NewParsedCommandLine(p.compilerOptions(), p.rootFileNames(), p.comparePathsOptions())
	if p.config != nil {
		config.ConfigFile = p.config.ConfigFile
	}
	p.program = compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: p.host})
	p.program.BindSourceFiles()
	return p.program
}

// invalidate drops the cached program; it is rebuilt lazily on next access.
// All Node and SourceFile AST state from the old program is stale after this.
func (p *Project) invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.program = nil
}

// isLibFile reports whether fileName is a bundled lib.*.d.ts file.
func isLibFile(fileName string) bool {
	return bundled.IsBundled(fileName)
}

// SourceFiles returns all source files in the project, excluding bundled
// library files (lib.*.d.ts).
func (p *Project) SourceFiles() []*SourceFile {
	var result []*SourceFile
	for _, f := range p.getProgram().SourceFiles() {
		if isLibFile(f.FileName()) {
			continue
		}
		result = append(result, &SourceFile{project: p, file: f})
	}
	return result
}

// SourceFile returns the source file at path, or nil if it is not part of
// the project.
func (p *Project) SourceFile(path string) *SourceFile {
	abs := p.absPath(path)
	key := tspath.ToPath(abs, p.cwd, p.fsys.UseCaseSensitiveFileNames())
	if f, ok := p.getProgram().FilesByPath()[key]; ok && !isLibFile(f.FileName()) {
		return &SourceFile{project: p, file: f}
	}
	return nil
}

// AddSourceFileAtPath adds an existing file from the file system to the
// project as a root file.
func (p *Project) AddSourceFileAtPath(path string) (*SourceFile, error) {
	abs := p.absPath(path)
	if !p.fsys.FileExists(abs) {
		return nil, fmt.Errorf("tsmorph: file not found: %s", abs)
	}
	p.extraRoots = append(p.extraRoots, abs)
	p.invalidate()
	if sf := p.SourceFile(abs); sf != nil {
		return sf, nil
	}
	return nil, fmt.Errorf("tsmorph: %s was not added to the project (unsupported extension?)", abs)
}

// CreateSourceFile creates a source file with the given text. The file only
// exists in memory until Save is called.
func (p *Project) CreateSourceFile(path, text string) *SourceFile {
	abs := p.absPath(path)
	p.fsys.setOverlay(abs, text)
	p.extraRoots = append(p.extraRoots, abs)
	p.invalidate()
	return p.SourceFile(abs)
}

// Save writes all files with unsaved changes to disk. It is a no-op for
// in-memory projects.
func (p *Project) Save() error {
	if p.opts.UseInMemoryFileSystem {
		return nil
	}
	if err := p.fsys.flush(); err != nil {
		return fmt.Errorf("tsmorph: save: %w", err)
	}
	p.invalidate()
	return nil
}
