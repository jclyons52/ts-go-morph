package tsmorph

import "github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tspath"

// tspathGetBaseName is a small indirection so callers don't need to import
// tspath for one helper.
func tspathGetBaseName(path string) string { return tspath.GetBaseFileName(path) }
