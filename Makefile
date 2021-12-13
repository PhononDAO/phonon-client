ifeq ($(OS),Windows_NT)
	Win-CC = $(CC)
else
	Win-CC = x86_64-w64-mingw32-gcc
endif

run: generate
	go run main/phonon.go

build: generate
	go build main/phonon.go

windows-build: generate
	GOOS=windows CGO_ENABLED=1 CC=$(Win-CC) go build main/phonon.go

test:
	go test -v -count=1 ./...

short-test:
	go test -v -count=1 ./... -short

generate:
	go generate ./...

android-sdk:
	mkdir -p androidSDK/
	cd session && gomobile bind  -target android -o ../androidSDK/phononAndroid.aar

