# ts-go-morph

A Go port of [ts-morph](https://github.com/dsherret/ts-morph) built on
[microsoft/typescript-go](https://github.com/microsoft/typescript-go), the
native Go port of the TypeScript compiler ("TypeScript 7" / `tsgo`).

It wraps the compiler in a simple API for setting up, navigating, and
manipulating TypeScript source files.

```go
p, err := tsmorph.NewProject(tsmorph.ProjectOptions{
    TsConfigFilePath: "tsconfig.json",
})
if err != nil {
    log.Fatal(err)
}

sf := p.SourceFile("src/person.ts")
c, _ := sf.Class("Person")
fmt.Println(c.Name())          // navigation
fmt.Println(c.Type().Text())   // type checker access

c.AddMethod(tsmorph.MethodStructure{ // manipulation (applies immediately)
    Name:       "greet",
    ReturnType: "string",
    Body:       `return "hi " + this.name;`,
})

err = p.Save() // flush all edited files to disk
```

## How it works

typescript-go exposes only `internal/` packages, so its source is vendored
into `third_party/typescript-go` by `tools/update-tsgo.sh`, which renames
`internal/` to `ts/` and rewrites import paths. The vendored commit is pinned
in `third_party/TSGO_COMMIT`; bump it and re-run the script to upgrade.

Manipulation is implemented as text splices: each edit updates the file's
text in an in-memory overlay file system and re-parses the file, so the
compiler always sees current text without touching disk. After an edit, all
previously obtained `Node` wrappers for that file are *forgotten* — using one
panics (same contract as ts-morph). Re-fetch nodes from the `SourceFile`
after edits.

## Feature matrix vs ts-morph

| Area | Status |
|---|---|
| Project from tsconfig / explicit root files | ✅ |
| In-memory file system | ✅ |
| Navigation: classes, interfaces, functions, enums, type aliases, variables, imports, exports | ✅ |
| Tree traversal (children/descendants/ancestors, by kind) | ✅ |
| Type access (`Type`, `Symbol`, signatures, union types, properties) | ✅ |
| Pre-emit diagnostics | ✅ |
| Manipulation: add/insert declarations & members, parameters, return types, imports/exports, rename (declaration site), remove, replace text | ✅ |
| Pretty-printing (`Node.Print`) | ✅ experimental |
| Cross-file rename / find references | ❌ (needs typescript-go's language service) |
| Code fixes, refactors, completions, formatting | ❌ (same) |
| Every AST node kind (TS has 300+) | ❌ (~15 most-used declaration kinds wrapped) |

## Known limitations

- **No cross-file rename.** `Node.Rename` only changes the declaration-site
  identifier.
- **Edits re-parse the whole file** (and rebuild the program). This is fast
  enough in practice (~25 ms for a 5k-line file including type binding).
- **Nodes from other files survive an edit, nodes from the edited file do
  not.** `SourceFile` handles survive; re-fetch nodes after each edit.
- The vendored compiler is an in-progress port; its packages may change
  between commits. The wrapper layer is deliberately thin to make upgrades
  mechanical.
- `Project` is not safe for concurrent use.

## Requirements

Go 1.26+.

## Development

```sh
go build ./...
go vet .
go test .
```

Manipulation behaviour is covered by golden-file tests under
`testdata/manipulation/<case>/{before.ts,after.ts}`, driven by the
`manipulationOps` map in `manipulation_test.go`.

To upgrade the vendored compiler: put a new commit hash in
`third_party/TSGO_COMMIT`, run `tools/update-tsgo.sh`, then `go mod tidy`,
fix any API breakage, and run the tests.
