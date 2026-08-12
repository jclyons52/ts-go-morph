package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	. "github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash/tests/util"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/lsp/lsproto"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestPathCompletionsPackageJsonExportsCustomConditions(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @module: node18
// @customConditions: custom-condition
// @Filename: /node_modules/foo/package.json
{
  "name": "foo",
  "exports": {
    "./only-with-custom-conditions": {
      "custom-condition": "./something.js"
    }
  }
}
// @Filename: /node_modules/foo/something.d.ts
export const index = 0;
// @Filename: /index.ts
import { } from "foo//**/";`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyCompletions(t, "", &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &[]string{},
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Exact: []fourslash.CompletionsExpectedItem{
				&lsproto.CompletionItem{
					Label:  "only-with-custom-conditions",
					Kind:   new(lsproto.CompletionItemKindFile),
					Detail: new("only-with-custom-conditions.js"),
				},
			},
		},
	})
}
