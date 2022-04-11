// remote-http : small CLI tool to send a request to a waker server (launched via waker-http)

package remote_http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type VerbArguments struct {
	HostName string `goptions:"-n, --hostname, obligatory, description='Hostname or address (with port) of the HTTP server'"`
}

func (va VerbArguments) Execute() {

	if len(va.HostName) == 0 {
		os.Stderr.WriteString("empty host given\n")
		os.Exit(1)
	}

	response, err := http.Get("http://" + va.HostName + "/wake")
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	message, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unable to send packet: %v", string(message))
	}

	fmt.Print(string(message))
}
