package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	bgo "github.com/pickjunk/bgo"
)

var authRegexp = regexp.MustCompile(`^Bearer (.*)`)

// Admin middleware, check whether user is an admin
func Admin(ctx context.Context, next bgo.Handle) {
	h := ctx.Value(bgo.CtxKey("http")).(*bgo.HTTP)
	w := h.Response
	r := h.Request

	auth := r.Header.Get("Authorization")
	m := authRegexp.FindStringSubmatch(auth)
	if m == nil || m[1] == "" {
		bgo.Log.Debug("Authorization format error")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	secret := bgo.Config["secret"].(string)
	token, err := jwt.Parse(m[1], func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		bgo.Log.Debug("JWT parse error")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		bgo.Log.Debug("JWT is invalid")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx = context.WithValue(ctx, bgo.CtxKey("id"), claims["id"])

	// less than 10 minutes to expire, refresh token
	exp := claims["exp"].(float64)
	if time.Now().After(time.Unix(int64(exp), 0).Add(-10 * time.Minute)) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":   claims["id"],
			"name": claims["name"],
			"exp":  time.Now().Add(time.Minute * 30).Unix(),
		})

		secret := bgo.Config["secret"].(string)
		tokenStr, err := token.SignedString([]byte(secret))
		if err != nil {
			bgo.Log.Panic(err)
		}

		cookie := http.Cookie{Name: "token", Value: tokenStr, Path: "/"}
		http.SetCookie(w, &cookie)
	}

	next(ctx)
}
