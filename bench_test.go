package tsmorph

import (
	"fmt"
	"strings"
	"testing"
)

// bigFixture generates a TypeScript file of roughly 7 lines per class.
func bigFixture(classes int) string {
	var sb strings.Builder
	for i := 0; i < classes; i++ {
		fmt.Fprintf(&sb, `export class K%d {
  propA: number = 1;
  propB: string = "s";
  method(x: number, y: string): boolean {
    return x > 0 && y.length > 0;
  }
}

`, i)
	}
	return sb.String()
}

// ~700 classes ≈ ~4900 lines.
func BenchmarkParseAndNavigate(b *testing.B) {
	text := bigFixture(700)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
		if err != nil {
			b.Fatal(err)
		}
		sf := p.CreateSourceFile("/big.ts", text)
		if got := len(sf.Classes()); got != 700 {
			b.Fatalf("classes: got %d", got)
		}
	}
}

func BenchmarkEditAndReparse(b *testing.B) {
	text := bigFixture(700)
	p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		b.Fatal(err)
	}
	sf := p.CreateSourceFile("/big.ts", text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := sf.Class("K0")
		c.AddProperty(PropertyStructure{Name: "added", Type: "number"})
	}
}
