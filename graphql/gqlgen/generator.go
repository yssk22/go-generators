package gqlgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/yssk22/go-generators/graphql"
)

const (
	GQLGenSchemaFile   = "schema.graphql"
	GQLGenResolverFile = "resolver.go"
	GQLGenConfigFile   = "gqlgen.yml"
)

type generator struct {
	Dir          string // target directory
	RunGQLGen    bool   // not only generate schema but also generate generate gqlgen code
	UseSnakeCase bool
}

func NewGenerator(dir string) graphql.Generator {
	return &generator{
		Dir:          dir,
		RunGQLGen:    true,
		UseSnakeCase: false,
	}
}

func (g *generator) Generate(list []graphql.GraphQLObject) error {
	if err := os.MkdirAll(g.Dir, 0755); err != nil {
		return err
	}
	if err := g.generateSchema(list); err != nil {
		return err
	}
	if g.RunGQLGen {
		if err := g.generateResolver(list); err != nil {
			return err
		}
		if err := g.generateGQLGenCode(); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generateSchema(list []graphql.GraphQLObject) error {
	// schema.graphql
	schemaFile, err := g.openFile(GQLGenSchemaFile)
	if err != nil {
		return err
	}
	defer schemaFile.Close()
	if err := gqlGenSchemaTemplate.Execute(schemaFile, map[string]interface{}{
		"Objects": list,
	}); err != nil {
		return err
	}
	return nil
}

func (g *generator) generateResolver(list []graphql.GraphQLObject) error {
	resolverFile, err := g.openFile(GQLGenResolverFile)
	if err != nil {
		return err
	}
	defer resolverFile.Close()
	var targetPackageName = filepath.Base(g.Dir)
	var mainPackagePath string
	var mainPackageRef string
	var queryStructName string
	var mutationStructName string
	for _, p := range list {
		if p.Name == graphql.RootQueryObjectName {
			tmp := strings.Split(p.GoModel, ".")
			mainPackagePath = strings.Join(tmp[0:len(tmp)-1], ".")
			queryStructName = tmp[len(tmp)-1]
		} else if p.Name == graphql.RootMutationObjectName {
			tmp := strings.Split(p.GoModel, ".")
			mutationStructName = tmp[len(tmp)-1]
		}
	}
	tmp := strings.Split(mainPackagePath, "/")
	mainPackageRef = tmp[len(tmp)-1]
	if err := gqlGenResolverTemplate.Execute(resolverFile, map[string]interface{}{
		"TargetPackage":      targetPackageName,
		"MainPackagePath":    mainPackagePath,
		"MainPackageRef":     mainPackageRef,
		"QueryStructName":    queryStructName,
		"MutationStructName": mutationStructName,
	}); err != nil {
		return err
	}
	return nil
}

func (g *generator) generateGQLGenCode() error {
	cfg := config.Config{
		SchemaFilename: config.StringList{filepath.Join(g.Dir, "schema.graphql")},
		Model:          config.PackageConfig{Filename: filepath.Join(g.Dir, "models_generated.go")},
		Exec:           config.PackageConfig{Filename: filepath.Join(g.Dir, "exec_generated.go")},
		Directives:     map[string]config.DirectiveConfig{},
		Models:         config.TypeMap{},
	}
	for _, filename := range cfg.SchemaFilename {
		filename = filepath.ToSlash(filename)
		schemaRaw, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		cfg.Sources = append(cfg.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}
	err := api.Generate(&cfg)
	return err
}

func (g *generator) openFile(name string) (*os.File, error) {
	return os.OpenFile(filepath.Join(g.Dir, name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func gqlGenObjectToString(obj graphql.GraphQLObject) string {
	goModelDerective := ""
	if obj.GoModel != "" {
		goModelDerective = fmt.Sprintf(" @goModel(model: %q)", obj.GoModel)
	}
	if obj.ObjectType == graphql.GraphQLObjectTypeScalar {
		return fmt.Sprintf("scalar %s%s", obj.Name, goModelDerective)
	}
	// even thought obj.ObjectType == "interface", we'll generate it as type since
	// gqlgen doesn't handle interface without implementator type.
	objectType := obj.ObjectType
	if objectType == graphql.GraphQLObjectTypeInterface {
		objectType = graphql.GraphQLObjectTypeType
	}
	var lines []string
	if objectType == graphql.GraphQLObjectTypeEnum {
		lines = append(lines, fmt.Sprintf("%s %s%s {", objectType, obj.Name, goModelDerective))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s%s {", objectType, obj.Name, goModelDerective))
	}
	for _, v := range obj.Values {
		lines = append(lines, fmt.Sprintf(
			"  %s", v,
		))
	}
	for _, f := range obj.Fields {
		lines = append(lines, fmt.Sprintf(
			"  %s: %s", f.Name, gqlGenFieldTypeString(&f),
		))
	}
	for _, m := range obj.Methods {
		if len(m.Parameters) == 0 {
			lines = append(lines, fmt.Sprintf(
				"  %s: %s", m.Name, gqlGenFieldTypeString(&m.ReturnValue),
			))
		} else {
			lines = append(lines, fmt.Sprintf("  %s(", m.Name))
			for i, p := range m.Parameters {
				comma := ","
				if i == len(m.Parameters)-1 {
					comma = ""
				}
				lines = append(lines, fmt.Sprintf("    %s: %s%s", p.Name, gqlGenFieldTypeString(&p), comma))
			}
			lines = append(lines, fmt.Sprintf("  ): %s", gqlGenFieldTypeString(&m.ReturnValue)))
		}
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func gqlGenFieldTypeString(f *graphql.GraphQLObjectField) string {
	suffix := "!"
	if f.Nullable {
		suffix = ""
	}
	if f.IsArray {
		underlyingSuffix := "!"
		if f.ElementNullable {
			underlyingSuffix = ""
		}
		return fmt.Sprintf("%s%s%s%s%s",
			strings.Repeat("[", f.NestDepth),
			f.Type,
			underlyingSuffix,
			strings.Repeat("]", f.NestDepth),
			suffix,
		)
	}
	return fmt.Sprintf(
		"%s%s",
		f.Type,
		suffix,
	)
}

var gqlGenConfigTemplate = template.Must(template.New("gqlGenSchemaTemplate").Parse(
	`# GENERATED BY go-gen-graphql-schema
schema:
	- ./schema.graphql
exec:
  filename: exec_generated.go
model:
  filename: models_generated.go
resolver:
  filename: resolver.go
`))

var gqlGenSchemaTemplate = template.Must(
	template.New("gqlGenSchemaTemplate").Funcs(map[string]interface{}{
		"objectToString": gqlGenObjectToString,
	}).Parse(`# GENERATED BY go-gen-graphql-schema
directive @goModel(
  model: String
  models: [String!]
) on OBJECT | INPUT_OBJECT | SCALAR | ENUM | INTERFACE | UNION

{{range .Objects}}
{{objectToString .}}
{{end -}}
`))

var gqlGenResolverTemplate = template.Must(
	template.New("gqlGenResolverTemplate").Parse(`// GENERATED BY go-gen-graphql-schema
package {{.TargetPackage}}

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

import "{{.MainPackagePath}}"

type Resolver struct{}

var query = &{{.MainPackageRef}}.{{.QueryStructName}}{}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return query }

{{if .MutationStructName }}
var mutation = &{{.MainPackageRef}}.{{.MutationStructName}}{}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return mutation }
{{end}}
`))
