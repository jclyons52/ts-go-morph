package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestCodeFixMissingTypeAnnotationOnExports_arrowParensParamOnly(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @isolatedDeclarations: true
// @declaration: true
export const func = /*a*/x/*b*/ => 0;`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.GoToMarker(t, "a")
	f.VerifyCodeFix(t, fourslash.VerifyCodeFixOptions{
		Description:    `Add annotation of type 'any'`,
		NewFileContent: `export const func = (x: any) => 0;`,
		ApplyChanges:   true,
	})
}
