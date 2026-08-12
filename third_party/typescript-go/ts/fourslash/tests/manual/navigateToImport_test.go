package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/lsp/lsproto"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestNavigateToImport(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @lib: es5
// @Filename: library.ts
export function [|foo|]() {}
export function [|bar|]() {}
// @Filename: user.ts
import {foo, bar as [|baz|]} from './library';`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyWorkspaceSymbol(t, []*fourslash.VerifyWorkspaceSymbolCase{
		{
			Pattern:     "foo",
			Preferences: nil,
			Exact: new([]*lsproto.SymbolInformation{
				{
					Name:     "foo",
					Kind:     lsproto.SymbolKindFunction,
					Location: f.Ranges()[0].LSLocation(),
				},
			}),
		}, {
			Pattern:     "bar",
			Preferences: nil,
			Exact: new([]*lsproto.SymbolInformation{
				{
					Name:     "bar",
					Kind:     lsproto.SymbolKindFunction,
					Location: f.Ranges()[1].LSLocation(),
				},
			}),
		}, {
			Pattern:     "baz",
			Preferences: nil,
			Exact: new([]*lsproto.SymbolInformation{
				{
					Name:     "baz",
					Kind:     lsproto.SymbolKindVariable,
					Location: f.Ranges()[2].LSLocation(),
				},
			}),
		},
	})
}
