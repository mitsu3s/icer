package icmp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// SendParams contains all parameters needed to craft and send an ICMP error packet.
type SendParams struct {
	// ICMP type (3, 5, or 11)
	Type uint8
	// ICMP code
	Code uint8

	// Outer (real) IP header: src/dst of the ICMP error message itself
	SrcIP string
	DstIP string

	// Inner (error) IP header: the "original packet" embedded in the ICMP error body
	ErrorSrcIP string
	ErrorDstIP string

	// IP header options
	TTL uint8

	// Behavior flags
	DryRun  bool
	Verbose bool
}

// Send crafts and sends an ICMP error message based on the given parameters.
func Send(p *SendParams) error {
	typeName := TypeName(p.Type)

	if err := ValidateTypeCode(p.Type, p.Code); err != nil {
		return err
	}

	if p.Verbose {
		log.Printf("[VERBOSE] Preparing %s (Type=%d, Code=%d)", typeName, p.Type, p.Code)
		log.Printf("[VERBOSE] Outer: %s -> %s", p.SrcIP, p.DstIP)
		log.Printf("[VERBOSE] Inner: %s -> %s", p.ErrorSrcIP, p.ErrorDstIP)
		log.Printf("[VERBOSE] TTL=%d", p.TTL)
	}

	if p.DryRun {
		log.Printf("[DRY RUN] Would send %s (Type=%d, Code=%d) from %s to %s",
			typeName, p.Type, p.Code, p.SrcIP, p.DstIP)
		return nil
	}

	packetData, err := buildPacket(p)
	if err != nil {
		return fmt.Errorf("failed to build packet: %w", err)
	}

	if err := sendRaw(p.DstIP, packetData); err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}

	log.Printf("Sent %s (Type=%d, Code=%d) to %s", typeName, p.Type, p.Code, p.DstIP)
	return nil
}

// buildPacket constructs the full IP+ICMP packet bytes.
func buildPacket(p *SendParams) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// Outer IPv4 header
	outerIP := &layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    net.ParseIP(p.SrcIP),
		DstIP:    net.ParseIP(p.DstIP),
		Protocol: layers.IPProtocolICMPv4,
		TTL:      p.TTL,
	}

	// ICMP error header
	icmpLayer := &layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(p.Type, p.Code),
		Id:       0x1234,
		Seq:      1,
	}

	// Build type-specific payload between ICMP header and inner packet
	var extraPayload []byte
	if p.Type == 3 && p.Code == 4 {
		// Fragmentation Needed: next-hop MTU in bytes 6-7 of the ICMP header
		mtu := uint16(1500)
		extraPayload = make([]byte, 2)
		binary.BigEndian.PutUint16(extraPayload, mtu)
		if p.Verbose {
			log.Printf("[VERBOSE] Setting next-hop MTU to %d (Code 4)", mtu)
		}
	}

	// Inner IPv4 header: the "original packet" that caused the error
	innerTTL := p.TTL
	if p.Type == 11 {
		innerTTL = 1 // TTL that would have expired
	}

	innerIP := &layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    net.ParseIP(p.ErrorSrcIP),
		DstIP:    net.ParseIP(p.ErrorDstIP),
		Protocol: layers.IPProtocolICMPv4,
		TTL:      innerTTL,
	}

	// Inner ICMP echo request (the "original" ICMP packet)
	innerICMP := &layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
		Id:       0x5678,
		Seq:      1,
	}

	if err := gopacket.SerializeLayers(buf, opts,
		outerIP,
		icmpLayer,
		innerIP,
		innerICMP,
		gopacket.Payload(extraPayload),
	); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// sendRaw opens a raw socket, sends the packet, and closes the socket.
func sendRaw(dstIP string, data []byte) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		return fmt.Errorf("raw socket: %w (are you root?)", err)
	}
	defer syscall.Close(fd)

	ip := net.ParseIP(dstIP).To4()
	if ip == nil {
		return fmt.Errorf("invalid destination IP: %s", dstIP)
	}

	addr := &syscall.SockaddrInet4{}
	copy(addr.Addr[:], ip)

	return syscall.Sendto(fd, data, 0, addr)
}
