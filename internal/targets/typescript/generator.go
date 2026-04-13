package typescript

import (
	"fmt"
	"sort"
	"strings"

	"github.com/81120/thrift2x/internal/ast"
)

type Generator struct {
	i64Type string
}

type Option func(*Generator)

func WithI64Type(tsType string) Option {
	return func(g *Generator) {
		t := strings.TrimSpace(tsType)
		if t == "" {
			return
		}
		g.i64Type = t
	}
}

func New(opts ...Option) *Generator {
	g := &Generator{i64Type: "string"}
	for _, opt := range opts {
		if opt != nil {
			opt(g)
		}
	}
	return g
}

func (g *Generator) Generate(f *ast.File) string {
	var body strings.Builder
	usedIncludeAliases := map[string]struct{}{}

	for _, d := range f.Decls {
		switch v := d.(type) {
		case ast.Enum:
			g.emitEnum(&body, v)
		case ast.StructLike:
			if v.Kind == ast.StructKind || v.Kind == ast.UnionKind || v.Kind == ast.ExceptionKind {
				g.emitInterface(&body, v)
				for _, field := range v.Fields {
					g.collectUsedAliasesInType(field.Type, usedIncludeAliases)
				}
			}
		case ast.Typedef:
			body.WriteString(fmt.Sprintf("export type %s = %s;\n\n", sanitizeTypeName(v.Alias), g.toTSType(v.Type)))
			g.collectUsedAliasesInType(v.Type, usedIncludeAliases)
		}
	}

	if body.Len() == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("/* eslint-disable */\n")
	out.WriteString("// generated from thrift\n\n")
	g.emitIncludes(&out, f.Includes, usedIncludeAliases)
	out.WriteString(body.String())
	return out.String()
}

func (g *Generator) collectUsedAliasesInType(t ast.TypeRef, used map[string]struct{}) {
	switch v := t.(type) {
	case ast.Identifier:
		if strings.Contains(v.Name, ".") {
			parts := strings.Split(v.Name, ".")
			if len(parts) >= 2 {
				used[sanitizeIdentifier(parts[0])] = struct{}{}
			}
		}
	case ast.List:
		g.collectUsedAliasesInType(v.Elem, used)
	case ast.Set:
		g.collectUsedAliasesInType(v.Elem, used)
	case ast.Map:
		g.collectUsedAliasesInType(v.Key, used)
		g.collectUsedAliasesInType(v.Value, used)
	}
}

func (g *Generator) emitIncludes(b *strings.Builder, includes []ast.Include, usedAliases map[string]struct{}) {
	if len(includes) == 0 {
		return
	}
	seen := map[string]struct{}{}
	emitted := 0
	for _, inc := range includes {
		alias := includeAlias(inc.Path)
		if _, ok := usedAliases[alias]; !ok {
			continue
		}
		path := includeTSPath(inc.Path)
		key := alias + "|" + path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		b.WriteString(fmt.Sprintf("import * as %s from %q;\n", alias, path))
		emitted++
	}
	if emitted > 0 {
		b.WriteString("\n")
	}
}

func (g *Generator) emitEnum(b *strings.Builder, e ast.Enum) {
	b.WriteString(fmt.Sprintf("export enum %s {\n", sanitizeTypeName(e.Name)))
	next := int64(0)
	for _, v := range e.Values {
		if v.Value != nil {
			next = *v.Value
		}
		b.WriteString(fmt.Sprintf("  %s = %d,\n", sanitizeMemberName(v.Name), next))
		next++
	}
	b.WriteString("}\n\n")
}

func (g *Generator) emitInterface(b *strings.Builder, s ast.StructLike) {
	b.WriteString(fmt.Sprintf("export interface %s {\n", sanitizeTypeName(s.Name)))
	for _, f := range s.Fields {
		op := ""
		if f.Requiredness == ast.RequirednessOptional {
			op = "?"
		}
		b.WriteString(fmt.Sprintf("  %s%s: %s;\n", sanitizeFieldName(f.Name), op, g.toTSType(f.Type)))
	}
	b.WriteString("}\n\n")
}

func (g *Generator) toTSType(t ast.TypeRef) string {
	switch v := t.(type) {
	case ast.Builtin:
		switch v.Name {
		case ast.TypeBool:
			return "boolean"
		case ast.TypeString:
			return "string"
		case ast.TypeVoid:
			return "void"
		case ast.TypeBinary:
			return "Uint8Array"
		case ast.TypeI64:
			return g.i64Type
		default:
			return "number"
		}
	case ast.Identifier:
		return toIdentifierTypeRef(v.Name)
	case ast.List:
		return g.wrapArrayElem(g.toTSType(v.Elem)) + "[]"
	case ast.Set:
		return g.wrapArrayElem(g.toTSType(v.Elem)) + "[]"
	case ast.Map:
		k := g.toTSType(v.Key)
		if !isRecordKeyType(k) {
			k = "string"
		}
		return fmt.Sprintf("Record<%s, %s>", k, g.toTSType(v.Value))
	default:
		return "any"
	}
}

func (g *Generator) wrapArrayElem(elem string) string {
	if strings.Contains(elem, "|") || strings.HasPrefix(elem, "Record<") {
		return "(" + elem + ")"
	}
	return elem
}

func isRecordKeyType(s string) bool {
	allowed := []string{"string", "number", "symbol"}
	for _, a := range allowed {
		if s == a {
			return true
		}
	}
	return false
}

func sanitizeTypeName(s string) string {
	return sanitizeIdentifier(s)
}

func sanitizeFieldName(s string) string {
	return sanitizeIdentifier(s)
}

func sanitizeMemberName(s string) string {
	return sanitizeIdentifier(s)
}

func sanitizeIdentifier(s string) string {
	if s == "" {
		return "_"
	}
	var b strings.Builder
	for i, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		out = "_"
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "_" + out
	}
	reserved := map[string]struct{}{
		"default": {}, "class": {}, "function": {}, "var": {}, "let": {}, "const": {}, "enum": {}, "interface": {},
	}
	if _, ok := reserved[out]; ok {
		out += "_"
	}
	return out
}

func toIdentifierTypeRef(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		if len(parts) == 2 {
			return sanitizeIdentifier(parts[0]) + "." + sanitizeTypeName(parts[1])
		}
	}
	return sanitizeTypeName(name)
}

func includeAlias(path string) string {
	base := path
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	base = strings.TrimSuffix(base, ".thrift")
	return sanitizeIdentifier(base)
}

func includeTSPath(path string) string {
	var out string
	if strings.HasSuffix(path, ".thrift") {
		out = strings.TrimSuffix(path, ".thrift") + ".ts"
	} else {
		out = path + ".ts"
	}
	if strings.HasPrefix(out, "./") || strings.HasPrefix(out, "../") || strings.HasPrefix(out, "/") {
		return out
	}
	return "./" + out
}

func CollectDeclNames(f *ast.File) []string {
	out := make([]string, 0, len(f.Decls))
	for _, d := range f.Decls {
		if d != nil {
			out = append(out, d.GetName())
		}
	}
	sort.Strings(out)
	return out
}
