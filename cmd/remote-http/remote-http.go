// remote-http : small CLI tool to send a request to a waker server (launched via waker-http)

package remote_http

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type VerbArguments struct {
	HostName string `goptions:"-n, --hostname, obligatory, description='Hostname or address (with port) of the HTTP server'"`
	Target   string `goptions:"-t, --target, description='Target to wake. Will wake the default if not provided'"`
}

func buildUrl(hostname, target string) (url string, err error) {
	// check scheme and hostname
	if strings.HasPrefix(hostname, "http://") || strings.HasPrefix(hostname, "https://") {
		url = hostname
	} else {
		url = "http://" + hostname
	}

	// add first segment
	url += "/wake"

	// default route
	if len(target) == 0 {
		return url, nil
	}

	// check target
	_, err = net.ParseMAC(target)
	if err != nil {
		return "", err
	}

	url += "/" + target
	return url, nil
}

func (va VerbArguments) Execute() {

	if len(va.HostName) == 0 {
		os.Stderr.WriteString("empty host given\n")
		os.Exit(1)
	}

	url, err := buildUrl(va.HostName, va.Target)
	if err != nil {
		log.Fatal("invalid target: " + err.Error())
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
