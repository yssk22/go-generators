package graphql

import (
	"errors"
	"os"
	"testing"
)

func TestBuild(t *testing.T) {
	cases := []struct {
		dir       string
		queryName string
		err       error
	}{
		{
			dir: "testdata/notadir",
			err: os.ErrNotExist,
		},
		{
			dir: "testdata/invalid",
			err: ErrSyntax,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithSupportedFields",
			err:       nil,
		},
		{
			dir:       "testdata/query",
			queryName: "NonExistentQuery",
			err:       ErrNoQuery,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithoutMethods",
			err:       ErrNoMethodsInQuery,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithInvalidMethodWithoutContext",
			err:       ErrFirstParamMustBeContext,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithInvalidReturnSignatureMethod",
			err:       ErrInvalidReturnSignature,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithInvalidErrorReturnSignatureMethod",
			err:       ErrSecondReturnMustBeError,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithUnsupportedType",
			err:       ErrUnsupportedType,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithUnsupportedTypeFloat",
			err:       ErrUnsupportedType,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithUnsupportedTypeMap",
			err:       ErrUnsupportedType,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithUnsupportedTypeBasic",
			err:       ErrUnsupportedType,
		},
		{
			dir:       "testdata/query",
			queryName: "QueryWithUnsupportedTypePointerOfPointer",
			err:       ErrUnsupportedType,
		},
	}
	for _, c := range cases {
		t.Run(c.queryName, func(tt *testing.T) {
			_, err := Build(c.dir, RootQueryName(c.queryName))
			if !errors.Is(err, c.err) {
				tt.Errorf("expected: %s, got: %s", c.err, err)
			}
		})
	}
}
