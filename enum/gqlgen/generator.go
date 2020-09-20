package gqlgen

import (
	"fmt"
	"io"

	"github.com/yssk22/go-generators/enum"
)

const (
	generatedFilename = "gqlgen_enums.go"
)

type generator struct {
}

func NewGenerator() enum.Generator {
	return &generator{}
}

func (g *generator) Filename() string {
	return "gqlgen_enums.go"
}

func (g *generator) Generate(out io.Writer, enums []enum.EnumType) error {
	fmt.Fprintf(out, "import (\n")
	fmt.Fprintf(out, "\t\"fmt\"\n")
	fmt.Fprintf(out, "\t\"io\"\n")
	fmt.Fprintf(out, "\t\"strconv\"\n")
	fmt.Fprintf(out, ")\n")
	fmt.Fprintf(out, "\n")
	for _, e := range enums {
		g.writeMarshalGraphQL(e, out)
		fmt.Fprint(out, "\n")
		g.writeUnmarshalGraphQL(e, out)
		fmt.Fprint(out, "\n")
	}
	return nil
}

func (g *generator) writeMarshalGraphQL(e enum.EnumType, w io.Writer) {
	fmt.Fprintf(w, "func (e %s) MarshalGQL(w io.Writer) {\n", e.Name)
	fmt.Fprintf(w, "\tswitch e {\n")
	for _, c := range e.Keys {
		fmt.Fprintf(w, "\tcase %s:\n", c.GoName)
		fmt.Fprintf(w, "\t\tfmt.Fprintf(w, strconv.Quote(%q))\n", c.Name)
		fmt.Fprintf(w, "\t\tbreak\n")
	}
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "}\n")
}

func (g *generator) writeUnmarshalGraphQL(e enum.EnumType, w io.Writer) {
	fmt.Fprintf(w, "func (e *%s) UnmarshalGQL(v interface{}) error {\n", e.Name)
	fmt.Fprintf(w, "\tswitch v.(string) {\n")
	for _, c := range e.Keys {
		fmt.Fprintf(w, "\tcase %q:\n", c.Name)
		fmt.Fprintf(w, "\t\t*e = %s\n", c.GoName)
		fmt.Fprintf(w, "\t\tbreak\n")
	}
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")
}
