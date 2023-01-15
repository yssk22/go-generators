package graphql

import (
	"errors"
	"fmt"
	"go/build"
	"go/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	ErrNoQuery = fmt.Errorf("no graphql root Query is found. type Query struct {...} is required")
	ErrSyntax  = fmt.Errorf("cannot parse package")
)

// standard packages to load with the target
var standardPackages = []string{
	"context",
	"time",
}

type TypeHelper interface {
	IsContext(t types.Type) bool
	IsError(t types.Type) bool
}

type builder struct {
	standardPackageMap map[string]*packages.Package
	targetPackage      *packages.Package

	contextType *types.Interface

	// options
	rootQueryName    string
	rootMutationName string
}

func (b *builder) IsContext(t types.Type) bool {
	return types.Implements(t, b.contextType)
}

func (b *builder) IsError(t types.Type) bool {
	return types.Implements(t, types.Universe.Lookup("error").Type().Underlying().(*types.Interface))
}

// Build analyzes the src package and returns a list of GraphQLObject
func Build(src string, options ...Option) ([]GraphQLObject, error) {
	builder, err := newBuilder(src)
	if err != nil {
		return nil, err
	}
	// apply optiions
	for _, opts := range options {
		builder = opts(builder)
	}
	rootQuery := builder.getRoot(builder.rootQueryName, rootDependencyTypeQuery)
	if rootQuery == nil {
		return nil, ErrNoQuery
	}
	var dependencies = []Dependency{
		rootQuery,
	}
	if rootMutation := builder.getRoot(builder.rootMutationName, rootDependencyTypeMutation); rootMutation != nil {
		dependencies = append(dependencies, rootMutation)
	}

	var graphQLObjects = make(map[string]*GraphQLObject)
	var graphQLObjectList []GraphQLObject
	// iterete dependency graph to collect GraphQLObject
	for len(dependencies) > 0 {
		dep := dependencies[0]
		obj, newDeps, err := dep.ToGraphQLObject(builder)
		if err != nil {
			return nil, fmt.Errorf("canont build GraphQLObject of %s: %w", dep, err)
		}
		// if the object is already registered in graphQLObject, then skip it's dependency resolution.
		if _, ok := graphQLObjects[obj.Name]; ok {
			// if already have GraphQLObject then we can skip newDep check
			dependencies = dependencies[1:]
			continue
		}
		graphQLObjects[obj.Name] = obj
		graphQLObjectList = append(graphQLObjectList, *obj)
		dependencies = append(dependencies[1:], newDeps...)
	}
	return graphQLObjectList, nil
}

func newBuilder(dir string) (*builder, error) {
	b := &builder{
		standardPackageMap: make(map[string]*packages.Package),

		// default option values
		rootQueryName:    "Query",
		rootMutationName: "Mutation",
	}
	importPath, err := b.resolveGoImportPath(dir)
	if err != nil {
		return nil, err
	}
	for _, s := range standardPackages {
		b.standardPackageMap[s] = nil
	}
	cfg := &packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedDeps | packages.NeedName | packages.NeedSyntax,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, append(standardPackages, importPath)...)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("%s: no such file or directory", dir)) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	for _, p := range pkgs {
		if _, ok := b.standardPackageMap[p.Name]; ok {
			b.standardPackageMap[p.Name] = p
		} else {
			if len(p.Errors) > 0 {
				return nil, fmt.Errorf("%w: %s", ErrSyntax, p.Errors[0])
			}
			b.targetPackage = p
		}
	}
	// fill build-in / standard types for TypesHelper
	b.contextType = b.standardPackageMap["context"].Types.Scope().Lookup("Context").Type().Underlying().(*types.Interface)
	return b, nil
}

func (b *builder) getRoot(name string, depType rootDependencyType) Dependency {
	obj := b.targetPackage.Types.Scope().Lookup(name)
	if obj == nil {
		return nil
	}
	named, strct := b.getNamedStruct(obj.Type())
	if named == nil || strct == nil {
		return nil
	}
	return &RootDependency{
		namedRef:  named,
		structRef: strct,
		depType:   depType,
	}
}

// resolveGoImportPath resolves directry path to go import statement path.
func (b *builder) resolveGoImportPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absModuleRootPath, moduleName, err := b.findGoModInfo(absPath)
	if err != nil {
		if err != ErrGoModNotFound {
			return "", err
		}
		// go.mod not found
		absGoPath, err := filepath.Abs(build.Default.GOPATH)
		if err != nil {
			return "", err
		}
		absGoPath = filepath.Join(absGoPath, "src")
		if !strings.HasPrefix(absPath, absGoPath) {
			return "", fmt.Errorf("not in $GOPATH/src (%s)", absGoPath)
		}
		offset := len(absGoPath) + 1
		return absPath[offset:], nil
	}
	// absPath is ${absModuleRootPath}/my/package/path/dir
	// and the import path should be ${moduleName}/my/package/path/dir
	return path.Join(moduleName, strings.TrimPrefix(absPath, absModuleRootPath)), nil
}

var (
	moduleDefRe      = regexp.MustCompile("module\\s+(\\S+)\n")
	ErrGoModNotFound = errors.New("go.mod file is not found in the directory and it's parents")
)

// findGoModInfo finds go.mod file under the directory (or parent directories)
// and resturns the directory path where go.mod exists and module string declared in go.mod file.
func (b *builder) findGoModInfo(dir string) (string, string, error) {
	gomod := filepath.Join(dir, "go.mod")
	_, err := os.Stat(gomod)
	if err != nil {
		if os.IsNotExist(err) {
			if dir == "/" {
				return "", "", ErrGoModNotFound
			}
			return b.findGoModInfo(filepath.Join(dir, "..") + "/")
		}
		return "", "", err
	}
	contents, err := ioutil.ReadFile(gomod)
	if err != nil {
		return "", "", err
	}
	found := moduleDefRe.Copy().FindSubmatch(contents)
	if len(found) == 0 {
		return "", "", fmt.Errorf("could not find module declaration in %s", gomod)
	}
	return dir, string(found[1]), nil
}

func (b *builder) getNamedStruct(t types.Type) (*types.Named, *types.Struct) {
	named, ok := t.(*types.Named)
	if !ok {
		return nil, nil
	}
	strct, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, nil
	}
	return named, strct
}
