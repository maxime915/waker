# waker

HTTP interface to send WakeOnLan magic packet remotely

## Usage

Launch with `waker [address:port [configfile]]`

- If `address` is an IPv6 address it has to be surounded by brackets, e.g. `waker "[0::0]:9009"`
- `configfile` is a file containing the bytes of the targetted MAC address, `wake_test.go` can encode a MAC address to a config file if needed

Access with GET `address:port/wake`, the success message is "200 - Magic packet send", if any error that couldn't be detected at startup occurs, "500 - Error while sending magic packet: ..." is returned.

Example:

```sh
$ ./waker "[0::0]:9009" &
[1] 56980
$ curl "[0::0]:9009/wake"
200 - Magic packet send
$ kill 56980
[1]  + 56980 terminated  ./waker "[0::0]:9009"
```

## Building

There are two target: "waker" and "piwaker". The first one is the default for the architecture building the package and the second one is targetted toward RaspberrPI products. `make all` builds both version.

made with go1.15.7