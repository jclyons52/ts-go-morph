package tsmorph

import (
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/printer"
)

// Print renders the node's subtree using the compiler's pretty-printer.
// The output is normalized (it will not preserve the original formatting)
// and is intended for debugging and inspection. Experimental: output details
// depend on the vendored compiler version.
func (n Node) Print() string {
	n.check()
	p := printer.NewPrinter(printer.PrinterOptions{}, printer.PrintHandlers{}, printer.NewEmitContext())
	return p.Emit(n.node, n.sf.astFile())
}

// Print renders the file using the compiler's pretty-printer. The output is
// normalized and will not preserve the original formatting. Experimental.
func (s *SourceFile) Print() string {
	p := printer.NewPrinter(printer.PrinterOptions{}, printer.PrintHandlers{}, printer.NewEmitContext())
	return p.EmitSourceFile(s.astFile())
}
