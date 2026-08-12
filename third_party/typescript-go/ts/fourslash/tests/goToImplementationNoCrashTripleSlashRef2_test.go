package fourslash_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/fourslash"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil"
)

func TestGoToImplementationNoCrashTripleSlashRef2(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "Panic on fourslash test")
	const content = `// @Filename: /node_modules/@types/react/index.d.ts
export type JSX = {};

// @Filename: /node_modules/excalidraw/index.d.ts
/// <reference types="react" />

// @Filename: /index.ts
import type {JSX} from '/*m*/react';
`
	f, done := fourslash.NewFourslash(t, nil /*capabilities*/, content)
	defer done()
	f.VerifyBaselineGoToImplementation(t, "m")
}
