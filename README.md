[![Phonon Logo](./assets/logo.svg)](https://phonon.network)
![Mozilla Public 2.0 License](https://img.shields.io/badge/license-MozillaPublic2.0-green)
![Go Programming Language](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)

# Come see us

[![Reddit](https://img.shields.io/badge/Reddit-FF4500?style=flat&logo=reddit&logoColor=white)](https://reddit.com/r/PhononDAO)
[![Twitter](https://img.shields.io/badge/Twitter-1DA1F2?style=flat&logo=twitter&logoColor=white)](https://twitter.com/PhononDAO)

[![Discord](https://img.shields.io/badge/Discord-7289DA?style=flat&logo=discord&logoColor=white)](https://discord.gg/RNQtyBaKMH)[![Discord](https://img.shields.io/discord/921799167779672064?labelColor=5b209a)](https://discord.gg/RNQtyBaKMH)

# Phonon-client Alpha

phonon-client is a library, user interface, and set of utilities for interacting with phonon cards and the phonon network as a whole.
This library can be cloned and used directly, or prebuilt binaries are available for the latest master under the releases page.

# Usage

This client can be used in multiple ways depending on your goals.

For general interactive use the repl (Read Eval Print Loop) is the most user friendly and fleshed out interface. It can be opened like so

```
go run main/phonon.go repl
help
```

Calling help will provide the complete list of available commands

For quickly testing changes or performing simple one off operations the cmd interface is preferable. Available commands can be brought up by running the below.

```
go run main/phonon.go
```

Commands are then called like so:

```
go run main/phonon.go init
```

For users interested in integrating phonon library code into applications, building new user interfaces, or generally interfacing with phonon cards programmatically, the primary interface to be concerned with is PhononCard in model/card.go. This describes the full set of commands which the phonon javacard applet is capable of processing. The actual card implementation is under card/phononCommandSet.go. There is a mostly complete mock implementation of the javacard applet which can be used in testing under card/mockCard.go

## Initialization

A new phonon card must have a certificate installed and a pin initialized in order to be able to perform most functions. A new development phonon card can be set up after the applet is installed by running the following two commands. These are available from within the repl as well if preferred.

```
go run main/phonon.go installCardCert -d
go run main/phonon.go init
```

# Building

## Requirements:

- a go compiler
- a c compiler for whatever architecture you are compiling for
- the go stringer tool

## Dependencies

A recent version of golang, up to date requirement given in the go.mod file. (1.19 as of this writing.)

To build the frontend submodule, you'll need to have nodejs and npm installed prior to running the `make build` command.

The go stringer tool, which generates human readable strings based on constant variable names for display and data exchange purposes. Install with:

```
go install golang.org/x/tools/cmd/stringer@latest
```

Ensure that your go/bin folder is set in your system's $PATH variable

```
export PATH=$PATH:`go env GOPATH`/bin
```

It's recommended to set this variable in your shell profile to make this value persist in new sessions.

## Process

Build for your local machine architecture

```
make build
```

Build for windows x86

```
make build-windows
```

NOTE: For windows compilation from Mac or Linux, you will need to install `mingw-64` from your favorite package manager.

## Additional Notes

The default pin for mock cards is `111111`.

### Running the webUI

To run the webUI you'll need to checkout the submodules in this repository.

You'll want to run the following:

```
make checkout-submodules
```

Here's the order of commands for running the webUI for the first time:

1. `make checkout-submodules`
2. `make build` or `make build-windows`
3. `go run main/phonon.go webUI`

NOTE: You don't need to checkout the submodules every time you run the webUI.

# License

The source code files in this repository are license under the terms of the Mozilla Public License 2.0
