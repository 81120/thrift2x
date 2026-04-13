package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	normalizeLegacySingleDashLongFlags()
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func normalizeLegacySingleDashLongFlags() {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			os.Args[i] = "-" + arg
		}
	}
}
