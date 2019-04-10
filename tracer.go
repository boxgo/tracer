package tracer

import (
	"context"
	"io"
	"time"

	"github.com/boxgo/box/minibox"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

type (
	// Tracer godoc
	Tracer struct {
		name                           string
		app                            minibox.App
		Disabled                       bool          `config:"disabled"`
		RPCMetrics                     bool          `config:"rpc_metrics"`
		SamplerType                    string        `config:"samplerType"`
		SamplerParam                   float64       `config:"samplerParam"`
		SamplerSamplingServerURL       string        `config:"samplerSamplingServerURL"`
		SamplerMaxOperations           int           `config:"samplerMaxOperations"`
		SamplerSamplingRefreshInterval time.Duration `config:"samplerSamplingRefreshInterval"`
		ReporterQueueSize              int           `config:"reporterQueueSize"`
		ReporterBufferFlushInterval    time.Duration `config:"reporterBufferFlushInterval"`
		ReporterLogSpans               bool          `config:"reporterLogSpans"`
		ReporterLocalAgentHostPort     string        `config:"reporterLocalAgentHostPort"`
		ReporterCollectorEndpoint      string        `config:"reporterCollectorEndpoint"`
		ReporterUser                   string        `config:"reporterUser"`
		ReporterPassword               string        `config:"reporterPassword"`
		tracer                         opentracing.Tracer
		closer                         io.Closer
	}
)

var (
	// Default default tracer
	Default = New("tracer")
)

// Name config name
func (tracer *Tracer) Name() string {
	return tracer.name
}

// Exts get app info
func (tracer *Tracer) Exts() []minibox.MiniBox {
	return []minibox.MiniBox{&tracer.app}
}

// ConfigWillLoad before config load
func (tracer *Tracer) ConfigWillLoad(ctx context.Context) {

}

// ConfigDidLoad after config load
func (tracer *Tracer) ConfigDidLoad(ctx context.Context) {
	if tracer.SamplerType == "" {
		tracer.SamplerType = jaeger.SamplerTypeConst
	}
	if tracer.SamplerParam == 0 {
		tracer.SamplerParam = 1
	}

	jaegerCfg := &jaegerConfig.Configuration{
		ServiceName: tracer.app.AppName,
		Disabled:    tracer.Disabled,
		RPCMetrics:  tracer.RPCMetrics,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:                    tracer.SamplerType,
			Param:                   tracer.SamplerParam,
			SamplingServerURL:       tracer.SamplerSamplingServerURL,
			MaxOperations:           tracer.SamplerMaxOperations,
			SamplingRefreshInterval: tracer.SamplerSamplingRefreshInterval,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			QueueSize:           tracer.ReporterQueueSize,
			BufferFlushInterval: tracer.ReporterBufferFlushInterval,
			LogSpans:            tracer.ReporterLogSpans,
			LocalAgentHostPort:  tracer.ReporterLocalAgentHostPort,
			CollectorEndpoint:   tracer.ReporterCollectorEndpoint,
			User:                tracer.ReporterUser,
			Password:            tracer.ReporterPassword,
		},
	}

	t, c, err := jaegerCfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		panic("Tracer init error: " + err.Error())
	}

	tracer.tracer = t
	tracer.closer = c
}

func (tracer *Tracer) Tracer() opentracing.Tracer {
	return tracer.tracer
}

// Serve godoc
func (tracer *Tracer) Serve(ctx context.Context) error {
	return nil
}

// Shutdown close
func (tracer *Tracer) Shutdown(ctx context.Context) error {
	if tracer.closer != nil {
		return tracer.closer.Close()
	}

	return nil
}

// New new a tracer
func New(name string) *Tracer {
	return &Tracer{
		name: name,
	}
}
