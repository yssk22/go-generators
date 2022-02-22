package graphql

const (
	RootQueryObjectName    = "Query"
	RootMutationObjectName = "Mutation"
)

const (
	BasicTypeID      = "ID"
	BasicTypeString  = "String"
	BasicTypeInteger = "Int"
	BasicTypeFloat   = "Float"
	BasicTypeBoolean = "Boolean"

	BuiltInTypeAny  = "Any"
	BuiltInTypeMap  = "Map"
	BuiltInTypeTime = "Time"
)

var builtInTypes = []string{
	BasicTypeID, BasicTypeString, BasicTypeInteger, BasicTypeFloat, BasicTypeBoolean, BuiltInTypeAny, BuiltInTypeMap, BuiltInTypeTime,
}

func IsCustomType(typeName string) bool {
	for _, t := range builtInTypes {
		if typeName == t {
			return false
		}
	}
	return true
}

type GraphQLObjectType string

const (
	GraphQLObjectTypeType      = GraphQLObjectType("type")
	GraphQLObjectTypeInterface = GraphQLObjectType("interface")
	GraphQLObjectTypeInput     = GraphQLObjectType("input")
	GraphQLObjectTypeScalar    = GraphQLObjectType("scalar")
	GraphQLObjectTypeEnum      = GraphQLObjectType("enum")
)

// GraphQLObject represents type Name {...}
type GraphQLObject struct {
	Name       string
	IsNode     bool
	GoModel    string
	ObjectType GraphQLObjectType // type, input, scalar, enum
	Fields     []GraphQLObjectField
	Methods    []GraphQLObjectMethod
	Values     []string // enum
}

// GraphQLObjectField represents a field in GraphQLObject
type GraphQLObjectField struct {
	Name            string
	Type            string
	Nullable        bool
	IsArray         bool
	ElementNullable bool
	NestDepth       int
	IsCustomType    bool
}

// GraphQLObjectMethod represents a method in GraphQLType
type GraphQLObjectMethod struct {
	Name        string
	Parameters  []GraphQLObjectField
	ReturnValue GraphQLObjectField
}
