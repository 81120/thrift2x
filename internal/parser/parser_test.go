package parser

import (
	"testing"
)

func TestParseBasicDecls(t *testing.T) {
	src := `
include "base.thrift"
namespace go demo.test

typedef i64 UserID

const i32 MaxN = 10

enum Status {
  Unknown = 0,
  Active = 1,
  Banned,
}

struct User {
  1: required UserID id
  2: optional string name
  3: list<i64> scores
  4: map<string, i32> attrs
}

union Payload {
  1: string s
  2: i64 n
}

service UserService {
  User getUser(1: i64 id) throws (1: string msg)
}
`
	f, err := Parse("a.thrift", src)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(f.Includes) != 1 {
		t.Fatalf("includes = %d", len(f.Includes))
	}
	if len(f.Namespaces) != 1 {
		t.Fatalf("namespaces = %d", len(f.Namespaces))
	}
	if len(f.Decls) < 6 {
		t.Fatalf("decls too few: %d", len(f.Decls))
	}
}

func TestParseFieldRequiredness(t *testing.T) {
	src := `
struct A {
  1: optional string x
  2: required i64 y
  3: i32 z
}
`
	f, err := Parse("a.thrift", src)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(f.Decls) != 1 {
		t.Fatalf("decls = %d", len(f.Decls))
	}
}
