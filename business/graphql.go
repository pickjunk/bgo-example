package business

import (
	bgo "github.com/pickjunk/bgo"
)

type resolver struct{}

// Graphql endpoint
var Graphql = bgo.NewGraphql(&resolver{})
