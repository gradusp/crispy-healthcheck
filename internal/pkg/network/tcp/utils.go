package tcp

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

//SockUtils ...
var SockUtils sockUtils

type (
	sockUtils        struct{}
	sockAddrStringer struct {
		syscall.Sockaddr
	}
	tcpAddrInfo struct {
		network string
		domain  int
		ip      net.IP
		port    int
	}
	connWrapper struct {
		net.Conn
	}
)

const (
	TCP  = "tcp"  //nolint
	TCP4 = "tcp4" //nolint
	TCP6 = "tcp6" //nolint
)

//String ...
func (ss sockAddrStringer) String() string {
	switch sa := ss.Sockaddr.(type) {
	case *syscall.SockaddrInet4:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *syscall.SockaddrInet6:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *syscall.SockaddrUnix:
		return sa.Name
	case nil:
		return "<nil>"
	default:
		return fmt.Sprintf("(unsupported - %T)", sa)
	}
}

//SockAddrStringer ...
func (utils sockUtils) SockAddrStringer(s syscall.Sockaddr) fmt.Stringer {
	return sockAddrStringer{Sockaddr: s}
}

func (inf tcpAddrInfo) makeSocketAddress() syscall.Sockaddr {
	var ret syscall.Sockaddr
	switch inf.network {
	case TCP4:
		sa := &syscall.SockaddrInet4{Port: inf.port}
		copy(sa.Addr[:], inf.ip)
		ret = sa
	case TCP6:
		sa := &syscall.SockaddrInet6{Port: inf.port}
		copy(sa.Addr[:], inf.ip)
		ret = sa
	}
	return ret
}

func (utils sockUtils) getTcpSocketInfo(ipPortAddress string) (tcpAddrInfo, error) { //nolint:revive
	const api = "getTcpSocketInfo"
	var ret tcpAddrInfo

	host, port, err := net.SplitHostPort(ipPortAddress)
	if err != nil {
		return ret, errors.Wrap(err, api)
	}

	if ret.ip = net.ParseIP(host); ret.ip == nil {
		return ret, errors.Errorf("%s: provided host[%s] is not IP", api, host)
	}
	if ret.port, err = strconv.Atoi(port); err != nil {
		return ret, fmt.Errorf("invalid provided port[%q]", port)
	}
	if strings.Contains(host, ":") {
		ret.network = TCP6
		ret.domain = syscall.AF_INET6
		if ret.ip = ret.ip.To16(); ret.ip == nil {
			return ret, errors.Errorf("%s: invalid provided host[%s] as is not IPv6", api, host)
		}
	} else {
		ret.network = TCP4
		ret.domain = syscall.AF_INET
		if ret.ip = ret.ip.To4(); ret.ip == nil {
			return ret, errors.Errorf("%s: invalid provided host[%s] as is not IPv4", api, host)
		}
	}
	return ret, nil
}

//SetSocketRdTimeout ...
func (utils sockUtils) SetSocketRdTimeout(fd int, duration time.Duration) error {
	const api = "SetSockRdTimeout"
	tv := syscall.NsecToTimeval(duration.Nanoseconds())
	err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
	return errors.Wrap(os.NewSyscallError("SetsockoptTimeval", err), api)
}

//SetSocketWrTimeout ...
func (utils sockUtils) SetSocketWrTimeout(fd int, duration time.Duration) error {
	const api = "SetSocketWrTimeout"
	tv := syscall.NsecToTimeval(duration.Nanoseconds())
	err := syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &tv)
	return errors.Wrap(os.NewSyscallError("SetsockoptTimeval", err), api)
}
