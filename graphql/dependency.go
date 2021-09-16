package graphql

import (
	"fmt"
	"go/types"

	"github.com/yssk22/go-generators/enum"
	hh "github.com/yssk22/go-generators/helper"
)

// Dependency is an object to identify a dependency on types used in target package and
// it is a source of truth for the corresponding GraphQL object
type Dependency interface {
	ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error)
	IsCustomType() bool
}

type rootDependencyType string

var (
	rootDependencyTypeQuery    = rootDependencyType("query")
	rootDependencyTypeMutation = rootDependencyType("mutation")
)

var (
	ErrNoMethodsInQuery        = fmt.Errorf("root Query struct doesn't have any methods")
	ErrFirstParamMustBeContext = fmt.Errorf("the first parameter must be context.Context")
	ErrInvalidReturnSignature  = fmt.Errorf("the return signature should be (something, error)")
	ErrSecondReturnMustBeError = fmt.Errorf("the second return value must be error")
)

// InputDependency is a wrapper of dependencies derived from function parameters
// It will generate `input MyTypeInput {}` instead of `type MyType {...}` and mark all the dependencies underlying this depedency as InputDependency
type InputDependency struct {
	inner Dependency
}

func (d *InputDependency) String() string {
	return fmt.Sprintf("(input)%s", d.inner)
}

func (d *InputDependency) IsCustomType() bool {
	return d.inner.IsCustomType()
}

func (d *InputDependency) ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error) {
	obj, deps, err := d.inner.ToGraphQLObject(helper)
	var ideps []Dependency
	if err != nil {
		return nil, nil, err
	}
	if obj.ObjectType == GraphQLObjectTypeType {
		obj.ObjectType = GraphQLObjectTypeInput
		obj.Name = fmt.Sprintf("%sInput", obj.Name)
		for i := range obj.Fields {
			if obj.Fields[i].IsCustomType {
				obj.Fields[i].Type = fmt.Sprintf("%sInput", obj.Fields[i].Type)
			}
		}
	}
	for _, dep := range deps {
		if id, ok := dep.(*InputDependency); ok {
			ideps = append(ideps, id)
		} else {
			ideps = append(ideps, &InputDependency{inner: dep})
		}
	}
	return obj, ideps, nil
}

// RootDependency is an implementation of a special struct dependency for
//  `type Query struct {...}` or `type Mutation struct {...}`
type RootDependency struct {
	namedRef  *types.Named
	structRef *types.Struct
	depType   rootDependencyType
}

func (d *RootDependency) String() string {
	return d.namedRef.String()
}

func (d *RootDependency) IsCustomType() bool {
	return false
}

func (d *RootDependency) ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error) {
	var dependencies []Dependency
	var methods []GraphQLObjectMethod
	for i := 0; i < d.namedRef.NumMethods(); i++ {
		method := d.namedRef.Method(i)
		if !method.Exported() {
			continue
		}
		if err := d.validateMethodSignature(method, helper); err != nil {
			return nil, nil, fmt.Errorf("method error in %s: %w", method, err)
		}
		obj, deps, err := getGraphQLMethodFromFunc(method, helper)
		if err != nil {
			return nil, nil, fmt.Errorf("method error in %s: %w", method, err)
		}
		methods = append(methods, *obj)
		if len(deps) > 0 {
			for _, dep := range deps {
				dependencies = append(dependencies, dep)
			}
		}
	}
	if len(methods) == 0 {
		return nil, nil, ErrNoMethodsInQuery
	}
	objectType := GraphQLObjectTypeType
	var name string
	switch d.depType {
	case rootDependencyTypeMutation:
		name = RootMutationObjectName
		break
	case rootDependencyTypeQuery:
		name = RootQueryObjectName
		break
	}
	gqlObject := &GraphQLObject{
		Name:       name,
		GoModel:    d.namedRef.String(),
		ObjectType: objectType,
		Methods:    methods,
	}
	return gqlObject, dependencies, nil
}

func (d *RootDependency) validateMethodSignature(fun *types.Func, helper TypeHelper) error {
	signature := fun.Type().(*types.Signature)
	params := signature.Params()
	if params.Len() == 0 {
		return ErrFirstParamMustBeContext
	}
	if !helper.IsContext(params.At(0).Type()) {
		return ErrFirstParamMustBeContext
	}
	// returns
	results := signature.Results()
	if results.Len() < 1 {
		return ErrInvalidReturnSignature
	}
	if results.Len() == 2 {
		errReturn := results.At(1)
		if !helper.IsError(errReturn.Type()) {
			return ErrSecondReturnMustBeError
		}
	}
	return nil
}

// StructDependency is an implementation of Dependency of `type X struct {...}`
type StructDependency struct {
	namedRef  *types.Named
	structRef *types.Struct
}

func (d *StructDependency) String() string {
	return d.namedRef.String()
}

func (d *StructDependency) IsCustomType() bool {
	return true
}

// ToGraphQLObject returns a corresponding &GraphQLObject and extract new dependencies found in the types.
func (d *StructDependency) ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error) {
	var dependencies []Dependency
	fields, deps, err := d.getGraphQLObjectFields(helper)
	if err != nil {
		return nil, nil, err
	}
	if deps != nil {
		dependencies = append(dependencies, deps...)
	}
	methods, deps, err := d.getGraphQLObjectMethods(helper)
	if err != nil {
		return nil, nil, err
	}
	if deps != nil {
		dependencies = append(dependencies, deps...)
	}
	gqlObject := &GraphQLObject{
		Name:       d.namedRef.Obj().Name(),
		GoModel:    d.namedRef.String(),
		ObjectType: GraphQLObjectTypeType,
		Fields:     fields,
		Methods:    methods,
	}
	return gqlObject, dependencies, nil
}

func (d *StructDependency) getGraphQLObjectFields(helper TypeHelper) ([]GraphQLObjectField, []Dependency, error) {
	var dependencies []Dependency
	var fields []GraphQLObjectField
	for i := 0; i < d.structRef.NumFields(); i++ {
		field := d.structRef.Field(i)
		if !field.Exported() {
			continue
		}
		tagValues, err := hh.ParseFieldTag("graphql-schema", d.structRef.Tag(i))
		if err != nil {
			return nil, nil, fmt.Errorf("field error in %s: %w", field, err)
		}
		obj, dep, err := getGraphQLObjectFromField(field, helper)
		if err != nil {
			return nil, nil, fmt.Errorf("field error in %s: %w", field, err)
		}
		obj = applyTag(tagValues[0], obj)
		if obj != nil {
			fields = append(fields, *obj)
			if dep != nil {
				dependencies = append(dependencies, dep)
			}
		}
	}
	return fields, dependencies, nil
}

func (d *StructDependency) getGraphQLObjectMethods(helper TypeHelper) ([]GraphQLObjectMethod, []Dependency, error) {
	var dependencies []Dependency
	var methods []GraphQLObjectMethod
	for i := 0; i < d.namedRef.NumMethods(); i++ {
		method := d.namedRef.Method(i)
		if !method.Exported() {
			continue
		}
		obj, deps, err := getGraphQLMethodFromFunc(method, helper)
		if err != nil {
			return nil, nil, fmt.Errorf("method error in %s: %w", method, err)
		}
		methods = append(methods, *obj)
		if len(deps) > 0 {
			for _, dep := range deps {
				dependencies = append(dependencies, dep)
			}
		}
	}
	return methods, dependencies, nil
}

// InterfaceDependency is an implementation of Dependency of `type X interface {...}`
type InterfaceDependency struct {
	namedRef     *types.Named
	interfaceRef *types.Interface
}

func (d *InterfaceDependency) String() string {
	return d.namedRef.String()
}

func (d *InterfaceDependency) IsCustomType() bool {
	return true
}

// ToGraphQLObject returns a corresponding &GraphQLObject and extract new dependencies found in the types.
func (d *InterfaceDependency) ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error) {
	var dependencies []Dependency
	var fields []GraphQLObjectField
	var methods []GraphQLObjectMethod
	for i := 0; i < d.interfaceRef.NumMethods(); i++ {
		method := d.interfaceRef.Method(i)
		// log.Println("IF method", method)
		if !method.Exported() {
			continue
		}
		obj, deps, err := getGraphQLMethodFromFunc(method, helper)
		if err != nil {
			return nil, nil, fmt.Errorf("method error in %s: %w", method, err)
		}
		methods = append(methods, *obj)
		if len(deps) > 0 {
			for _, dep := range deps {
				dependencies = append(dependencies, dep)
			}
		}
	}
	objectType := GraphQLObjectTypeInterface
	gqlObject := &GraphQLObject{
		Name:       d.namedRef.Obj().Name(),
		GoModel:    d.namedRef.String(),
		ObjectType: objectType,
		Fields:     fields,
		Methods:    methods,
	}
	return gqlObject, dependencies, nil
}

// ScalarDependency is an implementation of Dependency of `type X BasicType` where BasicType is one of scalar types.
// This model genereats:
//
//    scalar X @goModel(model: "path.to.package.X")
//
type ScalarDependency struct {
	scalerType types.Type
}

func (d *ScalarDependency) String() string {
	return d.scalerType.String()
}

func (d *ScalarDependency) IsCustomType() bool {
	return false
}

func (d *ScalarDependency) ToGraphQLObject(helper TypeHelper) (*GraphQLObject, []Dependency, error) {
	// d.scalarType should be a type supporte by built in or named type that implements MarshalGQL() and UnmarshalGQL()
	t := d.scalerType.String()
	if name, ok := buildInTypeMaps[t]; ok {
		return &GraphQLObject{
			Name:       name,
			ObjectType: GraphQLObjectTypeScalar,
		}, nil, nil
	}
	named, ok := d.scalerType.(*types.Named)
	if !ok {
		return nil, nil, fmt.Errorf("unnamed scalar type: %s", t)
	}
	enumType := enum.GetEnum(named)
	if len(enumType.Keys) > 0 {
		var keys []string
		for _, k := range enumType.Keys {
			keys = append(keys, k.Name)
		}
		return &GraphQLObject{
			Name:       named.Obj().Name(),
			GoModel:    named.String(),
			Values:     keys,
			ObjectType: GraphQLObjectTypeEnum,
		}, nil, nil
	}
	return &GraphQLObject{
		Name:       named.Obj().Name(),
		GoModel:    named.String(),
		ObjectType: GraphQLObjectTypeScalar,
	}, nil, nil
}

// getGraphQLObjectFromField parses the field type and compose GraphQLObjectField.
// It also returns *Dependency if the filed depends on other type.
func getGraphQLObjectFromField(field *types.Var, helper TypeHelper) (*GraphQLObjectField, Dependency, error) {
	fieldType, nullable, isArray, elementNullable, nestDepth := normalizeFieldType(field.Type())
	resolver, err := resolveResolver(fieldType)
	if err != nil {
		return nil, nil, fmt.Errorf("field error in %q: %w", field.Name(), err)
	}
	t, dep, err := resolver(fieldType)
	if err != nil {
		return nil, nil, fmt.Errorf("field error in %q: %w", field.Name(), err)
	}
	if t == BasicTypeString && field.Name() == "ID" {
		t = BasicTypeID
	}
	return &GraphQLObjectField{
		Name:            hh.ToLowerCamleCase(field.Name()),
		Type:            t,
		Nullable:        nullable,
		IsArray:         isArray,
		ElementNullable: elementNullable,
		NestDepth:       nestDepth,
		IsCustomType:    dep != nil && dep.IsCustomType(),
	}, dep, nil
}

func getGraphQLMethodFromFunc(fun *types.Func, helper TypeHelper) (*GraphQLObjectMethod, []Dependency, error) {
	var arguments []GraphQLObjectField
	var dependencies []Dependency
	signature := fun.Type().(*types.Signature)
	results := signature.Results()
	if results.Len() < 1 {
		return nil, nil, ErrInvalidReturnSignature
	}
	if results.Len() == 2 {
		errReturn := results.At(1)
		if !helper.IsError(errReturn.Type()) {
			return nil, nil, ErrSecondReturnMustBeError
		}
	}
	returnValue, dep, err := getGraphQLObjectFromField(results.At(0), helper)
	if err != nil {
		return nil, nil, err
	}
	if dep != nil {
		dependencies = append(dependencies, dep)
	}
	params := signature.Params()
	if params.Len() == 0 {
		if returnValue.Type == BasicTypeString && fun.Name() == "ID" {
			returnValue.Type = BasicTypeID
		}
		return &GraphQLObjectMethod{
			Name:        hh.ToLowerCamleCase(fun.Name()),
			Parameters:  arguments,
			ReturnValue: *returnValue,
		}, dependencies, nil
	}
	var startIdx = 0
	var firstParam = params.At(0)
	if helper.IsContext(firstParam.Type()) {
		startIdx = 1
	}
	for i := startIdx; i < params.Len(); i++ {
		obj, dep, err := getGraphQLObjectFromField(params.At(i), helper)
		if err != nil {
			return nil, nil, err
		}
		if obj.Name == "" {
			obj.Name = fmt.Sprintf("param%d", i)
		}
		obj.Name = params.At(i).Name() // use the original name
		if obj.IsCustomType {
			obj.Type = fmt.Sprintf("%sInput", obj.Type)
		}
		arguments = append(arguments, *obj)
		if dep != nil {
			dependencies = append(dependencies, &InputDependency{inner: dep})
		}
	}
	if returnValue.Type == BasicTypeString && fun.Name() == "ID" {
		returnValue.Type = BasicTypeID
	}
	return &GraphQLObjectMethod{
		Name:        hh.ToLowerCamleCase(fun.Name()),
		Parameters:  arguments,
		ReturnValue: *returnValue,
	}, dependencies, nil
}

func normalizeFieldType(t types.Type) (tt types.Type, nullable bool, isArray bool, elementNullable bool, nestDepth int) {
	tt = t
	nullable = false
	isArray = false
	elementNullable = false
	nestDepth = 0
	switch t.(type) {
	case *types.Pointer:
		tt = t.(*types.Pointer).Elem()
		nullable = true
		isArray = false
		elementNullable = false
		return
	case *types.Slice:
		isArray = true
		tt, elementNullable, _, _, nestDepth = normalizeFieldType(t.(*types.Slice).Elem())
		return tt, true, isArray, elementNullable, nestDepth + 1
		// go array is not supported for resolver binding so don't normalize here.
		// if it gets suppored, use the following logic.
		// case *types.Array:
		// 	isArray = true
		// 	tt, elementNullable, _, _, nestDepth = normalizeFieldType(t.(*types.Array).Elem())
		// 	return tt, true, isArray, elementNullable, nestDepth + 1
	}
	return
}

var (
	typeContext *types.Interface
)
