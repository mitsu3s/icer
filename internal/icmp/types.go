package icmp

import "fmt"

// CodeInfo describes a single ICMP code within a type.
type CodeInfo struct {
	Code        uint8
	Description string
}

// TypeInfo describes an ICMP error message type and its valid codes.
type TypeInfo struct {
	Type        uint8
	Name        string
	Description string
	Codes       []CodeInfo
}

// SupportedTypes lists all ICMP error message types that icer can send.
var SupportedTypes = []TypeInfo{
	{
		Type:        3,
		Name:        "Destination Unreachable",
		Description: "Indicates that a destination is unreachable",
		Codes: []CodeInfo{
			{0, "Network Unreachable"},
			{1, "Host Unreachable"},
			{2, "Protocol Unreachable"},
			{3, "Port Unreachable"},
			{4, "Fragmentation Needed and DF Set"},
			{5, "Source Route Failed"},
			{6, "Destination Network Unknown"},
			{7, "Destination Host Unknown"},
			{8, "Source Host Isolated"},
			{9, "Network Administratively Prohibited"},
			{10, "Host Administratively Prohibited"},
			{11, "Network Unreachable for TOS"},
			{12, "Host Unreachable for TOS"},
			{13, "Communication Administratively Prohibited"},
			{14, "Host Precedence Violation"},
			{15, "Precedence Cutoff in Effect"},
		},
	},
	{
		Type:        5,
		Name:        "Redirect",
		Description: "Requests a host to change its routing",
		Codes: []CodeInfo{
			{0, "Redirect for Network"},
			{1, "Redirect for Host"},
			{2, "Redirect for TOS and Network"},
			{3, "Redirect for TOS and Host"},
		},
	},
	{
		Type:        11,
		Name:        "Time Exceeded",
		Description: "Indicates that a packet's lifetime has expired",
		Codes: []CodeInfo{
			{0, "TTL Exceeded in Transit"},
			{1, "Fragment Reassembly Time Exceeded"},
		},
	},
}

// typeIndex provides O(1) lookup by ICMP type number.
var typeIndex map[uint8]*TypeInfo

func init() {
	typeIndex = make(map[uint8]*TypeInfo, len(SupportedTypes))
	for i := range SupportedTypes {
		typeIndex[SupportedTypes[i].Type] = &SupportedTypes[i]
	}
}

// LookupType returns the TypeInfo for the given ICMP type, or nil if unsupported.
func LookupType(icmpType uint8) *TypeInfo {
	return typeIndex[icmpType]
}

// TypeName returns a human-readable name for the given ICMP type.
func TypeName(icmpType uint8) string {
	if t := LookupType(icmpType); t != nil {
		return t.Name
	}
	return fmt.Sprintf("Unknown (%d)", icmpType)
}

// ValidateTypeCode checks if the given type/code combination is valid.
func ValidateTypeCode(icmpType, code uint8) error {
	t := LookupType(icmpType)
	if t == nil {
		return fmt.Errorf("unsupported ICMP type: %d (supported: 3, 5, 11)", icmpType)
	}
	for _, c := range t.Codes {
		if c.Code == code {
			return nil
		}
	}
	return fmt.Errorf("invalid code %d for ICMP type %d (%s): valid range 0-%d",
		code, icmpType, t.Name, t.Codes[len(t.Codes)-1].Code)
}
