package fixtures

import (
	"path/filepath"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/repo"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/testutil/filefixture"
)

var BenchFixtures = []filefixture.Fixture{
	filefixture.FromString("empty.ts", "empty.ts", ""),
	filefixture.FromFile("checker.ts", filepath.Join(repo.TypeScriptSubmodulePath(), "src/compiler/checker.ts")),
	filefixture.FromFile("dom.generated.d.ts", filepath.Join(repo.TypeScriptSubmodulePath(), "src/lib/dom.generated.d.ts")),
	filefixture.FromFile("Herebyfile.mjs", filepath.Join(repo.TypeScriptSubmodulePath(), "Herebyfile.mjs")),
	filefixture.FromFile("jsxComplexSignatureHasApplicabilityError.tsx", filepath.Join(repo.TypeScriptSubmodulePath(), "tests/cases/compiler/jsxComplexSignatureHasApplicabilityError.tsx")),
}
