package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestHoverCallSignatureDocumentation(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `
type X = {
    /** Description of invoking. */
    (): string

    /** Description of constructor. */
    new (): number
}

declare const x: X

/*1*/x()
new /*2*/x()
`

	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()

	f.VerifyQuickInfoAt(t, "1", "const x: () => string", "Description of invoking.")
	f.VerifyQuickInfoAt(t, "2", "const x: new () => number", "Description of constructor.")
}
