package tracer

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	ctxSpanKey = "tracing-span"
)

// Jaeger middleware
func Jaeger(tracer opentracing.Tracer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var span opentracing.Span
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		operation := method + " " + path

		if ctxSpan, ok := ctx.Get(ctxSpanKey); ok {
			span = startSpanWithParent(ctxSpan.(opentracing.Span).Context(), operation, method, path)
		} else {
			span = startSpanWithHeader(tracer, &ctx.Request.Header, operation, method, path)
		}

		ctx.Set(ctxSpanKey, span)
		defer span.Finish()

		ctx.Next()

		span.SetTag("http.url", ctx.Request.URL.Path)
		span.SetTag("http.method", ctx.Request.Method)
		span.SetTag("http.status_code", ctx.Writer.Status())
		span.SetTag("requestId", ctx.Value("requestId"))
		span.SetTag("uid", ctx.Value("uid"))
		span.SetTag("current-goroutines", runtime.NumGoroutine())
	}
}

// JaegerDefault default jaeger middleware
func JaegerDefault() gin.HandlerFunc {
	return Jaeger(Default.Tracer())
}

func startSpanWithParent(parent opentracing.SpanContext, operationName, method, path string) opentracing.Span {
	options := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value},
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: path},
	}

	if parent != nil {
		options = append(options, opentracing.ChildOf(parent))
	}

	return opentracing.StartSpan(operationName, options...)
}

func startSpanWithHeader(tracer opentracing.Tracer, header *http.Header, operationName, method, path string) opentracing.Span {
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(*header))
	span := tracer.StartSpan(operationName, ext.RPCServerOption(spanCtx))

	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(*header),
	)

	return span
}
