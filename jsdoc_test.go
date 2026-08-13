package tsmorph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const jsdocFixture = `/**
 * Adds two numbers.
 * @param {number} a - the first number
 * @param {number} b - the second number
 * @returns {number} the sum
 */
export function add(a: number, b: number): number {
  return a + b;
}

/**
 * A widget.
 * @typedef {Object} Widget
 * @property {string} name - the name
 * @property {number} [size] - the size
 * @see OtherType
 * @deprecated use NewWidget instead
 */
export class Widget {}

/**
 * A tagged template helper.
 * @template T
 * @param {string[]} strings
 * @returns {string}
 */
export function tag<T>(strings: TemplateStringsArray, ...values: T[]): string {
  return strings.join("");
}
`

func jsdocProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "jsdoc.ts")
	if err := os.WriteFile(path, []byte(jsdocFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func TestGetJsDocs(t *testing.T) {
	_, sf := jsdocProject(t)

	add, ok := sf.Function("add")
	if !ok {
		t.Fatal("add not found")
	}
	docs := add.GetJsDocs()
	if len(docs) != 1 {
		t.Fatalf("expected 1 jsdoc, got %d", len(docs))
	}
	if !strings.Contains(docs[0].GetComment(), "Adds two numbers") {
		t.Fatalf("unexpected jsdoc comment: %q", docs[0].GetComment())
	}
	if docs[0].GetInnerText() == "" {
		t.Fatal("expected inner text")
	}
	if !docs[0].IsMultiLine() {
		t.Fatal("expected multi-line jsdoc")
	}

	tags := docs[0].GetTags()
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d: %v", len(tags), tagNames(tags))
	}
	if tags[0].TagName() != "param" {
		t.Fatalf("expected param, got %q", tags[0].TagName())
	}
	if tags[2].TagName() != "returns" {
		t.Fatalf("expected returns, got %q", tags[2].TagName())
	}
}

func TestJSDocParameterTag(t *testing.T) {
	_, sf := jsdocProject(t)
	add, _ := sf.Function("add")

	param, ok := add.GetJsDocs()[0].GetTags()[0].AsJSDocParameterTag()
	if !ok {
		t.Fatal("expected JSDocParameterTag")
	}
	name, ok := param.Name()
	if !ok || name.Text() != "a" {
		t.Fatalf("expected param name 'a', got %q", name.Text())
	}
	if _, ok := param.TypeExpression(); !ok {
		t.Fatal("expected param type expression")
	}
	if param.IsBracketed() {
		t.Fatal("param should not be bracketed")
	}
	if !strings.Contains(param.GetComment(), "the first number") {
		t.Fatalf("unexpected param comment: %q", param.GetComment())
	}

	ret, ok := add.GetJsDocs()[0].GetTags()[2].AsJSDocReturnTag()
	if !ok {
		t.Fatal("expected JSDocReturnTag")
	}
	if _, ok := ret.TypeExpression(); !ok {
		t.Fatal("expected return type expression")
	}
}

func TestJSDocPropertyAndOtherTags(t *testing.T) {
	_, sf := jsdocProject(t)
	widget, _ := sf.Class("Widget")

	docs := widget.GetJsDocs()
	if len(docs) != 1 {
		t.Fatalf("expected 1 jsdoc, got %d", len(docs))
	}
	tags := docs[0].GetTags()

	var sawSee, sawDeprecated bool
	for _, tag := range tags {
		switch tag.TagName() {
		case "typedef":
			// @property tags after @typedef are nested in the type literal.
			td, ok := tag.AsJSDocTypedefTag()
			if !ok {
				continue
			}
			te, ok := td.TypeExpression()
			if !ok {
				t.Fatal("expected typedef type expression")
			}
			lit, ok := te.AsJSDocTypeLiteral()
			if !ok {
				t.Fatal("expected typedef type literal")
			}
			props := lit.GetPropertyTags()
			if len(props) != 2 {
				t.Fatalf("expected 2 @property tags, got %d", len(props))
			}
			if n, ok := props[0].Name(); !ok || n.Text() != "name" {
				t.Fatalf("expected first property 'name', got %q", n.Text())
			}
			if _, ok := props[0].TypeExpression(); !ok {
				t.Fatal("expected property type expression")
			}
			if !props[1].IsBracketed() {
				t.Fatal("expected bracketed property [size]")
			}
		case "see":
			if s, ok := tag.AsJSDocSeeTag(); ok {
				if ne, ok := s.NameExpression(); ok && ne.Text() == "OtherType" {
					sawSee = true
				}
			}
		case "deprecated":
			sawDeprecated = true
		}
	}
	if !sawSee {
		t.Fatal("expected @see OtherType")
	}
	if !sawDeprecated {
		t.Fatal("expected @deprecated")
	}
}

func TestJSDocTemplateTag(t *testing.T) {
	_, sf := jsdocProject(t)
	tagFn, _ := sf.Function("tag")

	tags := tagFn.GetJsDocs()[0].GetTags()
	for _, tag := range tags {
		if tag.TagName() == "template" {
			tpl, ok := tag.AsJSDocTemplateTag()
			if !ok {
				t.Fatal("expected JSDocTemplateTag")
			}
			if params := tpl.TypeParameters(); len(params) != 1 {
				t.Fatalf("expected 1 type parameter, got %d", len(params))
			}
			return
		}
	}
	t.Fatal("expected @template tag")
}

func tagNames(tags []JSDocTag) []string {
	var out []string
	for _, tag := range tags {
		out = append(out, tag.TagName())
	}
	return out
}
