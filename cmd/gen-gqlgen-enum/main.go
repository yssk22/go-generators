package main

import (
	"os"

	"github.com/yssk22/go-generators/enum"
	"github.com/yssk22/go-generators/enum/gqlgen"
	"github.com/yssk22/go-generators/helper"
)

const usage = `Usage:
	gen-enum {path/to/package}
`

func main() {
	var directories []string
	if len(os.Args) < 1 {
		directories = []string{"."}
	} else {
		directories = os.Args[1:]
	}
	for _, d := range directories {
		generator := gqlgen.NewGenerator()
		err := enum.Generate(d, generator)
		if err != nil {
			helper.ExitWithError(err, "")
		}
	}
}
