package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestGetEditsForFileRename_subDir(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: /src/foo/a.ts

// @Filename: /src/old.ts
import a from "./foo/a";`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyWillRenameFilesEdits(t, "/src/old.ts", "/src/dir/new.ts", map[string]string{
		"/src/dir/new.ts": `import a from "../foo/a";`,
	}, nil /*preferences*/)
}
