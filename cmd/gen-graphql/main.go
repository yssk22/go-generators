package main

import (
	"fmt"
	"os"

	"github.com/yssk22/go-generators/graphql"
	"github.com/yssk22/go-generators/graphql/gqlgen"
	"github.com/yssk22/go-generators/helper"
)

const usage = `Usage:
	gen-graphql {path/to/target}
`

func main() {
	if len(os.Args) < 3 {
		helper.ExitWithError(fmt.Errorf("source and target directory must be specified"), usage)
	}
	err := graphql.Generate(os.Args[1], gqlgen.NewGenerator(os.Args[2]))
	if err != nil {
		helper.ExitWithError(err, "")
	}
}
