package tsmorph

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFixtureProject creates a small tsconfig project in a temp dir:
//
//	tsconfig.json
//	src/index.ts   (imports ./util)
//	src/util.ts
//	src/unused.ts  (not imported, still in the program via include)
func writeFixtureProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "strict": true, "target": "es2020", "module": "commonjs" },
  "include": ["src"]
}
`,
		"src/index.ts": `import { greet } from "./util";

export const message: string = greet("world");
`,
		"src/util.ts": `export function greet(name: string): string {
  return "hello " + name;
}
`,
		"src/unused.ts": `export const unused = 42;
`,
	}
	for name, text := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestNewProjectFromTsConfig(t *testing.T) {
	dir := writeFixtureProject(t)

	p, err := NewProject(ProjectOptions{TsConfigFilePath: filepath.Join(dir, "tsconfig.json")})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}

	files := p.SourceFiles()
	if len(files) != 3 {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.FilePath()
		}
		t.Fatalf("expected 3 source files, got %d: %v", len(files), names)
	}

	sf := p.SourceFile(filepath.Join(dir, "src", "util.ts"))
	if sf == nil {
		t.Fatal("SourceFile(util.ts) returned nil")
	}
	if got := sf.BaseName(); got != "util.ts" {
		t.Fatalf("BaseName: got %q", got)
	}
	wantText := "export function greet(name: string): string {\n  return \"hello \" + name;\n}\n"
	if got := sf.Text(); got != wantText {
		t.Fatalf("Text round-trip failed:\ngot  %q\nwant %q", got, wantText)
	}

	if missing := p.SourceFile(filepath.Join(dir, "src", "missing.ts")); missing != nil {
		t.Fatal("expected nil for missing file")
	}
}

func TestProjectAddAndCreateSourceFile(t *testing.T) {
	dir := writeFixtureProject(t)

	p, err := NewProject(ProjectOptions{TsConfigFilePath: filepath.Join(dir, "tsconfig.json")})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}

	// AddSourceFileAtPath on a file outside the tsconfig include.
	extra := filepath.Join(dir, "extra.ts")
	if err := os.WriteFile(extra, []byte("export const extra = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sf, err := p.AddSourceFileAtPath(extra)
	if err != nil {
		t.Fatalf("AddSourceFileAtPath: %v", err)
	}
	if sf.Text() != "export const extra = 1;\n" {
		t.Fatalf("unexpected text: %q", sf.Text())
	}

	// CreateSourceFile is in-memory until Save.
	mem := p.CreateSourceFile(filepath.Join(dir, "created.ts"), "export const created = true;\n")
	if mem == nil {
		t.Fatal("CreateSourceFile returned nil")
	}
	if _, err := os.Stat(filepath.Join(dir, "created.ts")); !os.IsNotExist(err) {
		t.Fatal("created.ts should not exist on disk before Save")
	}

	if err := p.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(dir, "created.ts"))
	if err != nil {
		t.Fatalf("created.ts missing after Save: %v", err)
	}
	if string(contents) != "export const created = true;\n" {
		t.Fatalf("unexpected saved text: %q", contents)
	}

	// After save + invalidation, the file is still part of the project.
	if p.SourceFile(filepath.Join(dir, "created.ts")) == nil {
		t.Fatal("created.ts not in project after Save")
	}
}

func TestNewProjectWithRootFiles(t *testing.T) {
	dir := writeFixtureProject(t)

	p, err := NewProject(ProjectOptions{
		RootFilePaths: []string{filepath.Join(dir, "src", "index.ts")},
	})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	// index.ts and its import (util.ts) are reachable; unused.ts is not.
	if got := len(p.SourceFiles()); got != 2 {
		t.Fatalf("expected 2 source files, got %d", got)
	}
}

func TestInMemoryProject(t *testing.T) {
	p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	sf := p.CreateSourceFile("/src/main.ts", "export const x: number = 1;\n")
	if sf == nil {
		t.Fatal("CreateSourceFile returned nil")
	}
	if sf.FilePath() != "/src/main.ts" {
		t.Fatalf("FilePath: got %q", sf.FilePath())
	}
	if err := p.Save(); err != nil {
		t.Fatalf("Save (in-memory no-op): %v", err)
	}
}
