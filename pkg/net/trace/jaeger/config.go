package jaeger

import (
	"github.com/itering/kratos/pkg/conf/env"
	"github.com/itering/kratos/pkg/net/trace"
)

func Init() {
	c := &Config{Endpoint: "http://127.0.0.1:9191", BatchSize: 120}
	trace.SetGlobalTracer(trace.NewTracer(env.AppID, newReport(c), true))
}
