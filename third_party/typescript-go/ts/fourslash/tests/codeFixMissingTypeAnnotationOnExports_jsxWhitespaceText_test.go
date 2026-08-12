package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestCodeFixMissingTypeAnnotationOnExports_jsxWhitespaceText(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")

	const content = `// @isolatedDeclarations: true
// @declaration: true
// @module: preserve
// @Filename: /test.tsx
export const /**/elem = <div>
    <span />
</div>;`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()

	f.GoToMarker(t, "")
	f.VerifyCodeFixAvailable(t, nil)
}
