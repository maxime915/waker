// waker : send one magic packet to a target (MAC address) on a network

package send_wol

import (
	"fmt"
	"log"

	"github.com/maxime915/waker/pkg/waker"
)

type VerbArguments struct {
	Target    string `goptions:"-t, --target, obligatory, description='MAC address to wake up'"`
	Broadcast string `goptions:"-b, --broadcast, description='Broadcast address & port to send the packet'"`
}

func (va VerbArguments) Execute() {
	err := waker.SendPacketTo(va.Target, va.Broadcast)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Magic packet sent")
}
