package tsmorph

import (
	"strings"
	"testing"
)

func TestNodePrint(t *testing.T) {
	p, err := NewProject(ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.CreateSourceFile("/a.ts", "export   class   A{x:number}\n")

	c, _ := sf.Class("A")
	printed := c.Print()
	if !strings.Contains(printed, "class A") || !strings.Contains(printed, "x: number") {
		t.Fatalf("unexpected print output: %q", printed)
	}

	filePrinted := sf.Print()
	if !strings.Contains(filePrinted, "class A") {
		t.Fatalf("unexpected file print output: %q", filePrinted)
	}
}
