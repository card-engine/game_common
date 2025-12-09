package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitLightweightTracer 初始化轻量 TracerProvider（无 exporter，仅生成 traceID）
// serviceName: 服务标识（如 "apiserver" 或 "gameserver"）
func InitLightweightTracer(serviceName string, env string) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // 服务名，用于链路区分
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return nil, err
	}

	// 无 exporter，仅在内存中生成 traceID 和 span
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 全量采样
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}
