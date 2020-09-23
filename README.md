# go-generators

## Generator Commands

### gen-graphql

`gen-graphql` generates a GraphSQL schema file and gqlgen generated files from Go implement. It introduces `code-first` GraphQL development with zero configurations. You can add the following line on your package where your root query and mutation are placed so that you'll get `generated` directory by `go generate`. The generated directory contains everything required for gqlgen so you don't have to write your `gqlgen.yml`, `schema.graphql`, or resolvers.

```
//go:generate go run github.com/yssk22/go-generators/cmd/gen-graphql ./ ./generated
```

See [example.go](https://github.com/yssk22/go-generators/blob/master/testdata/e2e/models/example.go) about how you can code GraphQL queries and mutations.

### gen-enum-gqlgen

`gen-enum-gqlgen` generates `MarshalGQL()` and `UnmarshalGQL()` implemenation required to serve the GraphQL server on top of `gqlgen` which uses `Enum`. You can add the following line on your package where your enums are placed so that you'll get `gqlgen_enums.go` by `go generate`.

```
//go:generate go run github.com/yssk22/go-generators/cmd/gen-gqlgen-enum ./
```

### gen-enum-entgo

`gen-enum-entgo` generates `Values()` implementation required for a custom go enum on [entgo framework](https://entgo.io/). You can add the following line on your package where your enums are placed so that you'll get `entgo_enums.go` by `go generate`.

```
//go:generate go run github.com/yssk22/go-generators/cmd/gen-entgo-enum ./
```
