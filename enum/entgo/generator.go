package entgo

import (
	"fmt"
	"io"

	"github.com/yssk22/go-generators/enum"
)

const (
	generatedFilename = "entgo_enums.go"
)

type generator struct {
}

func NewGenerator() enum.Generator {
	return &generator{}
}

func (g *generator) Filename() string {
	return "entgo_enums.go"
}

func (g *generator) Generate(out io.Writer, enums []enum.EnumType) error {
	for _, e := range enums {
		fmt.Fprintf(out, "func (%s) Values() (types []string) {\n", e.Name)
		fmt.Fprintf(out, "\tfor _, r := range []%s{\n", e.Name)
		for _, c := range e.Keys {
			fmt.Fprintf(out, "\t\t%s,\n", c.GoName)
		}
		fmt.Fprintf(out, "\t} {\n")
		fmt.Fprintf(out, "\t\ttypes = append(types, string(r))\n")
		fmt.Fprintf(out, "\t}\n")
		fmt.Fprintf(out, "\treturn\n")
		fmt.Fprintf(out, "}\n")
	}
	return nil
}
