package business

import (
	bgo "github.com/pickjunk/bgo"
)

type resolver struct{}

// Graphql endpoint
var Graphql = bgo.NewGraphql(&resolver{})

func init() {
	Graphql.MergeSchema(`
	schema {
		query: Query
		mutation: Mutation
	}
	`)
}
