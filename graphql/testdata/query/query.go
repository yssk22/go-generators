package query

import "context"

type QueryWithoutMethods struct{}

type QueryWithInvalidMethodWithoutContext struct{}

func (*QueryWithInvalidMethodWithoutContext) Foo() {}

type QueryWithInvalidReturnSignatureMethod struct{}

func (*QueryWithInvalidReturnSignatureMethod) Foo(context.Context) {}

type QueryWithInvalidErrorReturnSignatureMethod struct{}

func (*QueryWithInvalidErrorReturnSignatureMethod) Foo(context.Context) (string, string) {
	return "", ""
}

type QueryWithUnsupportedType struct{}

func (*QueryWithUnsupportedType) Foo(ctx context.Context, f func()) (string, error) {
	return "", nil
}

type QueryWithUnsupportedTypeFloat struct{}

// float32 is not supported, use float64
func (*QueryWithUnsupportedTypeFloat) Foo(ctx context.Context, f float32) (string, error) {
	return "", nil
}

type QueryWithUnsupportedTypeMap struct{}

// map[string]string is not supported, use map[string]interface{}
func (*QueryWithUnsupportedTypeMap) Foo(ctx context.Context, m map[string]string) (string, error) {
	return "", nil
}

type QueryWithUnsupportedTypeBasic struct{}

// complex64 is not supported, only int, bool, float64, and string is supported
func (*QueryWithUnsupportedTypeBasic) Foo(ctx context.Context, c complex64) (string, error) {
	return "", nil
}

type QueryWithUnsupportedTypePointerOfPointer struct{}

func (*QueryWithUnsupportedTypePointerOfPointer) Foo(ctx context.Context, c **string) (string, error) {
	return "", nil
}

type QueryWithSupportedFields struct{}

func (*QueryWithSupportedFields) Foo(ctx context.Context) (*SupportedFieldsStruct, error) {
	return nil, nil
}

type MyStruct struct{}

type SupportedFieldsStruct struct {
	FieldString               string
	FieldNullableSrinrg       *string
	FieldInt                  int
	FieldNullableInt          *int
	FieldFloat                float64
	FieldNullableFloat        *float64
	FieldBoolean              bool
	FieldNullableBoolean      *bool
	FieldMap                  map[string]interface{}
	FieldStruct               MyStruct
	FieldStructNullable       *MyStruct
	FieldArray                []string
	FieldArrayElementNullable []*string
	// Go Array is not supported for resolver binding so omit
	// FieldGoArray                [3]string
	// FieldGoArrayElementNullable [3]*string
	FieldArrayOfArray [][]string
}
