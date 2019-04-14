package main

import (
	bgo "github.com/pickjunk/bgo"
	endpoints "github.com/pickjunk/bgo-example/endpoints"
	bgoDbr "github.com/pickjunk/bgo/dbr"
	config "github.com/uber/jaeger-client-go/config"
)

func main() {
	r := bgo.New()

	r.Swagger([]byte{})

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

	rWithDbr := r.Middlewares(bgoDbr.Middleware(nil))

	endpoints.MountGate(rWithDbr)
	endpoints.MountAdmin(rWithDbr)

	r.ListenAndServe()
}
