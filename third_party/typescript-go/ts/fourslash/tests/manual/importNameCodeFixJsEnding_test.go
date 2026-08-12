package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/core"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ls/lsutil"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestImportNameCodeFixJsEnding(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @lib: es5
// @module: commonjs
// @Filename: /node_modules/lit/package.json
{ "name": "lit", "version": "1.0.0" }
// @Filename: /node_modules/lit/index.d.ts
import "./decorators";
// @Filename: /node_modules/lit/decorators.d.ts
export declare function customElement(name: string): any;
// @Filename: /a.ts
customElement/**/`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyImportFixModuleSpecifiers(t, "", []string{"lit/decorators.js"}, &lsutil.UserPreferences{ImportModuleSpecifierEnding: "js", AutoImportEntrypointDirectorySearch: core.TSTrue})
}
