package main

import (
	"fmt"
	"os"

	remote_http "github.com/maxime915/waker/cmd/remote-http"
	send_wol "github.com/maxime915/waker/cmd/send-wol"
	serve_http "github.com/maxime915/waker/cmd/serve-http"

	serve_telegram "github.com/maxime915/waker/cmd/serve-telegram"
	"github.com/voxelbrain/goptions"
)

const defaultBroadcast = "192.168.0.255:9"

func main() {
	options := struct {
		Version bool          `goptions:"-v, --version, description='Print version'"`
		Help    goptions.Help `goptions:"-h, --help, description='Show this help'"`

		goptions.Verbs

		ServeHTTP     serve_http.VerbArguments     `goptions:"serve-http"`
		SendWOL       send_wol.VerbArguments       `goptions:"send-wol"`
		GetHTTP       remote_http.VerbArguments    `goptions:"remote-http"`
		ServeTelegram serve_telegram.VerbArguments `goptions:"serve-telegram"`
	}{
		// Default values
		ServeHTTP: serve_http.VerbArguments{
			BindingAddress: "0:0",
			Broadcast:      defaultBroadcast,
		},
		SendWOL: send_wol.VerbArguments{
			Broadcast: defaultBroadcast,
		},
		ServeTelegram: serve_telegram.VerbArguments{
			Broadcast: defaultBroadcast,
		},
	}
	goptions.ParseAndFail(&options)

	if options.Version {
		fmt.Println("waker version 0.3")
		return
	}

	switch options.Verbs {
	case "serve-http":
		options.ServeHTTP.Execute()
	case "send-wol":
		options.SendWOL.Execute()
	case "remote-http":
		options.GetHTTP.Execute()
	case "serve-telegram":
		options.ServeTelegram.Execute()
	default:
		fmt.Fprintln(os.Stderr, "A verb is required.")
		goptions.PrintHelp()
		os.Exit(1)
	}
}
