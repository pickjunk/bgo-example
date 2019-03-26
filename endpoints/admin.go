package endpoints

import (
	bgo "github.com/pickjunk/bgo"
	b "github.com/pickjunk/bgo-example/business"
	m "github.com/pickjunk/bgo-example/middlewares"
)

// MountAdmin endpoint
func MountAdmin(r *bgo.Router) {
	r.Middlewares(m.Admin).Graphql("/admin", b.Graphql)
}
