package inbound

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/qauzy/netat/common/nnip"
	C "github.com/qauzy/netat/constant"
	"github.com/qauzy/netat/transport/socks5"
)

func parseSocksAddr(target socks5.Addr) *C.Metadata {
	metadata := &C.Metadata{}

	switch target[0] {
	case socks5.AtypDomainName:
		// trim for FQDN
		metadata.Host = strings.TrimRight(string(target[2:2+target[1]]), ".")
		metadata.DstPort = uint16((int(target[2+target[1]]) << 8) | int(target[2+target[1]+1]))
	case socks5.AtypIPv4:
		metadata.DstIP = nnip.IpToAddr(net.IP(target[1 : 1+net.IPv4len]))
		metadata.DstPort = uint16((int(target[1+net.IPv4len]) << 8) | int(target[1+net.IPv4len+1]))
	case socks5.AtypIPv6:
		ip6, _ := netip.AddrFromSlice(target[1 : 1+net.IPv6len])
		metadata.DstIP = ip6.Unmap()
		metadata.DstPort = uint16((int(target[1+net.IPv6len]) << 8) | int(target[1+net.IPv6len+1]))
	}

	return metadata
}

func parseHTTPAddr(request *http.Request) *C.Metadata {
	host := request.URL.Hostname()
	port := request.URL.Port()
	if port == "" {
		port = "80"
	}

	// trim FQDN (#737)
	host = strings.TrimRight(host, ".")

	var uint16Port uint16
	if port, err := strconv.ParseUint(port, 10, 16); err == nil {
		uint16Port = uint16(port)
	}

	metadata := &C.Metadata{
		NetWork: C.TCP,
		Host:    host,
		DstIP:   netip.Addr{},
		DstPort: uint16Port,
	}

	ip, err := netip.ParseAddr(host)
	if err == nil {
		metadata.DstIP = ip
	}

	return metadata
}

func parseAddr(addr net.Addr) (netip.Addr, uint16, error) {
	// Filter when net.Addr interface is nil
	if addr == nil {
		return netip.Addr{}, 0, errors.New("nil addr")
	}
	if rawAddr, ok := addr.(interface{ RawAddr() net.Addr }); ok {
		ip, port, err := parseAddr(rawAddr.RawAddr())
		if err == nil {
			return ip, port, err
		}
	}
	addrStr := addr.String()
	host, port, err := net.SplitHostPort(addrStr)
	if err != nil {
		return netip.Addr{}, 0, err
	}

	var uint16Port uint16
	if port, err := strconv.ParseUint(port, 10, 16); err == nil {
		uint16Port = uint16(port)
	}

	ip, err := netip.ParseAddr(host)
	return ip, uint16Port, err
}
