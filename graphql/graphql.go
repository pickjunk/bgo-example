package graphql

import (
	bgo "github.com/pickjunk/bgo"
)

type resolver struct{}

// Graphql instance
var Graphql = bgo.NewGraphql(&resolver{})

func init() {
	Graphql.MergeSchema(`
	schema {
		query: Query
	}
	`)
}
