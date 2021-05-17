package main

import (
	"context"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type optionValues struct {
	tracerService string
}

func main() {
	opts := getOptions()

	tracer.Start(tracer.WithService(opts.tracerService))
	defer tracer.Stop()

	ctx := context.Background()
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		go fakeWebRequest(ctx)
	}
}

func getOptions() (opts optionValues) {
	pflag.StringVar(&opts.tracerService, "tracer-service", "hello-apm", "service name to set in the tracer")
	pflag.Parse()
	return
}

func fakeWebRequest(ctx context.Context) {
	span, ctx := tracer.StartSpanFromContext(ctx,
		"web.request",
		tracer.ResourceName("/hello"),
		tracer.SpanType(ext.SpanTypeWeb))
	defer span.Finish()

	fakeCacheRequest(ctx, "user_abc")
	fakeTemplateRender(ctx, "hello.tmpl")

	span.SetTag("http.status_code", "202")
}

func fakeCacheRequest(ctx context.Context, key string) {
	span, ctx := tracer.StartSpanFromContext(ctx,
		"cache.request",
		tracer.ResourceName(key),
		tracer.SpanType(ext.SpanTypeRedis))
	defer span.Finish()

	time.Sleep(20 * time.Millisecond)
}

func fakeTemplateRender(ctx context.Context, name string) {
	span, ctx := tracer.StartSpanFromContext(ctx,
		"template.render",
		tracer.ResourceName(name))
	defer span.Finish()

	time.Sleep(100 * time.Millisecond)
}
