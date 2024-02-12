// waker-http : start a server listening on HTTP to send WOL packets on demand

package serve_http

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/maxime915/waker/pkg/waker"
)

type VerbArguments struct {
	BindingAddress string `goptions:"--bind, description='Binding address & port for the HTTP server'"`
	Verbose        bool   `goptions:"-v, --verbose, description='Print server information on startup'"`
	Killable       bool   `goptions:"--killable, description='Creates a /kill route to terminate the server'"`
	Target         string `goptions:"-t, --target, obligatory, description='A CLS of MAC address to wake up'"`
	Broadcast      string `goptions:"--broadcast, description='Broadcast address & port to send the packet'"`
}

func wakerFn(target, broadcast string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := waker.SendPacketTo(target, broadcast)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err := fmt.Fprintln(w, "500 - Error while sending magic packet (see logs)")
			if err != nil {
				log.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, "200 - Magic packet sent")
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func badRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := fmt.Fprintln(w, "400 - Default route unsupported : more than one target possible")
	if err != nil {
		log.Println(err)
	}
}

func (va VerbArguments) Execute() {

	flag.Parse()

	var lastError error
	targetLst := strings.Split(va.Target, ",")
	for idx, addr := range targetLst {
		_, err := waker.NewMagicPacket(addr)
		if err != nil {
			log.Printf("entry %s (at pos %d) is invalid\n", addr, idx)
			lastError = err
		}
	}
	if lastError != nil { // at least one error found
		log.Panic("waker will not start with invalid entries")
	}
	if len(targetLst) == 0 {
		log.Panic("no valid entry found")
	}

	listener, err := net.Listen("tcp", va.BindingAddress)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	if va.Verbose {
		fmt.Println("waker is running, killable:", va.Killable)
		fmt.Println("\tTCP socket", listener.Addr())
		fmt.Println("\ttarget list", targetLst)
		fmt.Println("\tbroadcast", va.Broadcast)
	}

	mux := http.NewServeMux()

	// default route : OK if only 1 target only
	if len(targetLst) == 1 {
		mux.HandleFunc("/wake", wakerFn(targetLst[0], va.Broadcast))
	} else {
		mux.HandleFunc("/wake", badRequest)
	}

	// add a distinct route for each target
	for _, target := range targetLst {
		mux.HandleFunc(fmt.Sprintf("/wake/%s", target), wakerFn(target, va.Broadcast))
	}

	done := make(chan struct{})
	mux.HandleFunc("/kill", func(w http.ResponseWriter, _ *http.Request) {
		if va.Killable {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, "200 - shutting down")
			if err != nil {
				log.Println(err)
			}
			close(done)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, err := fmt.Fprintln(w, "400 - server cannot be killed remotely")
			if err != nil {
				log.Println(err)
			}
		}
	})

	go func() {
		log.Fatal(http.Serve(listener, mux))
	}()

	<-done
}
