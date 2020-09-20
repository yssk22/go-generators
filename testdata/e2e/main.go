package main

import (
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/yssk22/go-generators/testdata/e2e/gqlgen"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/graphql/", handler.Playground("GraphQL playground", "/query"))
	mux.Handle("/query", handler.GraphQL(gqlgen.NewExecutableSchema(
		gqlgen.Config{
			Resolvers: &gqlgen.Resolver{},
		},
	)))
	http.ListenAndServe(":8080", mux)
}
