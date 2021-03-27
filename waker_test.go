package main

import (
	"io/ioutil"
	"testing"
)

// Not actually a test file, but an easy way to make the config file
// if your MAC address is 00:0a:95:9d:68:16
// fill the arg with []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16}
func TestWriteAddr(t *testing.T) {
	t.Skip()
	ioutil.WriteFile("config.addr", []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16}, 0777)
}
