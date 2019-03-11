package endpoints

import (
	bgo "github.com/pickjunk/bgo"
	b "github.com/pickjunk/bgo-example/business"
)

// MountAdmin endpoint
func MountAdmin(r *bgo.Router) {
	r.Graphql("/admin", b.Graphql)
}
