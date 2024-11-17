package internal

import (
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mitsu3s/icer/config"
)

// Exceeded sends an ICMP Time Exceeded packet with an Error Ping packet to the specified destination
func Exceeded(code uint8) error {
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

	// Generate ICMP Time Exceeded packet
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
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeTimeExceeded, code),
		Id:       0x1234,
		Seq:      1,
	}

	// Generate the IP header of the packet that caused the error
	errorIP := layers.IPv4{
		Version:  4,
		IHL:      5,
		SrcIP:    net.ParseIP(cfg.ErrorIP.SrcIP),
		DstIP:    net.ParseIP(cfg.ErrorIP.DstIP),
		Protocol: layers.IPProtocolICMPv4,
		TTL:      1, // Original TTL that would expire
	}

	errorICMP := layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0), // Echo request code is always 0
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
		&errorIP,
		&errorICMP,
		gopacket.Payload([]byte("")),
	); err != nil {
		log.Fatalf("Failed to serialize packet: %v", err)
	}

	packetData := buffer.Bytes()

	// Create destination address structure
	addr := &syscall.SockaddrInet4{}
	copy(addr.Addr[:], ip.DstIP.To4())

	// Send packet via raw socket
	err = syscall.Sendto(socket, packetData, 0, addr)
	if err != nil {
		log.Fatalf("Failed to send packet: %v", err)
	}

	log.Printf("Sent ICMP Time Exceeded packet with code %d to %s", code, cfg.RealIP.DstIP)

	return nil
}
