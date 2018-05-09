package xbee

type ATCommand [2]byte

// Addressing commands
var (
	// Destination Address High.Set/Get the upper 32
	// bits of the 64-bit destination address. When
	// combined with DL, it defines the 64-bit
	// destination address for data transmission.
	// Special definitions for DH and DL include
	// 0x000000000000FFFF (broadcast) and
	// 0x0000000000000000 (coordinator).
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFFFFFF
	atDestinationAddressHigh = ATCommand([2]byte{'D', 'H'})
	// Destination Address Low. Set/Get the lower 32
	// bits of the 64-bit destination address. When
	// combined with DH, it defines the 64-bit
	// destination address for data transmissions.
	// Special definitions for DH and DL include
	// 0x000000000000FFFF (broadcast) and
	// 0x0000000000000000 (coordinator).
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFFFFFF
	atDestinationAddressLow = ATCommand([2]byte{'D', 'L'})
	// 16-bit Network Address. Read the 16-bit
	// network address of the module. A value of
	// 0xFFFE means the module has not joined a
	// ZigBee network.
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFE [read-only]
	at16BitNetworkAddress = ATCommand([2]byte{'M', 'Y'})
	// 16-bit Parent Network Address. Read the 16-bit
	// network address of the module's parent. A value
	// of 0xFFFE means the module does not have a parent.
	// Node Type: E
	// Parameter Range: 0 - 0xFFFE [read-only]
	at16BitParentNetworkAddress = ATCommand([2]byte{'M', 'P'})
	// Number of Remaining Children. Read the
	// number of end device children that can join the
	// device. If NC returns 0, then the device cannot
	// allow any more end device children to join.
	// Node Type: CR
	// Parameter Range: 0 - MAX_CHILDREN (maximum varies) [read-only]
	atChildrenRemaining = ATCommand([2]byte{'N', 'C'})
	// Serial Number High. Read the high 32 bits of the
	// module's unique 64-bit address.
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFFFFFF
	atSerialNumberHigh = ATCommand([2]byte{'S', 'H'})
	// Serial Number Low. Read the low 32 bits of the
	// module's unique 64-bit address.
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFFFFFF
	atSerialNumberLow = ATCommand([2]byte{'S', 'L'})
	// Node Identifier. Set/read a string identifier.
	// The register only accepts printable ASCII data.
	// Node Type: CRE
	// Parameter Range: 20-byte printable ASCII string
	// Default: ASCII space character (0x20)
	atNodeIdentifier = ATCommand([2]byte{'N', 'I'})
	// Maximum RF Payload Bytes. This value returns the maximum number
	// of RF payload bytes that can be sent in a unicast transmission.
	// If APS encryption is used (API transmit option bit enabled), the
	// maximum payload size is reduced by 9 bytes. If source routing is
	// used (AR < 0xFF), the maximum payload size is reduced further.
	// Note NP returns a hexadecimal value. (for example if NP returns
	// 0x54, this is equivalent to 84 bytes).
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFF
	// [read-only]
	atMaximumRFPayloadBytes = ATCommand([2]byte{'N', 'P'})
	// Device Type Identifier. Stores a device type value. This value
	// can be used to differentiate different XBee-based devices. Digi
	// reserves the range 0 - 0xFFFFFF.
	// For example, Digi currently uses the following DD values to
	// identify various ZigBee products:
	// 0x30001 - ConnectPort X8 Gateway
	// 0x30002 - ConnectPort X4 Gateway
	// 0x30003 - ConnectPort X2 Gateway
	// 0x30005 - RS-232 Adapter
	// 0x30006 - RS-485 Adapter
	// 0x30007 - XBee Sensor Adapter
	// 0x30008 - Wall Router
	// 0x3000A - Digital I/O Adapter
	// 0x3000B - Analog I/O Adapter
	// 0x3000C - XStick
	// 0x3000F - Smart Plug
	// 0x30011 - XBee Large Display
	// 0x30012 - XBee Small Display
	// Parameter range: 0 - 0xFFFFFFFF
	// Node Type: CRE
	// Default: 0x30000
	atDeviceTypeIdentifier = ATCommand([2]byte{'D', 'D'})
	// Conflict Report. The number of PAN id conflict reports
	// that must be received by the network manager within one
	// minute to trigger a PAN ID change. A corrupt beacon can
	// cause a report of a false PAN id conflict. A higher value
	// reduces the chance of a spurious PAN ID change.
	// Parameter range: 1-0x3f
	// Node Type: CRE
	// Default: 3
	atConflictReport = ATCommand([2]byte{'C', 'R'})
)

// Networking Commands
var (
	// CH - Operating Channel
	// DA - Force Disassociation

	// Extended PAN ID. Set/read the 64-bit extended PAN ID. If set to 0,
	// the coordinator will select a random extended PAN ID, and the
	// router / end device will join any extended PAN ID. Changes to ID
	// should be written to non-volatile memory using the WR command to
	// preserve the ID setting if a power cycle occurs.
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFFFFFFFFFFFFFF
	// Default: 0
	atExtendedPANID = ATCommand([2]byte{'I', 'D'})
	// Operating Extended PAN ID. Read the 64-bit extended PAN ID. The
	// OP value reflects the operating extended PAN ID that the module
	// is running on. If ID > 0, OP will equal ID.
	// Node Type: CRE
	// Parameter Range: 0x01 - 0xFFFFFFFFFFFFFFFF
	// Default: [read-only]
	atOperatingExtendedPANID = ATCommand([2]byte{'O', 'P'})

	// NH - Maximum Unicast Hops
	// BH - Broadcast Hops
	// OI - Operating 16-bit PAN ID

	// Node Discovery Timeout. Set/Read the node discovery timeout. When the
	// network discovery (ND) command is issued, the NT value is included in
	// the transmission to provide all remote devices with a response timeout.
	// Remote devices wait a random time, less than NT, before sending their
	// response.
	// Node Type: CRE
	// Parameter Range: 0x20 - 0xFF [x 100 msec]
	// Default: 0x3c (60d)
	atNodeDiscoveryTimeout = ATCommand([2]byte{'N', 'T'})
	// Network Discovery options. Set/Read the options value for the network
	// discovery command. The options bitfield value can change the behavior
	// of the ND (network discovery) command and/or change what optional
	// values are returned in any received ND responses or API node
	// identification frames. Options include: 0x01 = Append DD value (to ND
	// responses or API node identification frames), 0x02 = Local device
	// sends ND response frame when ND is issued)
	// Node Type: CRE
	// Parameter Range: 0 - 0x03 [bitfield]
	// Default: 0
	atNodeDiscoveryOptions = ATCommand([2]byte{'N', 'O'})

// SC - Scan Channels
// SD - Scan Duration
// ZS - ZigBee Stack Profile
// NJ - Node Join Time
// JV - Channel Verification
// NW - Network Watchdog Timeout
// JN - Join Notification
// AR - Aggregate Routing Notification
// DJ - Disable Joining
// II - Initial ID
)

// Security Commands
var (
	// Encryption Enabled
	// Node Type: CRE
	// Parameter Range: 0 - disabled, 1 - enabled
	// Default: 0
	atEncryptionEnabled = ATCommand([2]byte{'E', 'E'})
	// Encryption Options. Configure options for encryption. Unused
	// option bits should be set to 0. Options include:
	//
	// 0x01 - Send the security key unsecured over-the-air during joins
	// 0x02 - Use trust center (coordinator only)
	atEncryptionOptions = ATCommand([2]byte{'E', 'O'})
	// Network encryption key. Set the 128-bit AES network encryption key.
	// This command is write-only; NK cannot be read. If set to 0 (default),
	// the module will select a random network key.
	// Node Type: C
	atNetworkEncryptionKey = ATCommand([2]byte{'N', 'K'})
	// Link key. Set the 128-bit AES link key. This command is writeonly;
	// KY cannot be read. Setting KY to 0 will cause the coordinator to
	// transmit the network key in the clear to joining devices, and will
	// cause joining devices to acquire the network key in the clear when
	// joining.
	atLinkKey = ATCommand([2]byte{'K', 'Y'})
)

// RF Interface Commands
var (
// PL - Power Level
// PM - Power Mode
// DB - Received Signal Strength
// PP - Peak Power
)

// Serial Interfacing (I/O) Commands
var (
	// API Enable. Enable API Mode. The AP command is only supported when
	// using API firmware: 21xx (API coordinator), 23xx (API router), 29xx
	// (API end device).
	atAPIEnable = ATCommand([2]byte{'A', 'P'})
	// AO - API Options

	// Interface Data Rate. Set/Read the serial interface data rate
	// for communication between the module serial port and host.
	// Any value above 0x07 will be interpreted as an actual baud rate.
	// When a value above 0x07 is sent, the closest interface data rate
	// represented by the number is stored in the BD register.
	// Node Type: CRE
	// Paramter Range:
	// 0-7
	// (standard baud rates)
	//   0 = 1200 b/s 1 = 2400
	//   2 = 4800
	//   3 = 9600
	//   4 = 19200 5 = 38400 6 = 57600 7 = 115200
	// 0x80 - 0xE1000 (non-standard rates up to 921kb/s)
	// Default: 3
	atInterfaceDataRate = ATCommand([2]byte{'B', 'D'})

// NB - Serial Parity
// SB - Stop Bits
// D7 - DIO7 Configuration
// D6 - DIO6 Configuration
)

// I/O Commands

// Diagnostics Commands
var (
	// Firmware Version. Read firmware version of the CRE module.
	// The firmware version returns 4 hexadecimal values (2 bytes) “ABCD”. Digits ABC are the main release number
	// and D is the revision number from the main release. “B” is a variant designator.
	// XBee and XBee-PRO ZB modules return:
	// 0x2xxx versions.
	// XBee and XBee-PRO ZNet modules return:
	// 0x1xxx versions. ZNet firmware is not compatible with ZB firmware.
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFF [read-only]
	atFirmwareVersion = ATCommand([2]byte{'V', 'R'})
	// Hardware Version. Read the hardware version of the module version of the module. This command can be used to distinguish among different hardware platforms. The upper byte returns a value that is unique to each module type. The lower byte indicates the hardware revision.
	// XBee ZB and XBee ZNet modules return the following (hexadecimal) values:
	// 0x19xx - XBee module
	// 0x1Axx - XBee-PRO module
	// Node Type: CRE
	// Parameter Range: 0 - 0xFFFF [read-only] 0x1Exx
	// Default: factory-set
	atHardwareVersion = ATCommand([2]byte{'H', 'V'})
	// Association Indication. Read information regarding last node join request:
	// 0x00 - Successfully formed or joined a network. (Coordinators form a network, routers and end devices join a network).
	// 0x21 - Scan found no PANs
	// 0x22 - Scan found no valid PANs based on current SC and ID settings
	// 0x23 - Valid Coordinator or Routers found, but they are not allowing joining (NJ expired)
	// 0x24 - No joinable beacons were found
	// 0x25 - Unexpected state, node should not be attempting to join at this time
	// 0x27 - Node Joining attempt failed (typically due to incompatible security settings)
	// 0x2A - Coordinator Start attempt failed
	// 0x2B - Checking for an existing coordinator
	// 0x2C - Attempt to leave the network failed
	// 0xAB - Attempted to join a device that did not respond.
	// 0xAC - Secure join error - network security key received unsecured
	// 0xAD - Secure join error - network security key not received
	// 0xAF - Secure join error - joining device does not have the right preconfigured link key
	// 0xFF - Scanning for a ZigBee network (routers and end devices)
	// NOTE: New non-zero AI values may be added in later firmware versions. Applications should read AI until it returns 0x00, indicating a successful startup (coordinator) or join (routers and end devices)
	atAssociationIndication = ATCommand([2]byte{'A', 'I'})
)

// Sleep Commands

// Execution Commands
var (
	// Apply Changes. Applies changes to all command registers causing queued
	// command register values to be applied. For example, changing the serial
	// interface rate with the BD command will not change the UART interface
	// rate until changes are applied with the AC command. The CN command and
	// 0x08 API command frame also apply changes.
	atApplyChanges = ATCommand([2]byte{'A', 'C'})
	// Write. Write parameter values to non-volatile memory so that parameter
	// modifications persist through subsequent resets.
	//
	// Note Once WR is issued, no additional characters should be sent to the
	// module until after the “OK\r” response is received. The WR command
	// should be used sparingly. The EM250 supports a limited number of write
	// cycles.
	atWrite = ATCommand([2]byte{'W', 'R'})
	// Restore Defaults. Restore module parameters to factory defaults.
	atRestoreDefaults = ATCommand([2]byte{'R', 'E'})
	// Software Reset. Reset module. Responds immediately with an OK status,
	// and then performs a software reset about two seconds later.
	atSoftwareReset = ATCommand([2]byte{'F', 'R'})
	// Node Discover. Discovers and reports all RF modules found. The following
	// information is reported for each
	// module discovered. SH<CR>
	// SL<CR>
	// NI<CR> (Variable length)
	// PARENT_NETWORK ADDRESS (2 Bytes)<CR> DEVICE_TYPE<CR> (1 Byte: 0=Coord, 1=Router, 2=End Device)
	// STATUS<CR> (1 Byte: Reserved)
	// PROFILE_ID<CR> (2 Bytes)
	// MANUFACTURER_ID<CR> (2 Bytes)
	// <CR>
	// After (NT * 100) milliseconds, the command ends by returning a <CR>. ND also accepts a Node Identifier (NI) as a parameter (optional). In this case, only a module that matches the supplied identifier will respond.
	// If ND is sent through the API, each response is returned as a separate AT_CMD_Response packet. The data consists of the above listed bytes without the carriage return delimiters. The NI string will end in a “0x00” null character. The radius of the ND command is set by the BH command.
	// Node Type: CRE
	atNodeDiscover = ATCommand([2]byte{'N', 'D'})
	// Active Scan. Scans the neighborhood for beacon responses. The ATAS command is only valid as a local command. Response frames are structured as:
	// AS_type – unsigned byte = 2 - ZB firmware uses a different format than Wi-Fi XBee, which is type 1
	// Channel – unsigned byte
	// PAN – unsigned word in big endian format
	// Extended PAN – eight unsigned bytes in bit endian format Allow Join – unsigned byte – 1 indicates join is enabled, 0 that it is disabled
	// Stack Profile – unsigned byte
	// LQI – unsigned byte, higher values are better
	// RSSI – signed byte, lower values are better
	atActiveScan = ATCommand([2]byte{'A', 'S'})
)

var standardDataRates = []int{1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200}
