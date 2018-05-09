package xbee

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	frameDelimiter             = 0x7e
	frameATCommand             = 0x08
	frameATCommandQueue        = 0x09
	frameZigBeeTransmitRequest = 0x10
	frameATCommandResponse     = 0x88
	frameModemStatus           = 0x8a
	frameZigBeeTransmitStatus  = 0x8b
	frameZigBeeReceivePacket   = 0x90
)

var (
	ErrInvalidParameter = errors.New("xbee: invalid parameter")
	ErrResponse         = errors.New("xbee: generic error response")
	ErrTXFailure        = errors.New("xbee: TX failure")
)

type ErrInvalidCommand string

func (e ErrInvalidCommand) Error() string {
	return fmt.Sprintf("xbee: invalid command %s", string(e))
}

type ActiveScanDevice struct {
	Type         byte // 2 - ZB firmware uses a different format than Wi-Fi XBee, which is type 1
	Channel      byte
	PAN          uint16
	ExtendedPAN  uint64
	AllowJoin    bool
	StackProfile byte
	LQI          byte // higher values are better
	RSSI         int8 // lower values are better
}

type Node struct {
	SerialNumber         uint64
	NodeID               string
	ParentNetworkAddress uint16
	DeviceType           DeviceType
	Status               byte
	ProfileID            uint16
	ManufacturerID       uint16
}

type DeviceType byte

const (
	Coordinator DeviceType = 0
	Router      DeviceType = 1
	EndDevice   DeviceType = 2
)

func (dt DeviceType) String() string {
	switch dt {
	case Coordinator:
		return "Coordinator"
	case Router:
		return "Router"
	case EndDevice:
		return "EndDevice"
	}
	return fmt.Sprintf("DeviceType(%d)", dt)
}

type NodeDiscoveryOption int

const (
	NDOIncludeDeviceType  NodeDiscoveryOption = 1
	NDOIncludeLocalDevice NodeDiscoveryOption = 2
)

func (o NodeDiscoveryOption) Has(opt NodeDiscoveryOption) bool {
	return (o & opt) != 0
}

func (o NodeDiscoveryOption) String() string {
	if o == 0 {
		return "None"
	}
	var opts []string
	if o.Has(NDOIncludeDeviceType) {
		opts = append(opts, "IncludeDeviceType")
		o &^= NDOIncludeDeviceType
	}
	if o.Has(NDOIncludeLocalDevice) {
		opts = append(opts, "IncludeLocalDevice")
		o &^= NDOIncludeLocalDevice
	}
	if o != 0 {
		opts = append(opts, fmt.Sprintf("NodeDiscoveryOption(%d)", o))
	}
	return strings.Join(opts, "|")
}

type SecurityOption int

const (
	SOSendUnsecureKeyOTA SecurityOption = 1
	SOUseTrustCenter     SecurityOption = 2
)

func (o SecurityOption) Has(opt SecurityOption) bool {
	return (o & opt) != 0
}

func (o SecurityOption) String() string {
	if o == 0 {
		return "None"
	}
	var opts []string
	if o.Has(SOSendUnsecureKeyOTA) {
		opts = append(opts, "SendUnsecureKeyOTA")
		o &^= SOSendUnsecureKeyOTA
	}
	if o.Has(SOUseTrustCenter) {
		opts = append(opts, "UseTrustCenter")
		o &^= SOUseTrustCenter
	}
	if o != 0 {
		opts = append(opts, fmt.Sprintf("SecurityOption(%d)", o))
	}
	return strings.Join(opts, "|")
}

type ModemStatus byte

const (
	MSHardwareReset              ModemStatus = 0
	MSWatchdogTimerReset         ModemStatus = 1
	MSJoinedNetwork              ModemStatus = 2 // routers and end devices
	MSDisassociated              ModemStatus = 3
	MSCoordinatorStarted         ModemStatus = 6
	MSNetworkKeyUpdated          ModemStatus = 7
	MSVoltageSupplyLimitExceeded ModemStatus = 0x0d // PRO S2B only
	MSConfigChangeDuringJoin     ModemStatus = 0x11
)

func (ms ModemStatus) String() string {
	switch ms {
	case MSHardwareReset:
		return "HardwareReset"
	case MSWatchdogTimerReset:
		return "WatchdogTimerReset"
	case MSJoinedNetwork:
		return "JoinedNetwork"
	case MSDisassociated:
		return "Disassociated"
	case MSCoordinatorStarted:
		return "CoordinatorStarted"
	case MSNetworkKeyUpdated:
		return "NetworkKeyUpdated"
	case MSVoltageSupplyLimitExceeded:
		return "VoltageSupplyLimitExceeded"
	case MSConfigChangeDuringJoin:
		return "ConfigChangeDuringJoin"
	}
	if ms >= 0x80 {
		return "StackError"
	}
	return fmt.Sprintf("ModemStatus(%d)", ms)
}

type DeliveryStatus byte

const (
	DSSuccess                               DeliveryStatus = 0x00
	DSMACACKFailure                         DeliveryStatus = 0x01
	DSCCAFailure                            DeliveryStatus = 0x02
	DSInvalidDestinationEndpoint            DeliveryStatus = 0x15
	DSNetworkACKFailure                     DeliveryStatus = 0x21
	DSNotJoinedToNetwork                    DeliveryStatus = 0x22
	DSSelfAddressed                         DeliveryStatus = 0x23
	DSAddressNotFound                       DeliveryStatus = 0x24
	DSRouteNotFound                         DeliveryStatus = 0x25
	DSBroadcastFail                         DeliveryStatus = 0x26 // Broadcast source failed to hear a neighbor relay the message
	DSInvalidBindingTableIndex              DeliveryStatus = 0x2B
	DSResourceError                         DeliveryStatus = 0x2C // Resource error lack of free buffers, timers, and so forth.
	DSAttemptedBroadcastWithAPSTransmittion DeliveryStatus = 0x2D
	DSAttemptedUnicastWithAPSTransmission   DeliveryStatus = 0x2E // Attempted unicast with APS transmission, but EE=0
	DSResourceError2                        DeliveryStatus = 0x32 // Resource error lack of free buffers, timers, and so
	DSDataPayloadTooLarge                   DeliveryStatus = 0x74
)

func (ds DeliveryStatus) String() string {
	switch ds {
	case DSSuccess:
		return "Success"
	case DSMACACKFailure:
		return "MACACKFailure"
	case DSCCAFailure:
		return "CCAFailure"
	case DSInvalidDestinationEndpoint:
		return "InvalidDestinationEndpoint"
	case DSNetworkACKFailure:
		return "NetworkACKFailure"
	case DSNotJoinedToNetwork:
		return "NotJoinedToNetwork"
	case DSSelfAddressed:
		return "SelfAddressed"
	case DSAddressNotFound:
		return "AddressNotFound"
	case DSRouteNotFound:
		return "RouteNotFound"
	case DSBroadcastFail:
		return "BroadcastFail"
	case DSInvalidBindingTableIndex:
		return "InvalidBindingTableIndex"
	case DSResourceError:
		return "ResourceError"
	case DSAttemptedBroadcastWithAPSTransmittion:
		return "AttemptedBroadcastWithAPSTransmittion"
	case DSAttemptedUnicastWithAPSTransmission:
		return "AttemptedUnicastWithAPSTransmission"
	case DSResourceError2:
		return "ResourceError2"
	case DSDataPayloadTooLarge:
		return "DataPayloadTooLarge"
	}
	return fmt.Sprintf("DeliveryStatus(%d)", ds)
}

type DiscoveryStatus byte

const (
	DSNoDiscoveryOverhead      DiscoveryStatus = 0x00
	DSAddressDiscovery         DiscoveryStatus = 0x01
	DSRouteDiscovery           DiscoveryStatus = 0x02
	DSAddressAndRoute          DiscoveryStatus = 0x03
	DSExtendedTimeoutDiscovery DiscoveryStatus = 0x40
)

func (ds DiscoveryStatus) String() string {
	switch ds {
	case DSNoDiscoveryOverhead:
		return "NoDiscoveryOverhead"
	case DSAddressDiscovery:
		return "AddressDiscovery"
	case DSRouteDiscovery:
		return "RouteDiscovery"
	case DSAddressAndRoute:
		return "AddressAndRoute"
	case DSExtendedTimeoutDiscovery:
		return "ExtendedTimeoutDiscovery"
	}
	return fmt.Sprintf("DiscoveryStatus(%d)", ds)
}

type TransmitStatus struct {
	DestinationAddress uint16
	RetryCount         int
	DeliveryStatus     DeliveryStatus
	DiscoveryStatus    DiscoveryStatus
}

type ReceiveOption byte

const (
	ROAcknowledged  ReceiveOption = 0x01
	ROBroadcast     ReceiveOption = 0x02
	ROEncrypted     ReceiveOption = 0x20
	ROFromEndDevice ReceiveOption = 0x40
)

func (o ReceiveOption) Has(opt ReceiveOption) bool {
	return (o & opt) != 0
}

func (o ReceiveOption) String() string {
	if o == 0 {
		return "None"
	}
	var opts []string
	if o.Has(ROAcknowledged) {
		opts = append(opts, "Acknowledged")
		o &^= ROAcknowledged
	}
	if o.Has(ROBroadcast) {
		opts = append(opts, "Broadcast")
		o &^= ROBroadcast
	}
	if o.Has(ROEncrypted) {
		opts = append(opts, "Encrypted")
		o &^= ROEncrypted
	}
	if o.Has(ROFromEndDevice) {
		opts = append(opts, "FromEndDevice")
		o &^= ROFromEndDevice
	}
	if o != 0 {
		opts = append(opts, fmt.Sprintf("ReceiveOption(%d)", o))
	}
	return strings.Join(opts, "|")
}

type ReceivePacket struct {
	SourceAddress   uint64
	SourceAddress16 uint16
	ReceiveOptions  ReceiveOption
	Data            []byte
}

type CommandStatus byte

const (
	CSOK               CommandStatus = 0
	CSError            CommandStatus = 1
	CSInvalidCommand   CommandStatus = 2
	CSInvalidParameter CommandStatus = 3
	CSTxFailure        CommandStatus = 4
)

func (cs CommandStatus) String() string {
	switch cs {
	case CSOK:
		return "OK"
	case CSError:
		return "Error"
	case CSInvalidCommand:
		return "InvalidCommand"
	case CSInvalidParameter:
		return "InvalidParameter"
	case CSTxFailure:
		return "TxFailure"
	}
	return fmt.Sprintf("CommandStatus(%d)", cs)
}

type ATCommandResponse struct {
	ATCommand     ATCommand
	CommandStatus CommandStatus
	Data          []byte
}

type UnknownFrame []byte

type XBee struct {
	port    io.ReadWriter
	wbuf    []byte
	frameID byte
	eventCh chan Event
	mu      sync.Mutex
	idMap   map[byte]chan Event
}

type Event interface{}

const (
	AddressCoordinator uint64 = 0x0000000000000000
	AddressBroadcast   uint64 = 0x000000000000FFFF
	Address16Unknown   uint16 = 0xFFFE
	Address16Broadcast uint16 = 0xFFFE
)

type TransmitOption byte

const (
	TODisableRetriesAndRouteRepair TransmitOption = 0x01
	TOEnableAPSEncryption          TransmitOption = 0x20
	TOExtendedTxTimeout            TransmitOption = 0x40
)

func (o TransmitOption) Has(opt TransmitOption) bool {
	return (o & opt) != 0
}

func (o TransmitOption) String() string {
	if o == 0 {
		return "None"
	}
	var opts []string
	if o.Has(TODisableRetriesAndRouteRepair) {
		opts = append(opts, "ExtendedTxTimeout")
		o &^= TODisableRetriesAndRouteRepair
	}
	if o.Has(TOEnableAPSEncryption) {
		opts = append(opts, "TOEnableAPSEncryption")
		o &^= TOEnableAPSEncryption
	}
	if o.Has(TOExtendedTxTimeout) {
		opts = append(opts, "TOExtendedTxTimeout")
		o &^= TOExtendedTxTimeout
	}
	if o != 0 {
		opts = append(opts, fmt.Sprintf("TransmitOption(%d)", o))
	}
	return strings.Join(opts, "|")
}

func Open(device io.ReadWriter) (*XBee, error) {
	xb := &XBee{
		port:    device,
		wbuf:    make([]byte, 65535+4),
		eventCh: make(chan Event, 8),
		idMap:   make(map[byte]chan Event),
	}
	xb.wbuf[0] = frameDelimiter
	go func() {
		err := xb.readLoop()
		if err != nil {
			log.Printf("Error during read loop: %s", err)
		}
	}()
	return xb, nil
}

func (xb *XBee) Close() {
	close(xb.eventCh)
}

func (xb *XBee) EventChan() chan Event {
	return xb.eventCh
}

func (xb *XBee) atCommand(cmd ATCommand, val []byte) ([]byte, error) {
	if len(val) > 65536-8 {
		return nil, fmt.Errorf("xbee: value too long for at command write (%d bytes)", len(val))
	}
	frameID := xb.nextFrameID()
	xb.wbuf[3] = frameATCommand
	xb.wbuf[4] = frameID
	xb.wbuf[5] = cmd[0]
	xb.wbuf[6] = cmd[1]
	if len(val) != 0 {
		copy(xb.wbuf[7:], val)
	}
	ch := xb.registerListener(frameID)
	defer xb.unregisterListener(frameID)
	if err := xb.writeFrame(4 + len(val)); err != nil {
		return nil, err
	}
	ev := <-ch
	res, ok := ev.(*ATCommandResponse)
	if !ok {
		return nil, fmt.Errorf("xbee: wrong frame, expected AT response got %T", ev)
	}
	if err := validateATResponse(cmd, res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (xb *XBee) SerialNumber() (uint64, error) {
	res, err := xb.atCommand(atSerialNumberHigh, nil)
	if err != nil {
		return 0, err
	}
	if len(res) != 4 {
		return 0, fmt.Errorf("xbee.SerialNumber: expected 4 bytes got %d", len(res))
	}
	serial := uint64(binary.BigEndian.Uint32(res)) << 32
	res, err = xb.atCommand(atSerialNumberLow, nil)
	if err != nil {
		return 0, err
	}
	if len(res) != 4 {
		return 0, fmt.Errorf("xbee.SerialNumber: expected 4 bytes got %d", len(res))
	}
	serial |= uint64(binary.BigEndian.Uint32(res))
	return serial, nil
}

func (xb *XBee) NodeIdentifier() (string, error) {
	ni, err := xb.atCommand(atNodeIdentifier, nil)
	return string(ni), err
}

func (xb *XBee) SetNodeIdentifier(ni string) error {
	if len(ni) > 20 {
		ni = ni[:20]
	}
	_, err := xb.atCommand(atNodeIdentifier, []byte(ni))
	return err
}

func (xb *XBee) DeviceTypeIdentifier() (uint32, error) {
	b, err := xb.atCommand(atDeviceTypeIdentifier, nil)
	if err != nil {
		return 0, err
	}
	var did uint32
	for _, b := range b {
		did <<= 8
		did |= uint32(b)
	}
	return did, err
}

// FirmwareVersion returns the versin of the XBee firmware.
// The firmware version returns 4 hexadecimal values (2 bytes) “ABCD”.
// Digits ABC are the main release number and D is the revision number
// from the main release. “B” is a variant designator (0=coordinator AT,
// 1=coordinated API, 2=router AT, 3=router API, 8=end device AT,
// 9=end device API).
func (xb *XBee) FirmwareVersion() (uint16, error) {
	b, err := xb.atCommand(atFirmwareVersion, nil)
	if err != nil {
		return 0, err
	}
	if len(b) != 2 {
		return 0, fmt.Errorf("xbee.FirmwareVersion: expected 2 byte response got %d", len(b))
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

func (xb *XBee) HardwareVersion() (uint16, error) {
	b, err := xb.atCommand(atHardwareVersion, nil)
	if err != nil {
		return 0, err
	}
	if len(b) != 2 {
		return 0, fmt.Errorf("xbee.HardwareVersion: expected 2 byte response got %d", len(b))
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

func (xb *XBee) AssociationIndication() (int, error) {
	b, err := xb.atCommand(atAssociationIndication, nil)
	if err != nil {
		return 0, err
	}
	return int(decodeUint(b)), nil
}

func (xb *XBee) InterfaceDataRate() (int, error) {
	b, err := xb.atCommand(atInterfaceDataRate, nil)
	if err != nil {
		return 0, err
	}
	rate := decodeUint(b)
	if rate <= 7 {
		return standardDataRates[rate], nil
	}
	return int(rate), nil
}

func (xb *XBee) SetInterfaceDataRate(baud int) error {
	// Match against standard rates first
	si := -1
	for i, b := range standardDataRates {
		if b == baud {
			si = i
			break
		}
	}
	if si < 0 {
		si = baud
	}
	buf := []byte{
		byte(si >> 24),
		byte(si >> 16),
		byte(si >> 8),
		byte(si & 0xff),
	}
	_, err := xb.atCommand(atInterfaceDataRate, buf)
	return err
}

func (xb *XBee) ExtendedPANID() (uint64, error) {
	b, err := xb.atCommand(atExtendedPANID, nil)
	if err != nil {
		return 0, err
	}
	return uint64(decodeUint(b)), nil
}

func (xb *XBee) SetExtendedPANID(id uint64) error {
	b := []byte{
		byte(id >> 56),
		byte(id >> 48),
		byte(id >> 40),
		byte(id >> 32),
		byte(id >> 24),
		byte(id >> 16),
		byte(id >> 8),
		byte(id & 0xff),
	}
	_, err := xb.atCommand(atExtendedPANID, b)
	return err
}

func (xb *XBee) OperatingExtendedPANID() (uint64, error) {
	b, err := xb.atCommand(atOperatingExtendedPANID, nil)
	if err != nil {
		return 0, err
	}
	return uint64(decodeUint(b)), nil
}

func (xb *XBee) MaximumRFPayloadBytes() (int, error) {
	b, err := xb.atCommand(atMaximumRFPayloadBytes, nil)
	if err != nil {
		return 0, err
	}
	return int(decodeUint(b)), nil
}

func (xb *XBee) EncryptionEnabled() (bool, error) {
	b, err := xb.atCommand(atEncryptionEnabled, nil)
	if err != nil {
		return false, err
	}
	return len(b) != 0 && b[0] != 0, nil
}

func (xb *XBee) SetEncryptionEnabled(enabled bool) error {
	b := []byte{0}
	if enabled {
		b[0] = 1
	}
	_, err := xb.atCommand(atEncryptionEnabled, b)
	return err
}

func (xb *XBee) EncryptionOptions() (SecurityOption, error) {
	b, err := xb.atCommand(atEncryptionOptions, nil)
	if err != nil {
		return 0, err
	}
	return SecurityOption(decodeUint(b)), nil
}

func (xb *XBee) SetEncryptionOptions(opt SecurityOption) error {
	_, err := xb.atCommand(atEncryptionOptions, []byte{byte(opt)})
	return err
}

func (xb *XBee) SetNetworkEncryptionKey(key []byte) error {
	if len(key) == 0 {
		// The xbee will set a random key when provided with zero key
		key = make([]byte, 16)
	} else if len(key) != 16 {
		return fmt.Errorf("xbee.SetLinkKey: key must be 128-bits (16 bytes) not %d-bits", len(key)*8)
	}
	_, err := xb.atCommand(atNetworkEncryptionKey, key)
	return err
}

func (xb *XBee) SetLinkKey(key []byte) error {
	if len(key) != 16 {
		return fmt.Errorf("xbee.SetLinkKey: key must be 128-bits (16 bytes) not %d-bits", len(key)*8)
	}
	_, err := xb.atCommand(atLinkKey, key)
	return err
}

func (xb *XBee) NodeDiscoveryTimeout() (time.Duration, error) {
	b, err := xb.atCommand(atNodeDiscoveryTimeout, nil)
	if err != nil {
		return 0, err
	}
	return time.Duration(time.Duration(decodeUint(b)) * 100 * time.Millisecond), nil
}

func (xb *XBee) NodeDiscoveryOptions() (NodeDiscoveryOption, error) {
	b, err := xb.atCommand(atNodeDiscoveryOptions, nil)
	if err != nil {
		return 0, err
	}
	return NodeDiscoveryOption(int(decodeUint(b))), nil
}

func (xb *XBee) SetNodeDiscoveryOptions(o NodeDiscoveryOption) error {
	_, err := xb.atCommand(atNodeDiscoveryOptions, []byte{byte(o)})
	return err
}

func (xb *XBee) APIEnabled() (escaped bool, err error) {
	b, err := xb.atCommand(atAPIEnable, nil)
	return b[0] == 2, err
}

func (xb *XBee) Write() error {
	_, err := xb.atCommand(atWrite, nil)
	return err
}

func (xb *XBee) NodeDiscover(wait time.Duration) ([]*Node, error) {
	frameID := xb.nextFrameID()
	xb.wbuf[3] = frameATCommand
	xb.wbuf[4] = frameID
	xb.wbuf[5] = atNodeDiscover[0]
	xb.wbuf[6] = atNodeDiscover[1]
	ch := xb.registerListener(frameID)
	defer xb.unregisterListener(frameID)
	if err := xb.writeFrame(4); err != nil {
		return nil, err
	}
	waitCh := time.After(wait)
	var nodes []*Node
	for {
		var ev Event
		select {
		case <-waitCh:
			return nodes, nil
		case ev = <-ch:
		}
		res, ok := ev.(*ATCommandResponse)
		if !ok {
			return nil, fmt.Errorf("xbee: wrong frame, expected AT response got %T", ev)
		}
		if err := validateATResponse(atNodeDiscover, res); err != nil {
			return nodes, err
		}
		if len(res.Data) < 18 {
			return nodes, fmt.Errorf("xbee.NodeDiscover: device frame should be at least 18 bytes, got %d", len(res.Data))
		}

		// 2 bytes for what?
		// uint64 serial number
		// zero terminated node identifier
		// uint16 parent network address
		// byte   device type (0=coord, 1=router, 2=end device)
		// byte   status (reserved)
		// uint16 profile ID
		// uint16 manufacturer ID

		n := &Node{SerialNumber: decodeUint(res.Data[2:10])}
		res.Data = res.Data[10:]
		ix := bytes.IndexByte(res.Data, 0)
		if ix < 0 {
			return nodes, errors.New("xbee.NodeDiscover: null terminator not found for node identifier")
		}
		n.NodeID = string(res.Data[:ix])
		res.Data = res.Data[ix+1:]
		n.ParentNetworkAddress = (uint16(res.Data[0]) << 8) | uint16(res.Data[1])
		n.DeviceType = DeviceType(res.Data[2])
		n.Status = res.Data[3]
		n.ProfileID = (uint16(res.Data[4]) << 8) | uint16(res.Data[5])
		n.ManufacturerID = (uint16(res.Data[6]) << 8) | uint16(res.Data[7])
		nodes = append(nodes, n)
	}
}

func (xb *XBee) ActiveScan(wait time.Duration) ([]*ActiveScanDevice, error) {
	frameID := xb.nextFrameID()
	xb.wbuf[3] = frameATCommand
	xb.wbuf[4] = frameID
	xb.wbuf[5] = atActiveScan[0]
	xb.wbuf[6] = atActiveScan[1]
	ch := xb.registerListener(frameID)
	defer xb.unregisterListener(frameID)
	if err := xb.writeFrame(4); err != nil {
		return nil, err
	}
	waitCh := time.After(wait)
	var devices []*ActiveScanDevice
	for {
		var ev Event
		select {
		case <-waitCh:
			return devices, nil
		case ev = <-ch:
		}
		res, ok := ev.(*ATCommandResponse)
		if !ok {
			return nil, fmt.Errorf("xbee: wrong frame, expected AT response got %T", ev)
		}
		if err := validateATResponse(atActiveScan, res); err != nil {
			return devices, err
		}
		if len(res.Data) < 16 {
			return devices, fmt.Errorf("xbee.ActiveScan: device frame should be at least 16 bytes, got %d", len(res.Data))
		}
		devices = append(devices, &ActiveScanDevice{
			Type:         res.Data[0],
			Channel:      res.Data[1],
			PAN:          uint16(decodeUint(res.Data[2:4])),
			ExtendedPAN:  uint64(decodeUint(res.Data[4:12])),
			AllowJoin:    res.Data[12] != 0,
			StackProfile: res.Data[13],
			LQI:          res.Data[14],
			RSSI:         int8(res.Data[15]),
		})
	}
}

func validateATResponse(cmd ATCommand, res *ATCommandResponse) error {
	if res.ATCommand != cmd {
		return fmt.Errorf("xbee: expected AT command response cmd %02x%02x got %02x%02x", cmd[0], cmd[1], res.ATCommand[0], res.ATCommand[1])
	}
	switch res.CommandStatus {
	case CSOK: // OK
	case CSError:
		return ErrResponse
	case CSInvalidCommand:
		return ErrInvalidCommand(string(cmd[:]))
	case CSInvalidParameter:
		return ErrInvalidParameter
	default:
		return fmt.Errorf("xbee: unknown error %d", res.CommandStatus)
	}
	return nil
}

func (xb *XBee) Transmit(dest uint64, net uint16, broadcastRadius byte, options TransmitOption, data []byte) error {
	if len(data) > 65536-20 {
		return fmt.Errorf("xbee: data too long for transmit (%d bytes)", len(data))
	}
	frameID := xb.nextFrameID()
	xb.wbuf[3] = frameZigBeeTransmitRequest
	xb.wbuf[4] = frameID
	xb.wbuf[5] = byte(dest >> 56)
	xb.wbuf[6] = byte(dest >> 48)
	xb.wbuf[7] = byte(dest >> 40)
	xb.wbuf[8] = byte(dest >> 32)
	xb.wbuf[9] = byte(dest >> 24)
	xb.wbuf[10] = byte(dest >> 16)
	xb.wbuf[11] = byte(dest >> 8)
	xb.wbuf[12] = byte(dest & 0xff)
	xb.wbuf[13] = byte(net >> 8)
	xb.wbuf[14] = byte(net & 0xff)
	xb.wbuf[15] = broadcastRadius
	xb.wbuf[16] = byte(options)
	copy(xb.wbuf[17:], data)
	// ch := xb.registerListener(frameID)
	if err := xb.writeFrame(14 + len(data)); err != nil {
		return err
	}
	// frame := <-ch
	// xb.unregisterListener(frameID)
	return nil
	// if len(frame) < 5 {
	// 	return nil, errors.New("xbee: short frame for AT command response")
	// }
	// if frame[0] != frameATCommandResponse {
	// 	return nil, fmt.Errorf("xbee: expected AT command response frame type 0x%02x got 0x%02x", frameATCommandResponse, frame[0])
	// }
	// if frame[2] != cmd[0] || frame[3] != cmd[1] {
	// 	return nil, fmt.Errorf("xbee: expected AT command response cmd %02x%02x got %02x%02x", cmd[0], cmd[1], frame[2], frame[3])
	// }
	// switch frame[4] {
	// case 0: // Ok
	// case 1:
	// 	return nil, ErrResponse
	// case 2:
	// 	return nil, ErrInvalidCommand
	// case 3:
	// 	return nil, ErrInvalidParameter
	// default:
	// 	return nil, fmt.Errorf("xbee: unknown error %d", frame[4])
	// }
	// return frame[5:], nil
}

func (xb *XBee) registerListener(frameID byte) chan Event {
	xb.mu.Lock()
	defer xb.mu.Unlock()
	ch := make(chan Event, 1)
	xb.idMap[frameID] = ch
	return ch
}

func (xb *XBee) unregisterListener(frameID byte) {
	xb.mu.Lock()
	defer xb.mu.Unlock()
	delete(xb.idMap, frameID)
}

func (xb *XBee) nextFrameID() byte {
	xb.frameID++
	if xb.frameID == 0 {
		xb.frameID = 1
	}
	return xb.frameID
}

func (xb *XBee) writeFrame(n int) error {
	if n < 0 || n > 65535 {
		return fmt.Errorf("xbee: cannot write frame of size %d", n)
	}
	// length of frame
	xb.wbuf[1] = byte(n >> 8)
	xb.wbuf[2] = byte(n & 0xff)
	// calculate checksum
	var checksum byte
	for _, v := range xb.wbuf[3 : 3+n] {
		checksum += v
	}
	xb.wbuf[3+n] = 0xff - checksum
	_, err := xb.port.Write(xb.wbuf[:4+n])
	return err
}

func (xb *XBee) readLoop() error {
	rd := bufio.NewReader(xb.port)
	buf := make([]byte, 256)
	for {
		if buf == nil {
			buf = make([]byte, 256)
		} else {
			buf = buf[:cap(buf)]
		}

		// Read frame delimiter
		if b, err := rd.ReadByte(); err != nil {
			return err
		} else if b != frameDelimiter {
			log.Printf("xbee.readLoop: received 0x%02x while looking for frame delimiter\n", b)
			continue
		}
		// Read frame length
		buf = buf[:2]
		if _, err := io.ReadFull(rd, buf); err != nil {
			return err
		}
		frameLen := (int(buf[0]) << 8) | int(buf[1])
		// +1 for checksum
		if cap(buf) < frameLen+1 {
			buf = make([]byte, frameLen+1)
		} else {
			buf = buf[:frameLen+1]
		}
		if _, err := io.ReadFull(rd, buf); err != nil {
			return err
		}
		checksum := byte(0)
		for _, b := range buf {
			checksum += b
		}
		if checksum != 0xff {
			log.Printf("xbee: bad frame checksum %02x\n", checksum)
		} else if frameLen < 2 {
			// Normal frames have at least 2 bytes for type and ID
			log.Println("xbee: tiny frame received")
		} else {
			buf = buf[:len(buf)-1]

			var frameID byte
			var ev Event
			switch buf[0] {
			case frameModemStatus:
				ev = ModemStatus(buf[1])
			case frameATCommandResponse:
				frameID = buf[1]
				ev = &ATCommandResponse{
					ATCommand:     ATCommand([2]byte{buf[2], buf[3]}),
					CommandStatus: CommandStatus(buf[4]),
					Data:          buf[5:],
				}
				if len(buf) > 5 {
					buf = nil
				}
			case frameZigBeeTransmitStatus:
				frameID = buf[1]
				ev = &TransmitStatus{
					DestinationAddress: (uint16(buf[2]) << 8) | uint16(buf[3]),
					RetryCount:         int(buf[4]),
					DeliveryStatus:     DeliveryStatus(buf[5]),
					DiscoveryStatus:    DiscoveryStatus(buf[6]),
				}
			case frameZigBeeReceivePacket:
				ev = &ReceivePacket{
					SourceAddress: (uint64(buf[1]) << 56) | (uint64(buf[2]) << 48) |
						(uint64(buf[3]) << 40) | (uint64(buf[4]) << 32) |
						(uint64(buf[5]) << 24) | (uint64(buf[6]) << 16) |
						(uint64(buf[7]) << 8) | uint64(buf[8]),
					SourceAddress16: (uint16(buf[9]) << 8) | uint16(buf[10]),
					ReceiveOptions:  ReceiveOption(buf[11]),
					Data:            buf[12:],
				}
				buf = nil
			default:
				ev = UnknownFrame(buf)
				buf = nil
			}

			var ch chan Event
			if frameID != 0 {
				xb.mu.Lock()
				ch = xb.idMap[frameID]
				xb.mu.Unlock()
			}
			if ch != nil {
				select {
				case ch <- ev:
				default:
					// Should never happen but better to be safe
					log.Println("xbee: internal event channel full")
				}
			} else {
				select {
				case xb.eventCh <- ev:
				default:
					log.Println("xbee: event channel full")
				}
			}
		}
	}
}

func decodeUint(b []byte) uint64 {
	var v uint64
	for _, b := range b {
		v = (v << 8) | uint64(b)
	}
	return v
}
