package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestGetEditsForFileRename_unaffectedNonRelativePath(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: /sub/a.ts
export const a = 1;
// @Filename: /sub/b.ts
import { a } from "sub/a";
// @Filename: /tsconfig.json
{"compilerOptions":{"paths":{"*":["*"]}}}`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyWillRenameFilesEdits(t, "/sub/b.ts", "/sub/c/d.ts", map[string]string{}, nil /*preferences*/)
}
