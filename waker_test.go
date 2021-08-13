package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MagicPacket is simple enough that it doesn't require any test

func TestParseArguments(tester *testing.T) {
	tester.Run("empty-target", func(t *testing.T) {
		_, _, _, err := ParseArguments("", "0:9009", "192.168.0.255:9")
		assert.EqualError(t, err, "the MAC address of the target is required")
	})
	tester.Run("empty-addr", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:00:00:00:00:00", "", "0:9")
		assert.NoError(t, err, "empty address means listening on loopback with next available port")
	})
	tester.Run("empty-broadcast", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:00:00:00:00:00", "0:9009", "")
		assert.EqualError(t, err, "dial udp: missing address")
	})
	tester.Run("empty-all", func(t *testing.T) {
		_, _, _, err := ParseArguments("", "", "")
		assert.Error(t, err, "only check for an error, which one is irrelevant")
	})

	tester.Run("bad-target-1", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:00:00:00:00", "", "0:9")
		assert.EqualError(t, err, "address 00:00:00:00:00: invalid MAC address", "only 5 bytes")
	})
	tester.Run("bad-target-2", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:00:00:00:00:00:00", "", "0:9")
		assert.EqualError(t, err, "address 00:00:00:00:00:00:00: invalid MAC address", "7 bytes")
	})
	tester.Run("bad-target-3", func(t *testing.T) {
		_, _, _, err := ParseArguments("Z0:00:00:00:00:00", "", "0:9")
		assert.EqualError(t, err, "address Z0:00:00:00:00:00: invalid MAC address", "7 bytes")
	})

	tester.Run("bad-addr-1", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0.0", "0:9")
		assert.EqualError(t, err, "listen tcp: address 0.0: missing port in address", "missing port")
	})
	tester.Run("bad-addr-2", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0.0.0.256:0", "0:9")
		assert.EqualError(t, err, "listen tcp: lookup 0.0.0.256: no such host", "256 is not allowed")
	})
	tester.Run("bad-addr-3", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0:-1", "0:9")
		assert.EqualError(t, err, "listen tcp: address -1: invalid port", "port must be >= 0")
	})

	tester.Run("bad-broadcast-1", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0:0", "0.0")
		assert.EqualError(t, err, "dial udp: address 0.0: missing port in address", "missing port")
	})
	tester.Run("bad-broadcast-2", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0:0", "0.0.0.256:0")
		assert.EqualError(t, err, "dial udp: lookup 0.0.0.256: no such host", "256 is not allowed")
	})
	tester.Run("bad-broadcast-3", func(t *testing.T) {
		_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", "0:0", "0:-1")
		assert.EqualError(t, err, "dial udp: address -1: invalid port", "port must be >= 0")
	})

	// NOTE: it is not guaranteed that any address other that loopback or localhost work on a specific device
	// NOTE: cannot guarantee that any port will be available on any device -> must use 0
	// NOTE: IPv6 is support by default on most OS

	addrValid := []string{"", ":0", "[::]:0", "0:0", "localhost:0"}
	broadcastValid := []string{":8", "[::]:9", "0:9", "localhost:9"}

	for i := range addrValid {
		for j := range broadcastValid {
			tester.Run(fmt.Sprintf("good-%d-%d", i, j), func(t *testing.T) {
				_, _, _, err := ParseArguments("00:A0:C9:14:C8:29", addrValid[i], broadcastValid[j])
				assert.NoError(t, err)
			})
		}
	}
}
