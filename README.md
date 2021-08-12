# waker

HTTP interface to send WakeOnLan magic packet remotely

## Installation

`go install github.com/maxime915/waker@latest`

## Usage

Launch with `waker -target *target* [-address *address*] [-verbose [*verbose*]]`

- address is a string representing the local address and port to bind to (default "127.0.0.1:0")
- target is a string representing the MAC address of the target to wake
- verbose is a flag used to output a confirmation message before launching the server (default false)

Access with GET `address:port/wake`, the success message is "200 - Magic packet send", if any error that couldn't be detected at startup occurs, "500 - Error while sending magic packet: ..." is returned.

Example:

```sh
$ ./waker -target 00:00:00:00:00:00 -address "[0::0]:0" -verbose &
[1] 21419
waker is listening on [::]:63707 for target 00:00:00:00:00:00
$ curl "[0::0]:63707/wake"
200 - Magic packet send
$ kill 21419
[1]  + 21419 terminated  ./waker -target 00:00:00:00:00:00 -address "[0::0]:0" -verbose
```

`waker` sends the UDP datagram to launch the WakeOnLan procedure on the interface 00:00:00:00:00:00. The datagram is sent on the broadcast address of the local subnet (192.168.0.255:9).

## TODO

Make the UDP broadcast address & port be user configurable.

## Building

There are two target: "waker" and "piwaker". The first one is the default for the locall architecture and the second one is targeted towards Raspberry Pi. `make all` builds both version.

made with go1.16