# Chimp

Chimp is a simple input package (mice, tablets etc). Device support will be
added when there is a need for it. The initial purpose of the package is to
support drawing tablet input for a paint program yet to be written.

The package builds on top of (a forked) version of
[golang-evdev](https://github.com/gvalkov/golang-evdev) trying to provide a more
user-friendly API. For example Wacom tablets are exposed as three independent
devices by Linux and golang-evdev. This packages opens all three automatically
and multiplexes all events from them into one event stream.

## Documentation

Use the [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc) tool to view API
documentation.

## Supported OS

Only Linux is supported at the moment but the Linux device events are converted
to another representation exported by this package in a platform independent
manner.

## Sample Programs

There are two sample programs under `cmd/` which is built in the standard Go
fashion.

Pick your poision:

* `go get github.com/johan-bolmsjo/chimp/cmd/chimp-dump-events`
* `cd cmd/chimp-dump-events && go build`
* `go install ./...`

### chimp-dump-events

Open first supported device and dump all events read from it.

### chimp-list-devices

List all found supported devices.

## Supported devices

* Wacom Bamboo 16FG 6x8 (Linux)
