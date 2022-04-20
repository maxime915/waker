package waker

import (
	"fmt"
	"net"
)

type WOLPacket struct {
	address net.HardwareAddr
	body    []byte // contains the magic bytes, the address & the optional password
}

// NewMagicPacketWithPassword computes the magic packet to broadcast to activate
// the WakeOnLan device. See https://www.amd.com/system/files/TechDocs/20213.pdf
// address is expected to be a valid MAC address, password should be 0, 4 or 6 bytes
func NewMagicPacketWithPassword(target string, password []byte) (WOLPacket, error) {
	address, err := net.ParseMAC(target)
	if err != nil {
		return WOLPacket{}, err
	}

	if len(password) != 0 && len(password) != 4 && len(password) != 6 {
		return WOLPacket{}, fmt.Errorf("password should be 4 or 6 bytes")
	}

	// empty slice with full capacity
	payloadSize := 6 + 16*len(address) + len(password)
	payload := make([]byte, 0, payloadSize)

	// add the 6 magic bytes
	payload = append(payload, 255, 255, 255, 255, 255, 255)

	// repeat address 16 times for body
	for c := 0; c < 16; c++ {
		payload = append(payload, address...)
	}

	// add password
	if len(password) > 0 {
		payload = append(payload, password...)
	}

	return WOLPacket{address, payload}, nil
}

// NewMagicPacket computes the magic packet to broadcast to activate the WakeOnLan
// device. See https://www.amd.com/system/files/TechDocs/20213.pdf
// address is expected to be a valid MAC address
func NewMagicPacket(address string) (WOLPacket, error) {
	return NewMagicPacketWithPassword(address, nil)
}

// SendTo sends the magic packet to the given broadcast address
func (packet WOLPacket) SendTo(broadcast string) error {
	// UDP destination (broadcast)
	dest, err := net.Dial("udp", broadcast)
	if err != nil {
		return err
	}

	defer dest.Close()

	_, err = dest.Write(packet.body)

	return err
}

func (packet WOLPacket) String() string {
	return fmt.Sprintf("Magic Packet targetting %v", packet.address)
}

// SendPacketTo creates a magic packet for the given target and sends its to
// the given broadcast address
func SendPacketTo(target, broadcast string) error {
	if len(target) == 0 {
		return fmt.Errorf("the MAC address of the target is required")
	}

	// create packet
	packet, err := NewMagicPacket(target)
	if err != nil {
		return err
	}

	return packet.SendTo(broadcast)
}
