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

	badger "github.com/dgraph-io/badger/v3"
	"github.com/maxime915/waker/pkg/waker"
	telegram "gopkg.in/tucnak/telebot.v3"
)

const user_key = "user_key"

type VerbArguments struct {
	Token     string `goptions:"--token, obligatory, description='Telegram API token for the bot'"`
	Verbose   bool   `goptions:"-v, --verbose, description='Print server information on startup'"`
	Killable  bool   `goptions:"--killable, description='Creates a /kill route to terminate the server'"`
	Target    string `goptions:"-t, --target, obligatory, description='MAC address to wake up'"`
	Broadcast string `goptions:"-b, --broadcast, description='Broadcast address & port to send the packet'"`
	StorePath string `goptions:"--store-path, description='Path to store bot data'"`
	InMemory  bool   `goptions:"--in-memory, description='Do not use disk to retain bot data'"`
}

func (va VerbArguments) Execute() {

	flag.Parse()

	if va.InMemory != (len(va.StorePath) == 0) {
		log.Fatalf("exactly only one of --store-path='%v' , --in-memory='%v' must be set", va.StorePath, va.InMemory)
	}

	if !va.InMemory && len(va.StorePath) == 0 {
		log.Fatalf("only one of --store-path='%v' , --in-memory='%v' must be set", va.StorePath, va.InMemory)
	}

	options := badger.DefaultOptions(va.StorePath).WithInMemory(va.InMemory)
	db, err := badger.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	// Add authentication middleware: only one user has access to the bot
	bot.Use(func(next telegram.HandlerFunc) telegram.HandlerFunc {
		return func(ctx telegram.Context) error {
			chat_id := fmt.Sprint(ctx.Chat().ID)
			authenticated := false

			err := db.Update(func(txn *badger.Txn) error {

				// fetch data
				fetched_id, err := txn.Get([]byte(user_key))

				// no user -> save this as the user
				if err == badger.ErrKeyNotFound {
					// insert it
					err = txn.Set([]byte(user_key), []byte(chat_id))
					if err != nil {
						return err
					}

					// log the user if no error
					authenticated = true
					return nil
				}

				// compare the chat with the saved user
				return fetched_id.Value(func(val []byte) error {
					// no value found: no user will ever be able to login
					if val == nil {
						log.Fatal(fmt.Errorf("no value found for key %v", user_key))
					}

					authenticated = string(val) == chat_id

					return nil
				})
			})

			if err != nil {
				log.Print(err)
				return nil
			}

			if authenticated {
				return next(ctx)
			}

			log.Printf("unauthorized loggin attempt (chat id: %v, username: %v, firstname: %v, lastname: %v)", chat_id, ctx.Chat().Username, ctx.Chat().FirstName, ctx.Chat().LastName)

			_, err = bot.Send(ctx.Sender(), "Sorry, this bot is private.")
			return err
		}
	})

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
