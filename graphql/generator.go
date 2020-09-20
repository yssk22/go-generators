package graphql

type Generator interface {
	Generate([]GraphQLObject) error
}

type Option func(*builder) *builder

func RootQueryName(str string) Option {
	return func(builder *builder) *builder {
		builder.rootQueryName = str
		return builder
	}
}

func Generate(dir string, g Generator, options ...Option) error {
	objectList, err := Build(dir, options...)
	if err != nil {
		return err
	}
	// output
	return g.Generate(objectList)
}
