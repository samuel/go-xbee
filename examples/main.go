package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/samuel/go-xbee/xbee"
)

var (
	flagBaud   = flag.Int("b", 115200, "Baud rate")
	flagDevice = flag.String("d", "", "Device path (e.g. /dev/ttyUSB0")
)

func main() {
	flag.Parse()

	port, err := xbee.OpenPort(*flagDevice, *flagBaud)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	xb, err := xbee.Open(port)
	if err != nil {
		log.Fatal(err)
	}
	defer xb.Close()

	go func() {
		ch := xb.EventChan()
		for ev := range ch {
			fmt.Printf("%+v\n", ev)
			if rx, ok := ev.(*xbee.ReceivePacket); ok {
				if bytes.Equal(rx.Data, []byte("ping")) {
					// TODO
				}
			}
		}
	}()

	cmd := flag.Arg(0)
	switch cmd {
	case "server":
		// Do nothing for this
		time.Sleep(time.Hour)
	case "client":
		for {
			if err := xb.Transmit(xbee.AddressCoordinator, xbee.Address16Unknown, 0, 0, []byte("ping")); err != nil {
				log.Println(err)
			}
			time.Sleep(time.Second * 5)
		}
	}
}
