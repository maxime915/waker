package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const broadcast = "192.168.0.255:9"

func Wake(macAddr []byte) error {
	addr, err := net.ResolveUDPAddr("udp", broadcast)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	defer conn.Close()

	msg := make([]byte, 102)

	// add prefix
	i := 0
	for ; i < 6; i++ {
		msg[i] = 255
	}

	// repeat 16 times
	for c := 0; c < 16; c++ {
		for m := 0; m < 6; m++ {
			msg[i] = macAddr[m]
			i++
		}
	}

	count, err := conn.Write(msg)
	if err != nil {
		return err
	}
	if count != len(msg) {
		return fmt.Errorf("expected to write %d more bytes", len(msg)-count)
	}

	return nil
}

func getHandler(msgAddr []byte) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := Wake(msgAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 - Error while sending magic packet: %s\n", err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "200 - Magic packet send\n")
		}
	}
}

func main() {
	addr := "192.168.0.15:9009"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	config := "config.addr"
	if len(os.Args) > 2 {
		config = os.Args[2]
	}

	// get mac address from config file
	msgAddr, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal("unable to open the config file")
	}
	if len(msgAddr) != 6 {
		log.Fatalf("invalid MAC address, expected 6 byte, found: %X\n", msgAddr)
	}

	http.HandleFunc("/wake", getHandler(msgAddr))
	log.Fatal(http.ListenAndServe(addr, nil))
}
