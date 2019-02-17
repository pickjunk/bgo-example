package main

import (
	"net/http"

	dbr "github.com/gocraft/dbr"
	httprouter "github.com/julienschmidt/httprouter"
	bgo "github.com/pickjunk/bgo"
	g "github.com/pickjunk/bgo-example/graphql"
	bgoDbr "github.com/pickjunk/bgo/dbr"
	config "github.com/uber/jaeger-client-go/config"
)

func main() {
	closer := bgo.Jaeger(&config.Configuration{
		ServiceName: "bgo-example",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	})
	defer closer.Close()

	r := bgo.New()

	r.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("hello world!"))
	})

	rWithDbr := r.Middlewares(bgoDbr.Middleware(nil))

	rWithDbr.GET("/dbr", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		db := ctx.Value(bgo.CtxKey("dbr")).(*dbr.Session)

		var test struct{}
		err := db.Select(`"empty"`).LoadOneContext(ctx, &test)
		if err != nil {
			bgo.Log.Panic(err)
		}

		w.Write([]byte(`dbr: SELECT "empty"`))
	})

	rWithDbr.Graphql("/graphql", g.Graphql)

	r.ListenAndServe()
}
