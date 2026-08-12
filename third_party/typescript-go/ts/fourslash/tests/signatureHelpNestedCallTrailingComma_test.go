package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/lsp/lsproto"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestSignatureHelpNestedCallTrailingComma(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	// Regression test for crash when requesting signature help on a call target
	// where the nested call has a trailing comma.
	// Both outer and inner calls must have trailing commas, and outer must be generic.
	const content = `declare function outer<T>(range: T): T;
declare function inner(a: any): any;

outer(inner/*1*/(undefined,),);`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.GoToMarker(t, "1")
	f.VerifySignatureHelpPresent(t, &lsproto.SignatureHelpContext{
		IsRetrigger: false,
		TriggerKind: lsproto.SignatureHelpTriggerKindInvoked,
	})
}
