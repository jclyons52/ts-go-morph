package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestNavigationBarFunctionPrototype4(t *testing.T) {
	t.Parallel()

	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @allowJs: true
// @Filename: foo.js
var A; 
A.prototype = { };
A.prototype = { m() {} };
A.prototype.a = function() { };
A.b = function() { };

var B; 
B["prototype"] = { };
B["prototype"] = { m() {} };
B["prototype"]["a"] = function() { };
B["b"] = function() { };`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyBaselineDocumentSymbol(t)
}
