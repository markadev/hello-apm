package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/spf13/pflag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var StatsdClient statsd.ClientInterface

type optionValues struct {
	serviceName string
	statsdAddr  string
	ecsHost     bool
	jobMode     bool
}

func main() {
	opts := getOptions()
	initECSHost(opts)
	StatsdClient = initStatsdClient(opts)

	tracer.Start(tracer.WithService(opts.serviceName))
	defer tracer.Stop()

	ctx := context.Background()

	if opts.jobMode {
		fakeWebRequest(ctx)
		return
	}

	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		go fakeWebRequest(ctx)
	}
}

func getOptions() (opts optionValues) {
	pflag.StringVar(&opts.serviceName, "service", "hello-apm", "service name to set in the tracer")
	pflag.StringVar(&opts.statsdAddr, "statsd-addr", "localhost:8125", "<host>:<port> of the statsd server")
	pflag.BoolVar(&opts.ecsHost, "ecs-host", false, "set DD_AGENT_HOST from the ECS instance metadata")
	pflag.BoolVar(&opts.jobMode, "job", false, "run as a one-shot job instead of looping forever")
	pflag.Parse()
	return
}

func initECSHost(opts optionValues) {
	if opts.ecsHost {
		resp, err := http.Get("http://169.254.169.254/latest/meta-data/local-ipv4")
		if err != nil {
			log.Panicf("failed to get ECS hostname: %v", err)
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		host := string(bodyBytes)
		if err != nil {
			log.Panicf("failed to get ECS hostname: %v", err)
		}

		os.Setenv("DD_AGENT_HOST", host)
	}
}

func initStatsdClient(opts optionValues) statsd.ClientInterface {
	tags := []string{
		"service:" + opts.serviceName,
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
	span, _ := tracer.StartSpanFromContext(ctx,
		"cache.request",
		tracer.ResourceName(key),
		tracer.SpanType(ext.SpanTypeRedis))
	defer span.Finish()

	time.Sleep(20 * time.Millisecond)
}

func fakeTemplateRender(ctx context.Context, name string) {
	span, _ := tracer.StartSpanFromContext(ctx,
		"template.render",
		tracer.ResourceName(name))
	defer span.Finish()

	time.Sleep(100 * time.Millisecond)
}
