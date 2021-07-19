package tcp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"
)

//Dialer makes TCP dialer
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

//NewDialer creates new dialer
func NewDialer(opts ...DialerOption) Dialer {
	ret := new(tcpDialerImpl)
	ret.tcpDialerOptions.fill(opts...)
	return ret
}

// ----------------------------------------- IMPL ---------------------------------------//
var _ Dialer = (*tcpDialerImpl)(nil)

type tcpDialerImpl struct {
	tcpDialerOptions
}

func (dialer *tcpDialerImpl) Dial(network, address string) (net.Conn, error) {
	return dialer.DialContext(context.Background(), network, address)
}

func (dialer *tcpDialerImpl) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	const api = "TCP.Dial"

	switch strings.ToLower(network) {
	case TCP, TCP4, TCP6:
		break
	default:
		return nil, errors.Wrap(net.UnknownNetworkError(network), api)
	}

	var (
		err      error
		socketD  int
		f        *os.File
		addrInfo tcpAddrInfo
	)
	defer func() {
		if socketD > 0 {
			_ = syscall.Close(socketD)
		}
		if f != nil {
			_ = f.Close()
		}
	}()
	if addrInfo, err = SockUtils.getTcpSocketInfo(address); err != nil {
		return nil, errors.Wrap(&net.AddrError{Err: err.Error(), Addr: address}, api)
	}
	if socketD, err = syscall.Socket(addrInfo.domain, syscall.SOCK_STREAM, 0); err != nil {
		return nil, errors.Wrap(os.NewSyscallError("socket", err), api)
	}
	syscall.CloseOnExec(socketD)
	options := dialer.tcpDialerOptions
	if options.rdSocketTmo > 0 {
		err = SockUtils.SetSocketRdTimeout(socketD, options.rdSocketTmo)
		if err != nil {
			return nil, errors.Wrap(err, api)
		}
	}
	if options.wrSocketTmo > 0 {
		err = SockUtils.SetSocketWrTimeout(socketD, options.wrSocketTmo)
		if err != nil {
			return nil, errors.Wrap(err, api)
		}
	}
	if options.sockMark != 0 {
		err = SockUtils.SetSocketMark(socketD, options.sockMark)
		if err != nil {
			return nil, errors.Wrap(err, api)
		}
	}

	chDone := make(chan error, 1)
	go func() {
		var e error
		defer close(chDone)
		sa := addrInfo.makeSocketAddress()
		for {
			if e = syscall.Connect(socketD, sa); e == nil {
				break
			}
			// Blocking socket connect may be interrupted with EINTR
			if e != syscall.EINTR {
				e = os.NewSyscallError("connect", e)
				break
			}
		}
		chDone <- e
	}()
	ctx1 := ctx
	if options.connectTmo > 0 {
		var c func()
		ctx1, c = context.WithTimeout(ctx1, options.connectTmo)
		defer c()
	}
	select {
	case <-ctx1.Done():
		return nil, errors.Wrap(ctx1.Err(), api)
	case err = <-chDone:
		if err != nil {
			return nil, errors.Wrap(ErrUnableConnect{Reason: err, Address: address}, api)
		}
	}
	var lsa, rsa syscall.Sockaddr
	if lsa, err = syscall.Getsockname(socketD); err != nil {
		return nil, errors.Wrap(err, api)
	}
	if rsa, err = syscall.Getpeername(socketD); err != nil {
		return nil, errors.Wrap(err, api)
	}
	name := fmt.Sprintf("%s %s -> %s", addrInfo.network,
		SockUtils.SockAddrStringer(lsa), SockUtils.SockAddrStringer(rsa))
	if f = os.NewFile(uintptr(socketD), name); f == nil {
		return nil, errors.Errorf("%s: unable os.NewFile from socket", api)
	}
	socketD = -1
	retConn := new(connWrapper)
	if retConn.Conn, err = net.FileConn(f); err != nil {
		return nil, errors.Wrap(err, api)
	}
	runtime.SetFinalizer(retConn, func(o *connWrapper) {
		_ = o.Close()
	})
	return retConn, nil
}
