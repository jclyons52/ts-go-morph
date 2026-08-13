package tsmorph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/ast"
)

const kindsFixture = `import type { User } from "./user";
import { log, type Config } from "./util";
import * as ns from "./ns";
import mod = require("./mod");
export { type A, B } from "./a";
export = 42;

namespace Outer {
  export namespace Inner {
    export const x: number = 1;
  }
}

interface Shape { width: number; height: number }

function demo(flag: boolean, list: number[]) {
  if (flag) {
    return "yes";
  } else {
    return "no";
  }

  for (let i = 0; i < 10; i++) {
    log(i);
  }

  for (const item of list) {
    log(item);
  }

  while (flag) {
    break;
  }

  try {
    throw new Error("boom");
  } catch (err) {
    log(err);
  } finally {
    log("done");
  }

  switch (flag) {
    case true:
      log("true");
      break;
    default:
      log("false");
  }

  const total: number = 1 + 2 * 3;
  const neg: number = -total;
  let count = 0;
  count++;
  const choice: string = flag ? "a" : "b";
  const arr: number[] = [1, 2, 3];
  const obj: { a: number } = { a: 1 };
  const first: number = arr[0];
  const inst = new Date();
  const asserted = total as number;
  await Promise.resolve();

  type Name = string;
  type Num = number;
  type Maybe = Name | null;
  type Arr = Name[];
  type Lit = "x";
  type Idx = Shape["width"];
  type Cond<T> = T extends string ? true : false;
  type Mapped = { [K in keyof Shape]: Shape[K] };
  const n: Name = "s";
  return n;
}
`

func kindsProject(t *testing.T) (*Project, *SourceFile) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "kinds.ts")
	if err := os.WriteFile(path, []byte(kindsFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	return p, p.SourceFile(path)
}

func mustNode(t *testing.T, sf *SourceFile, kind Kind) Node {
	t.Helper()
	n, ok := sf.FirstDescendantByKind(kind)
	if !ok {
		t.Fatalf("no descendant of kind %s", kind)
	}
	return n
}

func TestGeneratedKindPredicates(t *testing.T) {
	_, sf := kindsProject(t)

	cases := []struct {
		kind   Kind
		check  func(Node) bool
		asName string
	}{
		{ast.KindIfStatement, func(n Node) bool { return n.IsIfStatement() }, "IfStatement"},
		{ast.KindForStatement, func(n Node) bool { return n.IsForStatement() }, "ForStatement"},
		{ast.KindWhileStatement, func(n Node) bool { return n.IsWhileStatement() }, "WhileStatement"},
		{ast.KindTryStatement, func(n Node) bool { return n.IsTryStatement() }, "TryStatement"},
		{ast.KindCatchClause, func(n Node) bool { return n.IsCatchClause() }, "CatchClause"},
		{ast.KindSwitchStatement, func(n Node) bool { return n.IsSwitchStatement() }, "SwitchStatement"},
		{ast.KindBinaryExpression, func(n Node) bool { return n.IsBinaryExpression() }, "BinaryExpression"},
		{ast.KindPrefixUnaryExpression, func(n Node) bool { return n.IsPrefixUnaryExpression() }, "PrefixUnaryExpression"},
		{ast.KindPostfixUnaryExpression, func(n Node) bool { return n.IsPostfixUnaryExpression() }, "PostfixUnaryExpression"},
		{ast.KindConditionalExpression, func(n Node) bool { return n.IsConditionalExpression() }, "ConditionalExpression"},
		{ast.KindArrayLiteralExpression, func(n Node) bool { return n.IsArrayLiteralExpression() }, "ArrayLiteralExpression"},
		{ast.KindObjectLiteralExpression, func(n Node) bool { return n.IsObjectLiteralExpression() }, "ObjectLiteralExpression"},
		{ast.KindElementAccessExpression, func(n Node) bool { return n.IsElementAccessExpression() }, "ElementAccessExpression"},
		{ast.KindNewExpression, func(n Node) bool { return n.IsNewExpression() }, "NewExpression"},
		{ast.KindTypeReference, func(n Node) bool { return n.IsTypeReferenceNode() }, "TypeReferenceNode"},
		{ast.KindUnionType, func(n Node) bool { return n.IsUnionTypeNode() }, "UnionTypeNode"},
		{ast.KindArrayType, func(n Node) bool { return n.IsArrayTypeNode() }, "ArrayTypeNode"},
		{ast.KindLiteralType, func(n Node) bool { return n.IsLiteralTypeNode() }, "LiteralTypeNode"},
		{ast.KindIndexedAccessType, func(n Node) bool { return n.IsIndexedAccessTypeNode() }, "IndexedAccessTypeNode"},
		{ast.KindConditionalType, func(n Node) bool { return n.IsConditionalTypeNode() }, "ConditionalTypeNode"},
		{ast.KindMappedType, func(n Node) bool { return n.IsMappedTypeNode() }, "MappedTypeNode"},
		{ast.KindImportEqualsDeclaration, func(n Node) bool { return n.IsImportEqualsDeclaration() }, "ImportEqualsDeclaration"},
		{ast.KindExportAssignment, func(n Node) bool { return n.IsExportAssignment() }, "ExportAssignment"},
		{ast.KindModuleDeclaration, func(n Node) bool { return n.IsModuleDeclaration() }, "ModuleDeclaration"},
	}

	for _, c := range cases {
		n := mustNode(t, sf, c.kind)
		if !c.check(n) {
			t.Fatalf("Is%s returned false for a %s node", c.asName, c.asName)
		}
		if n.Kind() != c.kind {
			t.Fatalf("kind mismatch: got %s want %s", n.Kind(), c.kind)
		}
	}
}

func TestStatementAccessors(t *testing.T) {
	_, sf := kindsProject(t)

	ifStmt, ok := mustNode(t, sf, ast.KindIfStatement).AsIfStatement()
	if !ok {
		t.Fatal("AsIfStatement failed")
	}
	if _, ok := ifStmt.ThenStatement(); !ok {
		t.Fatal("expected then statement")
	}
	if _, ok := ifStmt.ElseStatement(); !ok {
		t.Fatal("expected else statement")
	}
	if _, ok := ifStmt.GetExpression(); !ok {
		t.Fatal("expected if condition via GetExpression")
	}

	forStmt, _ := mustNode(t, sf, ast.KindForStatement).AsForStatement()
	if _, ok := forStmt.Initializer(); !ok {
		t.Fatal("expected for initializer")
	}
	if _, ok := forStmt.Condition(); !ok {
		t.Fatal("expected for condition")
	}
	if _, ok := forStmt.Incrementor(); !ok {
		t.Fatal("expected for incrementor")
	}

	tryStmt, _ := mustNode(t, sf, ast.KindTryStatement).AsTryStatement()
	if _, ok := tryStmt.TryBlock(); !ok {
		t.Fatal("expected try block")
	}
	catch, ok := tryStmt.CatchClause()
	if !ok {
		t.Fatal("expected catch clause")
	}
	if _, ok := catch.Block(); !ok {
		t.Fatal("expected catch block")
	}
	if _, ok := tryStmt.FinallyBlock(); !ok {
		t.Fatal("expected finally block")
	}

	sw, _ := mustNode(t, sf, ast.KindSwitchStatement).AsSwitchStatement()
	caseBlock, ok := sw.CaseBlock()
	if !ok {
		t.Fatal("expected case block")
	}
	if clauses := caseBlock.Clauses(); len(clauses) != 2 {
		t.Fatalf("expected 2 case clauses, got %d", len(clauses))
	}
}

func TestExpressionKindAccessors(t *testing.T) {
	_, sf := kindsProject(t)

	bin, _ := mustNode(t, sf, ast.KindBinaryExpression).AsBinaryExpression()
	if _, ok := bin.Left(); !ok {
		t.Fatal("expected binary left")
	}
	if _, ok := bin.Right(); !ok {
		t.Fatal("expected binary right")
	}
	if bin.OperatorText() == "" {
		t.Fatal("expected binary operator text")
	}

	cond, _ := mustNode(t, sf, ast.KindConditionalExpression).AsConditionalExpression()
	if _, ok := cond.Condition(); !ok {
		t.Fatal("expected conditional condition")
	}
	if _, ok := cond.WhenTrue(); !ok {
		t.Fatal("expected when-true")
	}
	if _, ok := cond.WhenFalse(); !ok {
		t.Fatal("expected when-false")
	}

	prefix, _ := mustNode(t, sf, ast.KindPrefixUnaryExpression).AsPrefixUnaryExpression()
	if _, ok := prefix.Operand(); !ok {
		t.Fatal("expected prefix operand")
	}

	postfix, _ := mustNode(t, sf, ast.KindPostfixUnaryExpression).AsPostfixUnaryExpression()
	if _, ok := postfix.Operand(); !ok {
		t.Fatal("expected postfix operand")
	}

	elem, _ := mustNode(t, sf, ast.KindElementAccessExpression).AsElementAccessExpression()
	if _, ok := elem.ArgumentExpression(); !ok {
		t.Fatal("expected element access argument")
	}

	arr, _ := mustNode(t, sf, ast.KindArrayLiteralExpression).AsArrayLiteralExpression()
	if elems := arr.GetElements(); len(elems) != 3 {
		t.Fatalf("expected 3 array elements, got %d", len(elems))
	}

	obj, _ := mustNode(t, sf, ast.KindObjectLiteralExpression).AsObjectLiteralExpression()
	if props := obj.GetProperties(); len(props) != 1 {
		t.Fatalf("expected 1 object property, got %d", len(props))
	}
}

func TestTemplateLiteralAccessors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tmpl.ts")
	fixture := "const text: string = `hello ${list.length}`;\nconst tagged: string = String.raw`raw`;\n"
	if err := os.WriteFile(path, []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := NewProject(ProjectOptions{RootFilePaths: []string{path}})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.SourceFile(path)

	tpl, ok := mustNode(t, sf, ast.KindTemplateExpression).AsTemplateExpression()
	if !ok {
		t.Fatal("AsTemplateExpression failed")
	}
	if _, ok := tpl.Head(); !ok {
		t.Fatal("expected template head")
	}
	if spans := tpl.TemplateSpans(); len(spans) != 1 {
		t.Fatalf("expected 1 template span, got %d", len(spans))
	}
	if _, ok := tpl.TemplateSpans()[0].AsTemplateSpan(); !ok {
		t.Fatal("AsTemplateSpan failed")
	}

	tagged, ok := mustNode(t, sf, ast.KindTaggedTemplateExpression).AsTaggedTemplateExpression()
	if !ok {
		t.Fatal("AsTaggedTemplateExpression failed")
	}
	if _, ok := tagged.Template(); !ok {
		t.Fatal("expected tagged template")
	}
	if _, ok := tagged.Tag(); !ok {
		t.Fatal("expected tag via Tag()")
	}
}

func TestTypeNodeAccessors(t *testing.T) {
	_, sf := kindsProject(t)

	ref, _ := mustNode(t, sf, ast.KindTypeReference).AsTypeReferenceNode()
	if _, ok := ref.TypeName(); !ok {
		t.Fatal("expected type reference name")
	}

	union, _ := mustNode(t, sf, ast.KindUnionType).AsUnionTypeNode()
	if types := union.Types(); len(types) != 2 {
		t.Fatalf("expected 2 union members, got %d", len(types))
	}

	arr, _ := mustNode(t, sf, ast.KindArrayType).AsArrayTypeNode()
	if _, ok := arr.ElementType(); !ok {
		t.Fatal("expected array element type")
	}

	idx, _ := mustNode(t, sf, ast.KindIndexedAccessType).AsIndexedAccessTypeNode()
	if _, ok := idx.ObjectType(); !ok {
		t.Fatal("expected indexed access object type")
	}
	if _, ok := idx.IndexType(); !ok {
		t.Fatal("expected indexed access index type")
	}

	cond, _ := mustNode(t, sf, ast.KindConditionalType).AsConditionalTypeNode()
	if _, ok := cond.CheckType(); !ok {
		t.Fatal("expected conditional check type")
	}
	if _, ok := cond.TrueType(); !ok {
		t.Fatal("expected conditional true type")
	}
	if _, ok := cond.FalseType(); !ok {
		t.Fatal("expected conditional false type")
	}

	mapped, _ := mustNode(t, sf, ast.KindMappedType).AsMappedTypeNode()
	if _, ok := mapped.TypeParameter(); !ok {
		t.Fatal("expected mapped type parameter")
	}
}

func TestModuleAndSpecifierAccessors(t *testing.T) {
	_, sf := kindsProject(t)

	if mods := sf.ModuleDeclarations(); len(mods) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(mods))
	}
	if _, ok := sf.ModuleDeclaration("Outer"); !ok {
		t.Fatal("expected namespace Outer")
	}

	if ies := sf.ImportEqualsDeclarations(); len(ies) != 1 {
		t.Fatalf("expected 1 import=, got %d", len(ies))
	}
	ie := sf.ImportEqualsDeclarations()[0]
	if _, ok := ie.ModuleReference(); !ok {
		t.Fatal("expected import= module reference")
	}

	if eas := sf.ExportAssignments(); len(eas) != 1 {
		t.Fatalf("expected 1 export=, got %d", len(eas))
	}
	if !sf.ExportAssignments()[0].IsExportEquals() {
		t.Fatal("expected export= to be export equals")
	}

	// import type specifier
	imports := sf.ImportDeclarations()
	if len(imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(imports))
	}
}

func TestGenericAccessorsAreSafe(t *testing.T) {
	_, sf := kindsProject(t)
	root := sf.RootNode()

	// These must not panic on nodes that don't support the requested child.
	if _, ok := root.GetExpression(); ok {
		t.Fatal("GetExpression on source file should return false")
	}
	if args := root.GetArguments(); args != nil {
		t.Fatal("GetArguments on source file should return nil")
	}
	if stmts := root.GetStatements(); stmts == nil {
		t.Fatal("GetStatements on source file should return statements")
	}
	if _, ok := sf.Class("Nope"); ok {
		t.Fatal("unexpected class")
	}
	if got := mustNode(t, sf, ast.KindNumericLiteral); !got.HasQuestionDotToken() {
		// numeric literal has no optional chaining; must not panic
		_ = got
	}
	if props := mustNode(t, sf, ast.KindNumericLiteral).GetProperties(); props != nil {
		t.Fatal("GetProperties on a numeric literal should return nil")
	}
}
