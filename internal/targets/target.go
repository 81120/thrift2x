package targets

import "github.com/81120/thrift2x/internal/ast"

type Generator interface {
	Generate(file *ast.File) string
}

type Target interface {
	Name() string
	FileExtension() string
	NewGenerator(options map[string]string) (Generator, error)
}
