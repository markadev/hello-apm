package main

import (
	"context"
	"os"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/spf13/pflag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var StatsdClient statsd.ClientInterface

type optionValues struct {
	tracerService string
	statsdAddr    string
}

func main() {
	opts := getOptions()
	StatsdClient = initStatsdClient(opts)

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
	pflag.StringVar(&opts.statsdAddr, "statsd-addr", "localhost:8125", "<host>:<port> of the statsd server")
	pflag.Parse()
	return
}

func initStatsdClient(opts optionValues) statsd.ClientInterface {
	tags := []string{
		"service:" + opts.tracerService,
	}
	if env, ok := os.LookupEnv("DD_ENV"); ok {
		tags = append(tags, "env:"+env)
	}

	client, err := statsd.New(opts.statsdAddr, statsd.WithTags(tags))
	if err != nil {
		panic(err)
	}
	return client
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
	StatsdClient.Count("hello_apm.requests", 1, nil, 1.0)
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
