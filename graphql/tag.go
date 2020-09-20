package graphql

const (
	structTagNothing  = ""
	structTagNoExport = "-"
)

func applyTag(t string, obj *GraphQLObjectField) *GraphQLObjectField {
	switch t {
	case structTagNothing:
		return obj
	case structTagNoExport:
		return nil
	default:
		obj.Name = string(t)
		return obj
	}
}
