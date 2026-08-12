package tsmorph

// NewLineKind controls the newline used in generated code.
type NewLineKind string

const (
	NewLineKindLineFeed       NewLineKind = "\n"
	NewLineKindCarriageReturn NewLineKind = "\r\n"
)

// QuoteKind controls the quote character used in generated string literals
// and module specifiers.
type QuoteKind string

const (
	QuoteKindDouble QuoteKind = "\""
	QuoteKindSingle QuoteKind = "'"
)

// ManipulationSettings controls how generated code is formatted.
type ManipulationSettings struct {
	// IndentationText is one level of indentation. Default: two spaces.
	IndentationText string
	// NewLineKind is the newline used in generated code. Default: "\n".
	NewLineKind NewLineKind
	// QuoteKind is the quote character for generated module specifiers.
	// Default: double quotes.
	QuoteKind QuoteKind
	// UseTrailingCommas adds trailing commas in generated multi-line lists.
	// Default: false.
	UseTrailingCommas bool
}

// ManipulationSettings returns the project's settings with defaults applied.
func (p *Project) ManipulationSettings() ManipulationSettings {
	s := p.opts.ManipulationSettings
	if s.IndentationText == "" {
		s.IndentationText = "  "
	}
	if s.NewLineKind == "" {
		s.NewLineKind = NewLineKindLineFeed
	}
	if s.QuoteKind == "" {
		s.QuoteKind = QuoteKindDouble
	}
	return s
}
