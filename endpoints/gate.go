package endpoints

import (
	"context"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	dbr "github.com/gocraft/dbr"
	bgo "github.com/pickjunk/bgo"
	bcrypt "golang.org/x/crypto/bcrypt"
)

type resolver struct{}

var gate = bgo.NewGraphql(&resolver{})

func init() {
	gate.MergeSchema(`
	schema {
		mutation: Mutation
	}

	type Mutation {
		login(name: String!, passwd: String!): Boolean!
	}
	`)
}

func (r *resolver) Login(
	ctx context.Context,
	args struct {
		Name   string
		Passwd string
	},
) (bool, error) {
	db := ctx.Value(bgo.CtxKey("dbr")).(*dbr.Session)

	var admin struct {
		ID     string
		Name   string
		Passwd string
	}
	err := db.Select("*").
		Where(dbr.Eq("name", args.Name)).
		LoadOne(&admin)
	if err != nil {
		return false, bgo.Error(100, "name or password error")
	}

	err = bcrypt.CompareHashAndPassword([]byte(args.Passwd), []byte(admin.Passwd))
	if err != nil {
		return false, bgo.Error(100, "name or password error")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   admin.ID,
		"name": admin.Name,
		"exp":  time.Now().Add(time.Minute * 30).Unix(),
	})

	secret := bgo.Config["secret"].(string)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		bgo.Log.Panic(err)
	}

	cookie := http.Cookie{Name: "token", Value: tokenStr, Path: "/"}
	h := ctx.Value(bgo.CtxKey("http")).(*bgo.HTTP)
	http.SetCookie(h.Response, &cookie)

	return true, nil
}

// MountGate endpoint
func MountGate(r *bgo.Router) {
	r.Graphql("/gate", gate)
}
