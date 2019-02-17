package graphql

import (
	"context"

	dbr "github.com/gocraft/dbr"
	graphql "github.com/graph-gophers/graphql-go"
	bgo "github.com/pickjunk/bgo"
)

func init() {
	Graphql.MergeSchema(`
		type Query {
			user(id: ID!): User
		}

		type User {
			id: ID
			name: String
		}
	`)
}

type user struct {
	ID   *string
	Name *string
}

type ur struct {
	u *user
}

func (r *resolver) User(
	ctx context.Context,
	args struct{ ID graphql.ID },
) *ur {
	db := ctx.Value(bgo.CtxKey("dbr")).(*dbr.Session)

	id := string(args.ID)
	name := "example"
	var t struct{}
	err := db.Select(`"`+id+`"`).LoadOneContext(ctx, &t)
	if err != nil {
		bgo.Log.Panic(err)
	}

	return &ur{
		u: &user{
			ID:   &id,
			Name: &name,
		},
	}
}

// ID field
func (r *ur) ID() *graphql.ID {
	id := graphql.ID(*r.u.ID)
	return &id
}

// Name field
func (r *ur) Name(ctx context.Context) *string {
	db := ctx.Value(bgo.CtxKey("dbr")).(*dbr.Session)
	var t struct{}
	err := db.Select(`"`+*r.u.Name+`"`).LoadOneContext(ctx, &t)
	if err != nil {
		bgo.Log.Panic(err)
	}

	return r.u.Name
}
