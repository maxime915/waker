// waker-http : start a server listening on HTTP to send WOL packets on demand

package serve_http

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/maxime915/waker/pkg/waker"
)

type VerbArguments struct {
	BindingAddress string `goptions:"--bind, description='Binding address & port for the HTTP server'"`
	Verbose        bool   `goptions:"-v, --verbose, description='Print server information on startup'"`
	Killable       bool   `goptions:"--killable, description='Creates a /kill route to terminate the server'"`
	Target         string `goptions:"-t, --target, obligatory, description='MAC address to wake up'"`
	Broadcast      string `goptions:"--broadcast, description='Broadcast address & port to send the packet'"`
}

func (va VerbArguments) Execute() {

	flag.Parse()

	listener, err := net.Listen("tcp", va.BindingAddress)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	if va.Verbose {
		fmt.Println("waker is running, killable:", va.Killable)
		fmt.Println("\tTCP socket", listener.Addr())
		fmt.Println("\ttarget", va.Target)
		fmt.Println("\tbroadcast", va.Broadcast)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/wake", func(w http.ResponseWriter, _ *http.Request) {
		err := waker.SendPacketTo(va.Target, va.Broadcast)
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
	})

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
