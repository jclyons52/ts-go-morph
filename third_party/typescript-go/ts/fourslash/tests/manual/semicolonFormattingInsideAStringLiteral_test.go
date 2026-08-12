package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestSemicolonFormattingInsideAStringLiteral(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `    var x = "string/**/`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.GoToMarker(t, "")
	f.Insert(t, ";")
	f.VerifyCurrentLineContent(t, `   var x = "string;`)
}
