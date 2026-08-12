# ts-go-morph: A Go port of ts-morph on top of typescript-go

## 1. Viability Assessment

### 1.1 What we are building

A Go library (`ts-go-morph`) that ports the developer experience of
[ts-morph](https://github.com/dsherret/ts-morph) — a friendly wrapper around the
TypeScript compiler API — but built on
[microsoft/typescript-go](https://github.com/microsoft/typescript-go), the
native Go port of the TypeScript compiler ("TypeScript 7", shipped as `tsgo`).

### 1.2 Key findings

**typescript-go status (as of research date):**

| Capability | Status | Relevance |
|---|---|---|
| Parsing/scanning | done | Exact same syntax errors as TS 6.0 — solid foundation |
| Program creation / module resolution | done | Needed for `Project` |
| Type resolution / type checking | done | Needed for type queries |
| `tsconfig.json` parsing | done | Needed for `Project` from tsconfig |
| Printer (emit) | done | Needed for writing source back |
| Public API | **not ready** | Main risk (see below) |

Module path is `github.com/microsoft/typescript-go`, Go >= 1.26 required.

**Relevant typescript-go packages (all under `internal/`):**

- `internal/ast` — AST node types (`ast.Node`, `ast.SourceFile`, kind enums)
- `internal/parser` / `internal/scanner` — parse source text into AST
- `internal/binder`, `internal/checker` — symbols, types, type checker
- `internal/compiler` — `Program` (module resolution + check orchestration)
- `internal/tsoptions` — `tsconfig.json` parsing, `CompilerOptions`
- `internal/printer` — emit AST back to text
- `internal/core`, `internal/vfs`, `internal/tspath` — file system abstraction,
  paths, utilities
- `internal/astnav` — AST navigation helpers
- `internal/ls`, `internal/format` — language service / formatting (in progress)

### 1.3 The critical constraint: `internal/` packages

Go forbids importing `internal/` packages from outside the module that owns
them. `github.com/microsoft/typescript-go` exposes **no public Go packages**,
and its "API" is explicitly marked *not ready*. We therefore cannot simply
`go get github.com/microsoft/typescript-go`.

**Viable workarounds (decision required in Step 0):**

1. **Vendor with import-path rewriting (recommended).** Copy the typescript-go
   source tree into this repo (e.g. `third_party/typescript-go`) and rewrite
   import paths from `github.com/microsoft/typescript-go/internal/...` to
   `<our-module>/third_party/typescript-go/internal/...`. Because the packages
   then live inside *our* module, the `internal` restriction is satisfied.
   Provide an `update.sh` script that re-fetches a pinned commit and re-applies
   the rewrite. Pros: pinned, reproducible, no fork to maintain. Cons: large
   vendored tree; we own upgrades.
2. **Fork typescript-go** under our org with rewritten module path, depend on
   the fork. Pros: smaller repo. Cons: fork maintenance, harder for an agent
   to bootstrap locally.
3. **Wait for the official API.** Not viable — timeline unknown.

This plan assumes **option 1**.

### 1.4 Mapping ts-morph concepts to Go

| ts-morph (TS) | ts-go-morph (Go) |
|---|---|
| `new Project({ tsConfigFilePath })` | `tsmorph.NewProject(tsmorph.ProjectOptions{...})` |
| `project.addSourceFileAtPath()` | `project.AddSourceFileAtPath(path)` |
| `SourceFile.getClasses()` etc. | `sf.Classes()` etc. (Go has no getters; methods) |
| Node wrapper hierarchy (`ClassDeclaration`, ...) | Go interfaces + structs wrapping `*ast.Node` |
| `node.getType()` | `node.Type()` using `checker.Checker` |
| In-memory edits, `project.save()` | Text-splice edits + re-parse, then write via vfs |
| ts-morph's `Structure` objects | Go struct literals (works naturally) |

**Idiomatic deviations from ts-morph (intentional):**

- No exceptions — return `(T, error)` or panic only on programmer error.
- No mutable global state; `Project` owns everything.
- Node wrappers are thin value types wrapping `*ast.Node` + back-pointer to
  their `SourceFile`/`Project` context.
- Interface-based node types (`Node` interface with `Kind()`,
  `Text()`, `Children()`), with concrete typed wrappers
  (`ClassDeclaration`, `FunctionDeclaration`, ...) that embed `Node`.

### 1.5 Biggest risks and mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| typescript-go `internal` APIs churn between commits | High | Pin a commit hash; upgrade deliberately via script; keep wrapper layer thin |
| typescript-go has **no incremental rebind** public surface; ts-morph relies on TS's incremental parsing | High | Adopt ts-morph's *observable* behaviour, not its mechanism: apply edits as text splices, re-parse the affected file with `parser`, and rebuild wrapper caches. Re-parsing one file in Go is fast enough for interactive use |
| Binder/checker require a full `Program`; ad-hoc type queries on edited-but-unsaved files may need a custom `vfs` | Medium | Use `internal/vfs` with an in-memory overlay so the `Program` always sees current text |
| Printer output ≠ original formatting | Medium | For MVP, edits are text splices (no re-print of untouched code). Re-printing whole files is opt-in |
| typescript-go repo may be merged into microsoft/TypeScript and archived | Low | Vendoring insulates us; update script can be repointed |

### 1.6 Verdict

**Viable.** Parsing, checking, tsconfig handling, and printing in
typescript-go are all marked "done". The only real obstacle is the `internal/`
visibility, solved by vendoring. The port should target ts-morph's *behaviour
and ergonomics*, implemented with text-splice editing + re-parse rather than
TS's incremental AST rebinding. MVP (project loading, navigation, common
manipulations, type queries, save) is achievable incrementally; full ts-morph
feature parity (code fixes, formatting, renaming across files) is a later
phase building on `internal/ls`.

---

## 2. Execution Plan

Each phase is independently completable and verifiable. Steps are written to
be executed by a coding agent in order. Do not skip verification steps.

**Conventions for the executing agent:**

- Module name used below: `github.com/jclyons52/ts-go-morph` (adjust if the
  repo is published elsewhere).
- Package name for the public API: `tsmorph`.
- Run `go build ./... && go vet ./... && go test ./...` after every phase.
- All public API lives at the repo root or in clearly named subpackages
  (`ast`-wrapping code in root package is fine for MVP).
- Go 1.26+ toolchain is required (match typescript-go's `go.mod`).

### Phase 0 — Repository & dependency bootstrap

1. Initialise the module:
   ```sh
   go mod init github.com/jclyons52/ts-go-morph
   go 1.26   # edit go.mod to match typescript-go's toolchain
   ```
2. Create `third_party/` and `tools/update-tsgo.sh`. The script must:
   - Clone/fetch `microsoft/typescript-go` at a commit hash stored in
     `third_party/TSGO_COMMIT`.
   - Copy the repo (minus `.git`, `_extension`, `_packages`, `_submodules`,
     `testdata`) into `third_party/typescript-go/`.
   - Rewrite all `github.com/microsoft/typescript-go/` import strings to
     `github.com/jclyons52/ts-go-morph/third_party/typescript-go/` across the
     copied tree.
   - Copy its `require` entries into our `go.mod` (document this as a manual
     verification step the first time).
3. Run the script, then:
   ```sh
   go mod tidy
   go build ./third_party/...
   ```
   The whole vendored tree must compile before proceeding. Pin and commit the
   chosen commit hash in `third_party/TSGO_COMMIT`.
4. Write a smoke test `internal_test.go` (temporary): parse
   `const x: number = 1;` with
   `third_party/typescript-go/internal/parser`, assert the resulting
   `*ast.SourceFile` has one `VariableStatement`. This proves the vendoring +
   rewrite works end to end.

**Exit criteria:** `go build ./...` passes; smoke test green.

### Phase 1 — Core project model (read-only)

Goal: load a project from a `tsconfig.json` (or explicit file list) and expose
source files. No editing yet.

1. Define `ProjectOptions`:
   ```go
   type ProjectOptions struct {
       TsConfigFilePath string   // optional
       RootFilePaths    []string // optional, alternative to tsconfig
       CompilerOptions  *core.CompilerOptions // optional overrides
       UseInMemoryFileSystem bool
   }
   ```
2. Implement `Project`:
   - Internally builds a `vfs.FS` (OS-backed, or in-memory overlay when
     `UseInMemoryFileSystem`).
   - Uses `tsoptions` to load/parse the tsconfig when `TsConfigFilePath` is
     set (mirror what `execute`/`compiler` do to build a `Program`).
   - Lazily creates a `compiler.Program` and caches it; expose
     `Program()`-level access only internally for now.
3. Implement `SourceFile` wrapper:
   - Holds `*ast.SourceFile`, file path, owning `*Project`.
   - `FilePath()`, `Text()`, `FullText()` (with trivia).
4. Implement `Project` methods:
   - `SourceFiles() []*SourceFile` (exclude lib files by default; option to
     include).
   - `SourceFile(path string) *SourceFile` (nil if absent).
   - `AddSourceFileAtPath(path string) (*SourceFile, error)` — reads from FS,
     parses, registers; invalidates the cached `Program`.
   - `CreateSourceFile(path, text string) *SourceFile` — in-memory only until
     saved.
5. Tests:
   - Fixture: a small tsconfig project with 2–3 files, an import graph, and
     one syntax-error-free file.
   - Assert file enumeration, path lookups, and text round-tripping.

**Exit criteria:** tests green; `Project` loads a real tsconfig project.

### Phase 2 — Node wrapper layer & navigation

Goal: ts-morph-style navigation over a source file. Read-only.

1. Define the base wrapper:
   ```go
   type Node struct {
       node *ast.Node
       sf   *SourceFile
   }
   func (n Node) Kind() ast.Kind
   func (n Node) Text() string
   func (n Node) Pos(), End() int          // byte offsets
   func (n Node) LineAndColumn() (line, col int) // via ast line map
   func (n Node) Parent() *Node
   func (n Node) Children() []Node
   func (n Node) Descendants() []Node      // depth-first
   ```
2. Implement typed wrappers as thin structs embedding `Node`, each with a
   `AsXxx()` downcast on `Node` and typed accessors. MVP set:
   - `ClassDeclaration` (Name, HeritageClauses, Methods, Properties, Ctors)
   - `InterfaceDeclaration` (Name, Members, Extends)
   - `FunctionDeclaration` (Name, Parameters, ReturnTypeNode, Body)
   - `MethodDeclaration`, `PropertyDeclaration`, `ParameterDeclaration`
   - `EnumDeclaration`, `EnumMember`
   - `TypeAliasDeclaration`
   - `VariableStatement` / `VariableDeclaration`
   - `ImportDeclaration`, `ExportDeclaration`, `ExportAssignment`
   - `SourceFile`-level: `ImportDeclarations()`, `Classes()`, `Interfaces()`,
     `Functions()`, `Enums()`, `TypeAliases()`, `VariableStatements()`
3. Navigation utilities mirroring ts-morph:
   - `sf.DescendantsOfKind(kind)` / `FirstDescendantByKind(kind)`.
   - `Node.FirstAncestorByKind(kind)`.
4. Tests: fixture file exercising every wrapper above; assert names, counts,
   positions, and `Text()` contents byte-for-byte against the fixture.

**Exit criteria:** navigation test suite green.

### Phase 3 — Type access (read-only semantics)

Goal: answer type questions like ts-morph's `node.getType()`.

1. Add `Project.Checker()` (internal): build the `compiler.Program` on demand
   and return its `checker.Checker`. Document that **any edit invalidates the
   program**; the checker is rebuilt lazily on next access.
2. Implement:
   - `Node.Type() *Type` — wraps `checker.Type`; must map the node's
     `*ast.Node` through the checker (`TypeOfNode`-equivalent).
   - `Type` wrapper (MVP): `Text()` (via printer/type-formatting), `IsString()`
     /`IsNumber()`/`IsBoolean()`/`IsAny()`/`IsUnknown()`/`IsUnion()`/
     `IsInterface()`, `UnionTypes()`, `Properties()` (symbol name + type),
     `CallSignatures()` (basic).
   - `Symbol` wrapper (MVP): `Name()`, `Declarations() []Node`,
     `Type() *Type`.
   - `Node.Symbol() *Symbol`.
3. Diagnostics:
   - `Project.PreEmitDiagnostics() []Diagnostic` (message, file, start,
     line/col, category, code).
4. Tests: fixture with a typed class, a function with inferred return type, a
   union type alias, and one deliberate type error. Assert `Type().Text()`,
   flags, properties, and that diagnostics contain the expected error code.

**Exit criteria:** type/diagnostic tests green; checker rebuild after
re-parse verified by test.

### Phase 4 — Manipulation engine (the heart of the port)

Goal: ts-morph-style edits with the observable guarantee that all previously
obtained wrappers for the edited file are invalidated and re-fetched wrappers
reflect the new text.

Design (decided up front — do not redesign mid-implementation):

- **Edits are text splices.** A `SourceFile` accumulates ordered, non-
  overlapping edits `replace(start, end, newText)` in a transaction.
- `SourceFile.ApplyEdits()` (called automatically at the end of every public
  mutating method, matching ts-morph's eager model) applies splices to the
  in-memory text, **re-parses the file** with `parser`, swaps the
  `*ast.SourceFile`, invalidates the project's cached `Program`, and marks
  all previously handed-out `Node` wrappers for that file as forgotten
  (`Node.IsForgotten() == true`; any method call on a forgotten node panics
  with a clear message, same contract as ts-morph).
- Offset bookkeeping: after each splice, adjust the remaining queued edits by
  the delta.

1. Implement the splice engine (`manipulation.go`):
   - `type textEdit struct{ start, end int; newText string }`
   - Validation: sorted, non-overlapping, within bounds.
   - `applyTextEdits(text string, edits []textEdit) string`.
2. Implement forget/invalidation:
   - Generation counter on `SourceFile`; each `Node` stores the generation it
     was created in. Mismatch → forgotten.
3. Implement insertion helpers used by all mutating APIs:
   - `insertIntoBracedNode(parent Node, index int, text string)` — compute
     insertion point inside `{ ... }`, honouring `Project.ManipulationSettings`
       (indentation, newline kind).
   - `insertIntoParenthesized(...)` for parameter lists.
   - Comma handling: inserting into a non-empty list appends `, ` or prepends
     per index.
4. Implement `ManipulationSettings` on `ProjectOptions`:
   `IndentationText` (default two spaces), `NewLineKind` (`\n` default),
   `QuoteKind` (default double), `UseTrailingCommas`.
5. Mutating APIs (MVP, each with tests):
   - `sf.InsertClass(index, ClassStructure)`, `sf.AddClass(...)`,
     `sf.AddInterface`, `sf.AddFunction`, `sf.AddEnum`, `sf.AddTypeAlias`,
     `sf.AddVariableStatement`
   - `ClassDeclaration.AddMethod(MethodStructure)`, `AddProperty`,
     `AddConstructor`; `InterfaceDeclaration.AddMethod`, `AddProperty`
   - `FunctionDeclaration.AddParameter(ParameterStructure)`,
     `SetReturnType(text)`
   - `sf.AddImportDeclaration(ImportStructure)` (named/default/namespace
     imports), `sf.AddExportDeclaration`
   - `Node.ReplaceWithText(text)`, `Node.Remove()`
   - `ClassDeclaration.Rename(newName)`, `FunctionDeclaration.Rename`,
     `VariableDeclaration.Rename` — MVP: rename the identifier token only
     (not cross-file references; document this limitation).
   - `Structure` types: plain Go structs, e.g.
     ```go
     type ClassStructure struct {
         Name       string
         IsExported bool
         Extends    string
         Implements []string
         Methods    []MethodStructure
         Properties []PropertyStructure
     }
     ```
     with a `codeWriter` that renders structures to text honouring
     `ManipulationSettings`.
6. Persistence:
   - `SourceFile.Save() error`, `Project.Save() error` — write in-memory text
     through the vfs to disk. No-op for files with no pending edits.
7. Tests (this phase needs the most):
   - Golden-file tests: `testdata/manipulation/<case>/before.ts` +
     operations script + `after.ts`. Write a small test harness that applies
     each operation and compares output byte-for-byte.
   - Forget semantics: hold a wrapper, mutate the file, assert
     `IsForgotten()` and panic-on-use.
   - Idempotency: apply, save, reload project, assert same text.

**Exit criteria:** full manipulation test suite green; gofmt-clean; no
regressions in Phases 1–3 tests.

### Phase 5 — Ergonomics & utilities

1. `Project.CreateDirectory`/`FileSystem` helpers (only if in-memory FS is
   on; keep minimal).
2. `sf.FormatText()` via `internal/format` if stable at the pinned commit;
   otherwise skip and document.
3. `sf.FixMissingImports()`-style helpers — defer; requires `internal/ls`
   maturity. Track as stretch goal.
4. `Node.Print()` — print subtree via `internal/printer` (useful for
   debugging; mark experimental).
5. Convenience predicates on `Node`: `IsClassDeclaration()`, etc., matching
   the wrapper set from Phase 2.
6. Doc comments on every exported symbol; `doc.go` with a quick-start example
   mirroring ts-morph's README.

**Exit criteria:** `go vet` clean, docs complete, example in `doc.go`
compiles (as an `Example` test).

### Phase 6 — Hardening & release prep

1. Concurrency audit: document that `Project` is not goroutine-safe; add a
   mutex around program/checker rebuild if cheap.
2. Performance sanity: benchmark parse+navigate and edit+reparse on a
   ~5k-line file; re-parse must be well under 100 ms on commodity hardware.
3. Upgrade drill: run `tools/update-tsgo.sh` against a newer typescript-go
   commit, fix breakage, document the process in the script header.
4. CI: GitHub Actions running `go build`, `go vet`, `go test` on Linux +
   macOS, Go 1.26.
5. README with: viability notes from §1, quick start, feature matrix vs
   ts-morph, known limitations (no cross-file rename, no incremental binder,
   vendored compiler), and upgrade instructions.
6. Tag `v0.1.0` only after all the above are green.

---

## 3. Out of scope for v1 (record explicitly so the agent does not drift)

- Cross-file rename / find-references (needs `internal/ls` maturity).
- Language-service features: code fixes, refactors, completions.
- Emitting compiled JS (typescript-go can, but it's not ts-morph's purpose).
- Watching the file system / incremental builder integration.
- API parity with every ts-morph node kind (TS has 300+ kinds; v1 covers the
  ~15 most-used declaration kinds from Phase 2/4).

## 4. Definition of done per step

- Code compiles: `go build ./...`
- Lints clean: `go vet ./...` (and `gofmt -l .` empty)
- Tests pass: `go test ./...`
- New public API has doc comments
- Golden-file fixtures added for every new manipulation
