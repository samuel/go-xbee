package xbee

import (
	"io"

	"github.com/samofly/serial"
)

func OpenPort(dev string, baud int) (io.ReadWriteCloser, error) {
	return serial.Open(dev, baud)
}
