package typescript

import (
	"fmt"
	"strings"

	"github.com/81120/thrift2x/internal/targets"
)

const (
	Name          = "typescript"
	OptionI64Type = "i64-type"
)

type Target struct{}

func (Target) Name() string { return Name }

func (Target) FileExtension() string { return ".ts" }

func (Target) NewGenerator(options map[string]string) (targets.Generator, error) {
	if options == nil {
		return New(), nil
	}
	i64Type := strings.TrimSpace(options[OptionI64Type])
	if i64Type == "" {
		return New(), nil
	}
	switch i64Type {
	case "string", "number", "bigint":
		return New(WithI64Type(i64Type)), nil
	default:
		return nil, fmt.Errorf("unsupported i64-type %q for target %s", i64Type, Name)
	}
}

func init() {
	targets.Register(Target{})
}
