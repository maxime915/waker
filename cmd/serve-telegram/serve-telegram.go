// waker-telegram : Telegram bot to send WOL on demand

package serve_telegram

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maxime915/waker/pkg/waker"
	telegram "gopkg.in/tucnak/telebot.v3"
)

type VerbArguments struct {
	Token     string `goptions:"--token, obligatory, description='Telegram API token for the bot'"`
	Verbose   bool   `goptions:"-v, --verbose, description='Print server information on startup'"`
	Killable  bool   `goptions:"--killable, description='Creates a /kill route to terminate the server'"`
	Target    string `goptions:"-t, --target, obligatory, description='MAC address to wake up'"`
	Broadcast string `goptions:"-b, --broadcast, description='Broadcast address & port to send the packet'"`
}

func (va VerbArguments) Execute() {

	flag.Parse()

	packet, err := waker.NewMagicPacket(va.Target)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegram.NewBot(telegram.Settings{
		Token:  va.Token,
		Poller: &telegram.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	if va.Verbose {
		fmt.Println("waker is running, killable:", va.Killable)
		fmt.Println("\tpacket", packet)
		fmt.Println("\tbroadcast", va.Broadcast)
	}

	// listen to interrupts
	interrupted := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		close(interrupted)
	}()

	bot.Handle("/info", func(c telegram.Context) error {
		msg := fmt.Sprintf("Killable: %v\nTarget: %v\nBroadcast: %v\n", va.Killable, va.Target, va.Broadcast)
		_, err := bot.Send(c.Sender(), msg)
		return err
	})

	bot.Handle("/wake", func(c telegram.Context) error {
		err := packet.SendTo(va.Broadcast)
		if err != nil {
			log.Println(err)
			_, err = bot.Send(c.Sender(), "Error while sending magic packet (see logs)")
			if err != nil {
				log.Println(err)
			}
		} else {
			_, err = bot.Send(c.Sender(), "Magic packet sent")
			if err != nil {
				log.Println(err)
			}
		}
		return nil
	})

	done := make(chan struct{})
	bot.Handle("/kill", func(c telegram.Context) error {
		if va.Killable {
			_, err = bot.Send(c.Sender(), "Shutting down")
			if err != nil {
				log.Println(err)
			}
			close(done)
		} else {
			_, err = bot.Send(c.Sender(), "Server cannot be killed remotely")
			if err != nil {
				log.Println(err)
			}
		}

		return nil
	})

	bot.Handle(telegram.OnText, func(c telegram.Context) error {
		_, err := bot.Send(c.Sender(), "Unrecognized command")
		return err
	})

	go func() {
		bot.Start()
		close(done)
	}()

	// when interrupted, stop and wait until done
	// when done, proceed
	select {
	case <-interrupted:
		bot.Stop()
	case <-done:
	}
}
