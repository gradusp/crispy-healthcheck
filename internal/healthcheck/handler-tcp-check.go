package healthcheck

import (
	"context"
	"net"
	"time"

	"github.com/gradusp/crispy-healthcheck/internal/pkg/network/tcp"
	srvDef "github.com/gradusp/crispy-healthcheck/pkg/healthcheck"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (hcImpl *healthCheckImpl) TcpCheck(ctx context.Context, req *srvDef.TcpCheckRequest) (*srvDef.HealthCheckResponse, error) { ////nolint:revive
	const api = "API.HealthCheck.TcpCheck"

	var err error
	var tmo time.Duration
	addr := req.GetAddressToCheck()
	mark := req.GetSocketMark()

	if dur := req.GetTimeout(); dur != nil {
		err = dur.CheckValid()
		if err != nil {
			e := status.Error(codes.InvalidArgument, "invalid 'timeout' is provided from request")
			return nil, errors.Wrapf(e, "%s: %s", api, err.Error())
		}
		tmo = dur.AsDuration()
	}
	d := tcp.NewDialer(tcp.OptSocketMark{SocketMark: int(mark)},
		tcp.OptSocketConnectTimeout{Timeout: tmo})
	var conn net.Conn
	conn, err = d.DialContext(ctx, "tcp", addr)
	ret := new(srvDef.HealthCheckResponse)
	if err == nil {
		_ = conn.Close()
		ret.IsOk = true
	} else {
		var exp tcp.ErrUnableConnect
		if errors.As(err, &exp) {
			err = nil
		} else if errors.As(err, context.DeadlineExceeded) || errors.As(err, context.Canceled) {
			err = status.FromContextError(err).Err()
		}
	}
	return ret, errors.Wrap(err, api)
}
