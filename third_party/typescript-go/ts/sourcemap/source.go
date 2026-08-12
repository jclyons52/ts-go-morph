package sourcemap

import "github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/core"

type Source interface {
	Text() string
	FileName() string
	ECMALineMap() []core.TextPos
}
