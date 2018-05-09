package xbee

import (
	"io"

	"github.com/jacobsa/go-serial/serial"
)

func OpenPort(dev string, baud int) (io.ReadWriteCloser, error) {
	return serial.Open(serial.OpenOptions{
		PortName:        dev,
		BaudRate:        uint(baud),
		DataBits:        8,
		StopBits:        1,
		ParityMode:      serial.PARITY_NONE,
		MinimumReadSize: 1,
	})
}
