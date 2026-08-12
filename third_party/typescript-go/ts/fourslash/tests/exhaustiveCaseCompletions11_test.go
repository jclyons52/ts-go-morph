package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	. "github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash/tests/util"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/lsp/lsproto"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestExhaustiveCaseCompletions11(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `
declare const u: "$1" | "2";
switch (u) {
    case/*1*/
}`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyCompletions(t, "1", &fourslash.CompletionsExpectedList{
		IsIncomplete: false,
		ItemDefaults: &fourslash.CompletionsExpectedItemDefaults{
			CommitCharacters: &DefaultCommitCharacters,
			EditRange:        Ignored,
		},
		Items: &fourslash.CompletionsExpectedItems{
			Includes: []fourslash.CompletionsExpectedItem{
				&lsproto.CompletionItem{
					Label:            "case \"$1\": ...",
					InsertText:       new("case \"\\$1\":$1\ncase \"2\":$2"),
					InsertTextFormat: new(lsproto.InsertTextFormatSnippet),
					SortText:         new(string(ls.SortTextGlobalsOrKeywords)),
				},
			},
		},
	})
}
