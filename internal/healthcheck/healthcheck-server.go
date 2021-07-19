package healthcheck

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/go-openapi/spec"
	srvDef "github.com/gradusp/crispy-healthcheck/pkg/healthcheck"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

//var _ app.ServiceDef = (*healthCheckImpl)(nil)
var _ srvDef.HealthCheckerServer = (*healthCheckImpl)(nil)

//go:embed healthchecker.swagger.json
var healthCheckSwagger []byte

type healthCheckImpl struct {
	srvDef.UnimplementedHealthCheckerServer
	appCtx context.Context //nolint:structcheck,unused
}

//RegisterGRPC registers GRPC server
func (hcImpl *healthCheckImpl) RegisterGRPC(_ context.Context, srv *grpc.Server) error {
	srvDef.RegisterHealthCheckerServer(srv, hcImpl)
	return nil
}

//RegisterGateway registers GRPC-GW Mux
func (hcImpl *healthCheckImpl) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	const api = "RegisterGateway"
	return errors.Wrap(srvDef.RegisterHealthCheckerHandlerServer(ctx, mux, hcImpl), api)
}

//GetHealthCheckSwagger возвращает spec.Swagger для сервиса Health-checker
func GetHealthCheckSwagger(hostName string) (*spec.Swagger, error) {
	const api = "GetHealthCheckSwagger"
	ret := new(spec.Swagger)
	if err := json.Unmarshal(healthCheckSwagger, ret); err != nil {
		return nil, errors.Wrap(err, api)
	}
	if len(hostName) > 0 {
		ret.Host = hostName
	}
	return ret, nil
}
