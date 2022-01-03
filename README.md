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
## Process
Build for your local machine architecture
```
make build
```

Build for windows x86
```
make build-windows
```
note: for windows compilation from mac or linux, you will need to install mingw-64 from your favorite package manager

# License 
The source code files in this repository are license under the terms of the Mozilla Public License 2.0
