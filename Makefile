ifeq ($(OS),Windows_NT)
	Win-CC = $(CC)
else
	Win-CC = x86_64-w64-mingw32-gcc
endif
run:
	go run main/phonon.go

build:
	go build main/phonon.go

windows-build:
	GOOS=windows CGO_ENABLED=1 CC=$(Win-CC) go build main/phonon.go

test:
	go test -v -count=1 ./...

short-test:
	go test -v -count=1 ./... -short
