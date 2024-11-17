package internal

import (
	"encoding/binary"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mitsu3s/icer/config"
)

// Unreachable sends an ICMP Destination Unreachable packet with a Error Ping packet to the specified destination
func Unreachable(code uint8) error {
	// Get config
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	// Create a raw socket
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		log.Fatalf("Failed to create raw socket: %v", err)
	}
	defer syscall.Close(socket)

	// Generate ICMP Destination Unreachable packet
	buffer := gopacket.NewSerializeBuffer()

	ip := layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    net.ParseIP(cfg.RealIP.SrcIP),
		DstIP:    net.ParseIP(cfg.RealIP.DstIP),
		Protocol: layers.IPProtocolICMPv4,
		TTL:      64,
	}

	icmp := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeDestinationUnreachable, code),
		Id:       0x1234,
		Seq:      1,
	}

	// If code is 4 (Fragmentation Needed and DF Set), specify the MTU size
	var payload []byte
	if code == 4 {
		mtu := uint16(1500) // Example MTU size; adjust as necessary
		payload = make([]byte, 2)
		binary.BigEndian.PutUint16(payload, mtu)
		log.Printf("Setting MTU size to %d for code 4 (Fragmentation Needed)", mtu)
	} else {
		payload = []byte("")
	}

	// Generate ICMP Echo packet that cause errors
	echoIP := layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    net.ParseIP(cfg.ErrorIP.SrcIP),
		DstIP:    net.ParseIP(cfg.ErrorIP.DstIP),
		Protocol: layers.IPProtocolICMPv4,
		TTL:      64,
	}

	echoICMP := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, code),
		Id:       0x5678,
		Seq:      1,
	}

	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// Serialize and write to buffer
	if err = gopacket.SerializeLayers(buffer, options,
		&ip,
		&icmp,
		&echoIP,
		&echoICMP,
		gopacket.Payload(payload),
	); err != nil {
		log.Fatalf("Failed to serialize packet: %v", err)
	}

	packetData := buffer.Bytes()

	// Create of destination address structure
	addr := &syscall.SockaddrInet4{}
	copy(addr.Addr[:], ip.DstIP.To4())

	// Send packets via raw socket
	err = syscall.Sendto(socket, packetData, 0, addr)
	if err != nil {
		log.Fatalf("Failed to send packet: %v", err)
	}

	log.Printf("Sent ICMP Destination Unreachable packet to %s", cfg.RealIP.DstIP)

	return nil
}
