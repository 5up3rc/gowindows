package gowindows

import (
	"syscall"

	"unsafe"

	"os"

	"fmt"

	"golang.org/x/sys/windows"
)

var (
	iphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	getIpForwardTable    = iphlpapi.NewProc("GetIpForwardTable")
	createIpForwardEntry = iphlpapi.NewProc("CreateIpForwardEntry")
	deleteIpForwardEntry = iphlpapi.NewProc("DeleteIpForwardEntry")
	notifyAddrChange     = iphlpapi.NewProc("NotifyAddrChange")
	notifyRouteChange    = iphlpapi.NewProc("NotifyRouteChange")
)

// IP_ADAPTER_ADDRESSES_LH
// https://docs.microsoft.com/zh-cn/windows/desktop/api/iptypes/ns-iptypes-_ip_adapter_addresses_lh
/*
typedef struct _IP_ADAPTER_ADDRESSES_LH {
  union {
    ULONGLONG Alignment;
    struct {
      ULONG    Length;
      IF_INDEX IfIndex;
    };
  };
  struct _IP_ADAPTER_ADDRESSES_LH    *Next;
  PCHAR                              AdapterName;
  PIP_ADAPTER_UNICAST_ADDRESS_LH     FirstUnicastAddress;
  PIP_ADAPTER_ANYCAST_ADDRESS_XP     FirstAnycastAddress;
  PIP_ADAPTER_MULTICAST_ADDRESS_XP   FirstMulticastAddress;
  PIP_ADAPTER_DNS_SERVER_ADDRESS_XP  FirstDnsServerAddress;
  PWCHAR                             DnsSuffix;
  PWCHAR                             Description;
  PWCHAR                             FriendlyName;
  BYTE                               PhysicalAddress[MAX_ADAPTER_ADDRESS_LENGTH];
  ULONG                              PhysicalAddressLength;
  union {
    ULONG Flags;
    struct {
      ULONG DdnsEnabled : 1;
      ULONG RegisterAdapterSuffix : 1;
      ULONG Dhcpv4Enabled : 1;
      ULONG ReceiveOnly : 1;
      ULONG NoMulticast : 1;
      ULONG Ipv6OtherStatefulConfig : 1;
      ULONG NetbiosOverTcpipEnabled : 1;
      ULONG Ipv4Enabled : 1;
      ULONG Ipv6Enabled : 1;
      ULONG Ipv6ManagedAddressConfigurationSupported : 1;
    };
  };
  ULONG                              Mtu;
  IFTYPE                             IfType;
  IF_OPER_STATUS                     OperStatus;
  IF_INDEX                           Ipv6IfIndex;
  ULONG                              ZoneIndices[16];
  PIP_ADAPTER_PREFIX_XP              FirstPrefix;
  ULONG64                            TransmitLinkSpeed;
  ULONG64                            ReceiveLinkSpeed;
  PIP_ADAPTER_WINS_SERVER_ADDRESS_LH FirstWinsServerAddress;
  PIP_ADAPTER_GATEWAY_ADDRESS_LH     FirstGatewayAddress;
  ULONG                              Ipv4Metric;
  ULONG                              Ipv6Metric;
  IF_LUID                            Luid;
  SOCKET_ADDRESS                     Dhcpv4Server;
  NET_IF_COMPARTMENT_ID              CompartmentId;
  NET_IF_NETWORK_GUID                NetworkGuid;
  NET_IF_CONNECTION_TYPE             ConnectionType;
  TUNNEL_TYPE                        TunnelType;
  SOCKET_ADDRESS                     Dhcpv6Server;
  BYTE                               Dhcpv6ClientDuid[MAX_DHCPV6_DUID_LENGTH];
  ULONG                              Dhcpv6ClientDuidLength;
  ULONG                              Dhcpv6Iaid;
  PIP_ADAPTER_DNS_SUFFIX             FirstDnsSuffix;
} IP_ADAPTER_ADDRESSES_LH, *PIP_ADAPTER_ADDRESSES_LH;
*/
//TODO: 结构应该存在问题，GUID 之前应该有字段不对！
type IpAdapterAddresses struct {
	Length                uint32
	IfIndex               uint32
	Next                  *IpAdapterAddresses
	AdapterName           *byte
	FirstUnicastAddress   *windows.IpAdapterUnicastAddress
	FirstAnycastAddress   *windows.IpAdapterAnycastAddress
	FirstMulticastAddress *windows.IpAdapterMulticastAddress
	FirstDnsServerAddress *windows.IpAdapterDnsServerAdapter
	DnsSuffix             *uint16
	Description           *uint16
	FriendlyName          *uint16
	PhysicalAddress       [syscall.MAX_ADAPTER_ADDRESS_LENGTH]byte
	PhysicalAddressLength uint32
	Flags                 uint32
	Mtu                   uint32
	IfType                uint32
	OperStatus            uint32

	// 以下是 windows xp sp1 之后添加的
	ipv6IfIndex uint32
	zoneIndices [16]uint32
	firstPrefix *windows.IpAdapterPrefix

	// 以下是 windows Vista 之后添加的
	transmitLinkSpeed      uint64
	receiveLinkSpeed       uint64
	firstWinsServerAddress *IpAdapterWinsServerAddress
	firstGatewayAddress    *IpAdapterGatewayAddress
	ipv4Metric             uint32
	ipv6Metric             uint32
	luid                   IfLuid
	dhcpv4Server           windows.SocketAddress
	compartmentId          CompartmentId
	networkGuid            NetworkGuid //64位 错误位置
	connectionType         ConnectionType
	tunnelType             TunnelType
	dhcpv6Server           windows.SocketAddress
	dhcpv6ClientDuid       [MAX_DHCPV6_DUID_LENGTH]byte
	dhcpv6ClientDuidLength uint32
	dhcpv6Iaid             uint32

	// 以下是 windows Vista SP1 及 windows server 2008 之后添加的
	firstDnsSuffix *IpAdapterDnsSuffix
}

//typedef struct _IP_ADAPTER_WINS_SERVER_ADDRESS_LH {
//    union {
//        ULONGLONG Alignment;
//        struct {
//            ULONG Length;
//            DWORD Reserved;
//        };
//    };
//    struct _IP_ADAPTER_WINS_SERVER_ADDRESS_LH *Next;
//    SOCKET_ADDRESS Address;
//} IP_ADAPTER_WINS_SERVER_ADDRESS_LH, *PIP_ADAPTER_WINS_SERVER_ADDRESS_LH;
type IpAdapterWinsServerAddress struct {
	Length   uint32
	Reserved int32
	Next     *IpAdapterWinsServerAddress
	Address  windows.SocketAddress
}

type IpAdapterGatewayAddress struct {
	Length   uint32
	Reserved int32
	Next     *IpAdapterGatewayAddress
	Address  windows.SocketAddress
}

func (aa *IpAdapterAddresses) GetLuid() (IfLuid, error) {
	tz := aa.Length
	fz := unsafe.Offsetof(aa.luid) + unsafe.Sizeof(aa.luid)

	// 判断结构是否包含了指定的字段
	// 不同版本的 windows 包含的字段不同，老版本的不包含新版本的字段。
	if tz < uint32(fz) {
		return IfLuid(0), fmt.Errorf("Length(%v)<%v", tz, fz)
	}

	return aa.luid, nil
}
func (aa *IpAdapterAddresses) GetNetworkGuid() (NetworkGuid, error) {
	tz := aa.Length
	fz := unsafe.Offsetof(aa.networkGuid) + unsafe.Sizeof(aa.networkGuid)

	// 判断结构是否包含了指定的字段
	// 不同版本的 windows 包含的字段不同，老版本的不包含新版本的字段。
	if tz < uint32(fz) {
		return NetworkGuid{}, fmt.Errorf("Length(%v)<%v", tz, fz)
	}

	return aa.networkGuid, nil
}

func (aa *IpAdapterAddresses) GetFriendlyName() string {
	// C:/Go/src/net/interface_windows.go:77
	return syscall.UTF16ToString((*(*[10000]uint16)(unsafe.Pointer(aa.FriendlyName)))[:])
}
func (aa *IpAdapterAddresses) GetDescription() string {
	// C:/Go/src/net/interface_windows.go:77
	return syscall.UTF16ToString((*(*[10000]uint16)(unsafe.Pointer(aa.Description)))[:])
}

func (aa *IpAdapterAddresses) GetGatewayAddress() ([]*IpAdapterGatewayAddress, error) {
	tz := aa.Length
	fz := unsafe.Offsetof(aa.firstGatewayAddress) + unsafe.Sizeof(aa.firstGatewayAddress)

	// 判断结构是否包含了指定的字段
	// 不同版本的 windows 包含的字段不同，老版本的不包含新版本的字段。
	if tz < uint32(fz) {
		return nil, fmt.Errorf("Length(%v)<%v", tz, fz)
	}

	res := make([]*IpAdapterGatewayAddress, 0, 1)
	ga := aa.firstGatewayAddress

	for ga != nil {
		res = append(res, ga)
		ga = ga.Next
	}

	return res, nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/iphlpapi/nf-iphlpapi-getadaptersaddresses
func AdapterAddresses() ([]*IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, GAA_FLAG_INCLUDE_PREFIX|GAA_FLAG_INCLUDE_WINS_INFO|GAA_FLAG_INCLUDE_GATEWAYS, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*IpAdapterAddresses
	for aa := (*IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

func GetIpForwardTable() ([]MibIpForwardRow, error) {
	buf := []byte{0}
	bufSize := uint32(len(buf))
	var r1 uintptr
	var e1 error
	for i := 0; i < 10; i++ {
		buf = make([]byte, bufSize)
		r1, _, e1 = getIpForwardTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bufSize)), 0)
		if r1 == ERROR_INSUFFICIENT_BUFFER {
			// 空间不足
			continue
		}

		break
	}

	if r1 != 0 {
		if e1 != ERROR_SUCCESS {
			return nil, e1
		} else {
			return nil, fmt.Errorf("r1:%v", r1)
		}
	}

	table := (*MibIpForwardTable)(unsafe.Pointer(&buf[0]))
	rows := table.Table[:]
	err := ChangeSliceSize(&rows, int(table.NumEntries), int(table.NumEntries))
	if err != nil {
		return nil, fmt.Errorf("ChangeSliceSize, %v", err)
	}

	res := make([]MibIpForwardRow, len(rows))
	copy(res, rows)
	return res, nil
}

func CreateIpForwardEntry(row *MibIpForwardRow) error {
	r1, _, e1 := createIpForwardEntry.Call(uintptr(unsafe.Pointer(row)))
	if r1 != 0 {
		if e1 != ERROR_SUCCESS {
			return e1
		} else {
			return fmt.Errorf("r1:%v", r1)
		}
	}

	return nil
}

// 必须提供以下成员：dwForwardIfIndex，dwForwardDest，dwForwardMask，dwForwardNextHop和dwForwardProto
func DeleteIpForwardEntry(row *MibIpForwardRow) error {
	r1, _, e1 := deleteIpForwardEntry.Call(uintptr(unsafe.Pointer(row)))
	if r1 != 0 {
		if e1 != ERROR_SUCCESS {
			return e1
		} else {
			return fmt.Errorf("r1:%v", r1)
		}
	}

	return nil
}

//https://docs.microsoft.com/en-us/windows/desktop/api/iphlpapi/nf-iphlpapi-notifyaddrchange
//DWORD NotifyAddrChange(
//  PHANDLE      Handle,
//  LPOVERLAPPED overlapped
//);
func NotifyAddrChange(handle *Handle, overlapped *Overlapped) error {
	r1, _, e1 := notifyAddrChange.Call(uintptr(unsafe.Pointer(handle)), uintptr(unsafe.Pointer(overlapped)))
	if handle == nil && overlapped == nil {
		if r1 == NO_ERROR {
			return nil
		}
	} else {
		if r1 == ERROR_IO_PENDING {
			return nil
		}
	}

	if e1 != ERROR_SUCCESS {
		return e1
	} else {
		return fmt.Errorf("r1:%v", r1)
	}
}

//DWORD NotifyRouteChange(
//  PHANDLE      Handle,
//  LPOVERLAPPED overlapped
//);
//https://docs.microsoft.com/en-us/windows/desktop/api/iphlpapi/nf-iphlpapi-notifyroutechange
func NotifyRouteChange(handle *Handle, overlapped *Overlapped) error {
	r1, _, e1 := notifyRouteChange.Call(uintptr(unsafe.Pointer(handle)), uintptr(unsafe.Pointer(overlapped)))
	if handle == nil && overlapped == nil {
		if r1 == NO_ERROR {
			return nil
		}
	} else {
		if r1 == ERROR_IO_PENDING {
			return nil
		}
	}

	if e1 != ERROR_SUCCESS {
		return e1
	} else {
		return fmt.Errorf("r1:%v", r1)
	}
}
