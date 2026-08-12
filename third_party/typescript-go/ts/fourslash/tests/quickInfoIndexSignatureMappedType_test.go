package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestQuickInfoIndexSignatureMappedType(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	// Regression test for https://github.com/jclyons52/ts-go-morph/third_party/typescript-go/issues/3018
	// Quick info for property access resolved from an index signature on a mapped type
	// (e.g. Record<string, string>) should show the value type rather than nothing.
	const content = `
// @strict: true
// @filename: main.ts
declare const record: Record<string, string>;
record.fo/*1*/o;
`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyQuickInfoAt(t, "1", "string", "")
}
