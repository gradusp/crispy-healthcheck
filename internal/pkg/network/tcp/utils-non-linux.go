//+build !linux

package tcp

//SetSocketMark ...
func (utils sockUtils) SetSocketMark(fd, mark int) error {
	_, _ = fd, mark
	return nil
}
