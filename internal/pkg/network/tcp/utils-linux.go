//+build linux

package tcp

import (
	"os"
	"syscall"
)

//SetSocketMark ...
func (utils dialer.sockUtils) SetSocketMark(fd, mark int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_MARK, mark); err != nil {
		return os.NewSyscallError("failed to set mark", err)
	}
	return nil
}
