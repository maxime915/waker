// remote-http : small CLI tool to send a request to a waker server (launched via waker-http)

package remote_http

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type VerbArguments struct {
	HostName string `goptions:"-n, --hostname, obligatory, description='Hostname or address (with port) of the HTTP server'"`
	Target   string `goptions:"-t, --target, description='Target to wake. Will wake the default if not provided'"`
}

func (va VerbArguments) Execute() {

	if len(va.HostName) == 0 {
		os.Stderr.WriteString("empty host given\n")
		os.Exit(1)
	}

	url := "http://" + va.HostName + "/wake"
	if len(va.Target) > 0 {
		if _, err := net.ParseMAC(va.Target); err != nil {
			log.Fatal("invalid target: " + err.Error())
		}
		url += "/" + va.Target
	}

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	message, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unable to send packet: %v", string(message))
	}

	fmt.Print(string(message))
}
