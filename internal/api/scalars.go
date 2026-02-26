package api

// Numeric is a custom GraphQL scalar that maps to Hardcover's "numeric" type.
// The standard graphql.Float maps to "Float" which causes type mismatch errors.
type Numeric float64

// GetGraphQLType implements the go-graphql-client GraphQLType interface.
func (n Numeric) GetGraphQLType() string { return "numeric" }

// Date is a custom GraphQL scalar that maps to Hardcover's "date" type.
// The standard graphql.String maps to "String" which causes type mismatch errors.
type Date string

// GetGraphQLType implements the go-graphql-client GraphQLType interface.
func (d Date) GetGraphQLType() string { return "date" }
