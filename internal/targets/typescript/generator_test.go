package typescript

import (
	"strings"
	"testing"

	"github.com/81120/thrift2x/internal/ast"
)

func TestGenerateMappings(t *testing.T) {
	f := &ast.File{
		Includes: []ast.Include{{Path: "../../base.thrift"}, {Path: "./unused.thrift"}, {Path: "local.thrift"}},
		Decls: []ast.Decl{
			ast.StructLike{
				Kind: ast.StructKind,
				Name: "Demo",
				Fields: []ast.Field{
					{Name: "a", Requiredness: ast.RequirednessOptional, Type: ast.Builtin{Name: ast.TypeI64}},
					{Name: "b", Type: ast.Builtin{Name: ast.TypeDouble}},
					{Name: "c", Type: ast.Builtin{Name: ast.TypeString}},
					{Name: "d", Type: ast.Builtin{Name: ast.TypeBool}},
					{Name: "e", Type: ast.List{Elem: ast.Builtin{Name: ast.TypeI32}}},
					{Name: "ref", Type: ast.Identifier{Name: "base.Base"}},
					{Name: "localRef", Type: ast.Identifier{Name: "local.Node"}},
				},
			},
			ast.Enum{Name: "E", Values: []ast.EnumValue{{Name: "X", Value: int64ptr(1)}, {Name: "Y", Value: nil}}},
		},
	}

	out := New().Generate(f)

	checks := []string{
		"import * as base from \"../../base.ts\";",
		"import * as local from \"./local.ts\";",
		"export interface Demo",
		"a?: string;",
		"b: number;",
		"c: string;",
		"d: boolean;",
		"e: number[];",
		"ref: base.Base;",
		"localRef: local.Node;",
		"export enum E",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("output missing %q\n%s", c, out)
		}
	}
	if strings.Contains(out, "import * as unused") {
		t.Fatalf("unexpected unused import emitted:\n%s", out)
	}
}

func TestGenerateEmptyOutputForNoDecls(t *testing.T) {
	f := &ast.File{Includes: []ast.Include{{Path: "../../base.thrift"}}}
	out := New().Generate(f)
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty output for file with no declarations, got:\n%s", out)
	}
}

func TestGenerateMappingsWithI64Override(t *testing.T) {
	f := &ast.File{Decls: []ast.Decl{
		ast.StructLike{
			Kind:   ast.StructKind,
			Name:   "OnlyI64",
			Fields: []ast.Field{{Name: "id", Type: ast.Builtin{Name: ast.TypeI64}}},
		},
	}}
	out := New(WithI64Type("number")).Generate(f)
	if !strings.Contains(out, "id: number;") {
		t.Fatalf("expected i64 override to number, got:\n%s", out)
	}
}

func int64ptr(v int64) *int64 { return &v }
