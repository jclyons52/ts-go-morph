package tsmorph_test

import (
	"fmt"

	"github.com/jclyons52/ts-go-morph"
)

func Example() {
	p, err := tsmorph.NewProject(tsmorph.ProjectOptions{UseInMemoryFileSystem: true})
	if err != nil {
		panic(err)
	}

	sf := p.CreateSourceFile("/src/person.ts", "export class Person {\n  name: string;\n}\n")

	// Navigate.
	c, _ := sf.Class("Person")
	fmt.Println("class:", c.Name())
	for _, prop := range c.Properties() {
		fmt.Println("property:", prop.Name())
	}

	// Manipulate. Edits apply immediately and c is forgotten.
	c.AddMethod(tsmorph.MethodStructure{
		Name:       "greet",
		ReturnType: "string",
		Body:       `return "hi " + this.name;`,
	})

	// Re-fetch and inspect the result.
	c, _ = sf.Class("Person")
	fmt.Println("methods:", len(c.Methods()))
	fmt.Print(sf.Text())

	// Query types.
	m := c.Methods()[0]
	fmt.Println("return type:", m.Type().CallSignatures()[0].ReturnType().Text())

	// Output:
	// class: Person
	// property: name
	// methods: 1
	// export class Person {
	//   name: string;
	//
	//   greet(): string {
	//     return "hi " + this.name;
	//   }
	// }
	// return type: string
}
