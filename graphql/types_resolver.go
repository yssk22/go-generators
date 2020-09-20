package graphql

import (
	"fmt"
	"go/types"
)

var buildInTypeMaps = map[string]string{
	"time.Time":              BuiltInTypeTime,
	"map[string]interface{}": BuiltInTypeMap,
	"interface{}":            BuiltInTypeAny,
}

type typeResolver func(types.Type) (string, Dependency, error)

func namedResolver(t types.Type) (string, Dependency, error) {
	named := t.(*types.Named)
	if t, ok := buildInTypeMaps[named.String()]; ok {
		// return t, nil, nil
		return t, &ScalarDependency{scalerType: named}, nil
	}
	underlying := named.Underlying()
	switch underlying.(type) {
	case *types.Struct:
		return named.Obj().Name(), &StructDependency{
			namedRef:  named,
			structRef: underlying.(*types.Struct),
		}, nil
	case *types.Basic:
		return named.Obj().Name(), &ScalarDependency{scalerType: named}, nil
	case *types.Interface:
		return named.Obj().Name(), &InterfaceDependency{
			namedRef:     named,
			interfaceRef: underlying.(*types.Interface),
		}, nil
	}
	return "", nil, unsupportedError(t)
}

func basicResolver(t types.Type) (string, Dependency, error) {
	basic := t.(*types.Basic)
	// https://org/graphql-js/basic-types/
	// String, Int, Float, Boolean, and ID
	info := basic.Info()
	if info&types.IsString == types.IsString {
		return BasicTypeString, nil, nil
	}
	if info&types.IsInteger == types.IsInteger {
		return BasicTypeInteger, nil, nil
	}
	if info&types.IsFloat == types.IsFloat {
		if basic.Kind() != types.Float64 {
			return "", nil, unsupportedError(basic)
		}
		return BasicTypeFloat, nil, nil
	}
	if info&types.IsBoolean == types.IsBoolean {
		return BasicTypeBoolean, nil, nil
	}
	return "", nil, unsupportedError(t)
}

func mapResolver(t types.Type) (string, Dependency, error) {
	m := t.(*types.Map)
	if _, ok := m.Elem().(*types.Interface); !ok {
		return "", nil, unsupportedError(t)
	}
	return BuiltInTypeMap, &ScalarDependency{scalerType: m}, nil
}

func interfaceResolver(t types.Type) (string, Dependency, error) {
	return BuiltInTypeAny, &ScalarDependency{scalerType: t}, nil
}

func resolveResolver(t types.Type) (typeResolver, error) {
	switch t.(type) {
	case *types.Named:
		return namedResolver, nil
	case *types.Basic:
		return basicResolver, nil
	case *types.Map:
		return mapResolver, nil
	case *types.Interface:
		return interfaceResolver, nil
	}
	return nil, unsupportedError(t)
}

var (
	ErrUnsupportedType = fmt.Errorf("unsupported type")
)

func unsupportedError(t types.Type) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedType, t)
}
