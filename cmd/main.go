package main

import (
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
		}
	}()

	cmd := flag.Arg(0)
	switch cmd {
	case "scan":
		waitTime := time.Second * 6
		if s := flag.Arg(1); s != "" {
			waitTime, err = time.ParseDuration(s)
			if err != nil {
				log.Fatal("Failed to parse duration")
			}
		}
		fmt.Println("Devices:")
		if devices, err := xb.ActiveScan(waitTime); err != nil {
			log.Fatal(err)
		} else {
			for _, d := range devices {
				fmt.Printf("\t%+v\n", d)
			}
		}
	case "discover":
		waitTime := time.Second * 6
		if s := flag.Arg(1); s != "" {
			waitTime, err = time.ParseDuration(s)
			if err != nil {
				log.Fatal("Failed to parse duration")
			}
		}
		fmt.Println("Nodes:")
		if nodes, err := xb.NodeDiscover(waitTime); err != nil {
			log.Fatal(err)
		} else {
			for _, n := range nodes {
				fmt.Printf("\t%+v\n", n)
			}
		}
	case "info":
		if escaped, err := xb.APIEnabled(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Escaped: %t\n", escaped)
		}
		if serial, err := xb.SerialNumber(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Serial number: %016x\n", serial)
		}
		if ni, err := xb.NodeIdentifier(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Node identifier: %s\n", ni)
		}
		if v, err := xb.FirmwareVersion(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Firmware version: %04x\n", v)
		}
		if v, err := xb.HardwareVersion(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Hardware version: %04x\n", v)
		}
		if v, err := xb.AssociationIndication(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Association indication: %d\n", v)
		}
		if v, err := xb.ExtendedPANID(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Extended PAN ID: %016x\n", v)
		}
		if v, err := xb.OperatingExtendedPANID(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Operating extended PAN ID: %016x\n", v)
		}
		if v, err := xb.EncryptionEnabled(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Encryption enabled: %t\n", v)
		}
		if v, err := xb.EncryptionOptions(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Encryption options: %s\n", v)
		}
		if v, err := xb.MaximumRFPayloadBytes(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Maximum RF payload bytes: %d\n", v)
		}
		if v, err := xb.NodeDiscoveryTimeout(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Node discovery timeout: %s\n", v)
		}
		if v, err := xb.NodeDiscoveryOptions(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Node discovery options: %s\n", v)
		}
	}
}
