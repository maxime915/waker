package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
)

// MagicPacket computes the magic packet to broadcast to activate the WakeOnLan
// device. See https://www.amd.com/system/files/TechDocs/20213.pdf
// address is expected to be a MAC address composed of 6 bytes, MagicPacket
// returns nil if address isn't valid.
func MagicPacket(address net.HardwareAddr) []byte {
	if len(address) != 6 {
		return nil
	}

	payload := make([]byte, 102)

	// add prefix
	i := 0
	for ; i < 6; i++ {
		payload[i] = 255
	}

	// repeat 16 times
	for c := 0; c < 16; c++ {
		for m := 0; m < 6; m++ {
			payload[i] = address[m]
			i++
		}
	}

	return payload
}

// ParseArguments checks the arguments of the program and creates an open TCP server,
// the parsed MAC address of the target and an open UDP connection. The two server
// must be closed by the caller after use. target must be a valid 6 bytes MAC address,
// addr must be a valid IP address / TCP port combo to listen to, broadcast must be a
// valid IP address / UDP port to write to. It is not checked whether broadcast actually
// correspond to the broadcast address of its network.
func ParseArguments(target, addr, broadcast string) (net.Listener, net.HardwareAddr, net.Conn, error) {
	// target -> no default value
	if len(target) == 0 {
		return nil, nil, nil, fmt.Errorf("the MAC address of the target is required")
	}

	targetAddr, err := net.ParseMAC(target)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(targetAddr) != 6 {
		return nil, nil, nil, fmt.Errorf("unsupported MAC address: %v", targetAddr)
	}

	// TCP server binding
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, nil, err
	}

	// UDP destination (broadcast)
	dest, err := net.Dial("udp", broadcast)
	if err != nil {
		return nil, nil, nil, err
	}

	return listener, targetAddr, dest, nil
}

func main() {
	target := flag.String("target", "", "6 bytes MAC address of the target")
	addr := flag.String("address", "", "server binding address (with port)")
	broadcast := flag.String("broadcast", "192.168.0.255:9", "UDP address to send the datagram to (with port)")
	verbose := flag.Bool("verbose", false, "print a confirmation message before serving")

	flag.Parse()

	listener, targetAddr, dest, err := ParseArguments(*target, *addr, *broadcast)
	payload := MagicPacket(targetAddr)

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()
	defer dest.Close()

	if *verbose {
		fmt.Println("waker is running")
		fmt.Println("\tTCP socket", listener.Addr().String())
		fmt.Println("\tMAC target", targetAddr.String())
		fmt.Println("\tUDP target", dest.RemoteAddr().String())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/wake", func(w http.ResponseWriter, _ *http.Request) {
		_, err := dest.Write(payload)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, err := fmt.Fprintln(w, "500 - Error while sending magic packet (see logs)")
			if err != nil {
				log.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, "200 - Magic packet send")
			if err != nil {
				log.Println(err)
			}
		}
	})

	done := make(chan struct{})
	mux.HandleFunc("/kill", func(w http.ResponseWriter, _ *http.Request) {
		close(done)
	})

	go func() {
		log.Fatal(http.Serve(listener, mux))
	}()

	<-done
}
