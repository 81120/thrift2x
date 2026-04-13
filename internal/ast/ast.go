package ast

import "fmt"

type File struct {
	Path       string
	Includes   []Include
	Namespaces []Namespace
	Decls      []Decl
}

type Include struct {
	Path string
}

type Namespace struct {
	Scope string
	Name  string
}

type Decl interface {
	declNode()
	GetName() string
}

type Typedef struct {
	Alias string
	Type  TypeRef
}

func (Typedef) declNode()         {}
func (d Typedef) GetName() string { return d.Alias }

type Const struct {
	Name  string
	Type  TypeRef
	Value string
}

func (Const) declNode()         {}
func (d Const) GetName() string { return d.Name }

type Enum struct {
	Name   string
	Values []EnumValue
}

func (Enum) declNode()         {}
func (d Enum) GetName() string { return d.Name }

type EnumValue struct {
	Name  string
	Value *int64
}

type StructLikeKind string

const (
	StructKind    StructLikeKind = "struct"
	UnionKind     StructLikeKind = "union"
	ExceptionKind StructLikeKind = "exception"
)

type StructLike struct {
	Kind   StructLikeKind
	Name   string
	Fields []Field
}

func (StructLike) declNode()         {}
func (d StructLike) GetName() string { return d.Name }

type Service struct {
	Name    string
	Extends string
	Methods []Method
}

func (Service) declNode()         {}
func (d Service) GetName() string { return d.Name }

type Method struct {
	Name       string
	ReturnType TypeRef
	Params     []Field
	Throws     []Field
	Oneway     bool
}

type Requiredness string

const (
	RequirednessDefault  Requiredness = "default"
	RequirednessRequired Requiredness = "required"
	RequirednessOptional Requiredness = "optional"
)

type Field struct {
	ID           *int
	Requiredness Requiredness
	Type         TypeRef
	Name         string
	Default      string
}

type TypeRef interface {
	typeNode()
	String() string
}

type BuiltinType string

const (
	TypeBool   BuiltinType = "bool"
	TypeByte   BuiltinType = "byte"
	TypeI16    BuiltinType = "i16"
	TypeI32    BuiltinType = "i32"
	TypeI64    BuiltinType = "i64"
	TypeInt    BuiltinType = "int"
	TypeDouble BuiltinType = "double"
	TypeString BuiltinType = "string"
	TypeBinary BuiltinType = "binary"
	TypeVoid   BuiltinType = "void"
)

type Builtin struct {
	Name BuiltinType
}

func (Builtin) typeNode()        {}
func (t Builtin) String() string { return string(t.Name) }

type Identifier struct {
	Name string
}

func (Identifier) typeNode()        {}
func (t Identifier) String() string { return t.Name }

type List struct {
	Elem TypeRef
}

func (List) typeNode() {}
func (t List) String() string {
	return fmt.Sprintf("list<%s>", t.Elem.String())
}

type Set struct {
	Elem TypeRef
}

func (Set) typeNode() {}
func (t Set) String() string {
	return fmt.Sprintf("set<%s>", t.Elem.String())
}

type Map struct {
	Key   TypeRef
	Value TypeRef
}

func (Map) typeNode() {}
func (t Map) String() string {
	return fmt.Sprintf("map<%s,%s>", t.Key.String(), t.Value.String())
}
