package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	. "github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash/tests/util"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/lsp/lsproto"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestClassMembersAfterConstAssertionInitializer(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `
interface A {
	a: number
	def: string
}

class B implements A {
	a = 1 as const
	/**/
}
`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()

	f.VerifyCompletions(t, "", &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &[]string{},
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Exact: append([]fourslash.CompletionsExpectedItem{
				&lsproto.CompletionItem{
					Label:    "def",
					Kind:     new(lsproto.CompletionItemKindField),
					SortText: new(string(ls.SortTextLocationPriority)),
				},
			}, CompletionClassElementKeywords...),
		},
	})
}
