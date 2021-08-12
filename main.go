package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

const broadcast = "192.168.0.255:9"

func Wake(macAddr net.HardwareAddr) error {
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

func getHandler(msgAddr net.HardwareAddr) func(w http.ResponseWriter, _ *http.Request) {
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
	var targetStr, addr string
	var verbose bool

	flag.StringVar(&targetStr, "target", "", "MAC address of the target")
	flag.StringVar(&addr, "address", "127.0.0.1:0", "server binding address and port")
	flag.BoolVar(&verbose, "verbose", false, "print a confirmation message before serving")
	flag.Parse()

	// check target
	if len(targetStr) == 0 {
		log.Fatal("the MAC address of the target is required")
	}

	target, err := net.ParseMAC(targetStr)
	if err != nil {
		log.Fatal("could not parse target: ", err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		fmt.Println("waker is listening on", listener.Addr().String(), "for target", target.String())
	}

	http.HandleFunc("/wake", getHandler(target))
	log.Fatal(http.Serve(listener, nil))
}
