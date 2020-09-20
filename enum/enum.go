package enum

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yssk22/go-generators/helper"
	"golang.org/x/tools/go/packages"
)

/**
  GenEnum generates MarshalGraphQL and UnmarshalGraphQL for enum types

  ````
  type MyEnum string
  const (
	  MyEnumValueA = MyEnum("value_a")
	  MyEnumValueB = MyEnum("value_b")
  )
  ````

  This generates GraphQL enum MyEnum { ValueA ValueB } by go-gen-graphql-schema and
  and MyEnum is supposed to implement `MarshalGraphQL` and `UnmarshalGraphQL` to convert
  ValueA/ValueB to go constants MyEnumValueA/MyEnumValueB and vice versa.

  gqlgen-enum can be used to generate these methods so you don't have to write marshaler/unmarshaler by yourselve.
*/
func Generate(dir string, generators ...Generator) error {
	importPath, err := helper.ResolveGoImportPath(dir)
	if err != nil {
		return err
	}
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return err
	}
	pkg := pkgs[0]
	scope := pkg.Types.Scope()
	enums := getEnumList(scope)
	for _, g := range generators {
		err := func(g Generator) error {
			filename := g.Filename()
			file, err := os.OpenFile(filepath.Join(dir, filename), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			fmt.Fprintf(file, "package %s\n", pkg.Name)
			fmt.Fprintf(file, "\n")
			return g.Generate(file, enums)
		}(g)
		if err != nil {
			log.Printf("cannot generate %s: %v", filepath.Join(dir, g.Filename()), err)
		}
	}
	return nil
}

type EnumKey struct {
	GoName string
	Name   string
	Value  string
}

type EnumType struct {
	Name string
	Keys []EnumKey
}

func getEnumList(scope *types.Scope) []EnumType {
	nameToGQLGenEnum := make(map[*types.Named]*EnumType)
	names := scope.Names()
	for _, n := range names {
		obj := scope.Lookup(n)
		objType, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		_, ok = objType.Underlying().(*types.Basic)
		if !ok {
			continue
		}
		if _, ok = nameToGQLGenEnum[objType]; !ok {
			nameToGQLGenEnum[objType] = &EnumType{}
		}
		switch ot := obj.(type) {
		case *types.TypeName:
			nameToGQLGenEnum[objType].Name = ot.Id()
			break
		case *types.Const:
			nameToGQLGenEnum[objType].Keys = append(nameToGQLGenEnum[objType].Keys, *newEnumKey(ot, objType.Obj().Id()))
			break
		}
	}
	// convert map to slice
	var list []EnumType
	for _, v := range nameToGQLGenEnum {
		if len(v.Keys) > 0 {
			list = append(list, *v)
		}
	}
	return list
}

// GetEnum returns Enum for the named type
func GetEnum(t *types.Named) *EnumType {
	typeName := t.Obj().Name()
	scope := t.Obj().Pkg().Scope()
	var keys []EnumKey
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj.Type() == t {
			if c, ok := obj.(*types.Const); ok {
				keys = append(keys, *newEnumKey(c, typeName))
			}
		}
	}
	return &EnumType{
		Name: typeName,
		Keys: keys,
	}
}

func newEnumKey(c *types.Const, prefix string) *EnumKey {
	return &EnumKey{
		Name:   strings.TrimPrefix(c.Id(), prefix),
		GoName: c.Id(),
		Value:  c.Val().ExactString(),
	}
}
