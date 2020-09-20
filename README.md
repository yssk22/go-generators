# go-generators

## Generator Commands

### gen-graphql

`gen-graphql` generates a GraphSQL schema file and gqlgen generated files from Go implement. It introduce `code-first` GraphQL development with zero configurations.

### gen-enum-gqlgen

`gen-enum-gqlgen` generates mashaler/unmarshaler implemenation required to serve the GraphQL server on top of `gqlgen` which uses `Enum`.

### gen-enum-entgo

`gen-enum-entgo` generates `Values()` method required for a custome go enum on `entgo` framework
