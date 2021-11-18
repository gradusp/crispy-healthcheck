module github.com/gradusp/crispy-healthcheck

go 1.16

require (
	github.com/gradusp/go-platform v0.0.4-dev
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.0.0-RC3
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.0.0-RC3
	go.opentelemetry.io/otel/sdk v1.0.0-RC3
	go.opentelemetry.io/otel/trace v1.0.0-RC3
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
)

//replace github.com/gradusp/go-platform v0.0.1-dev => ./../go-platform
