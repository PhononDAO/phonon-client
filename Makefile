ifeq ($(OS),Windows_NT)
	Win-CC = $(CC)
else
	Win-CC = x86_64-w64-mingw32-gcc
endif

build: generate frontend
	go build main/phonon.go

client-build: generate #build just the golang code without the frontend
	go build main/phonon.go

windows-build: generate frontend
	GOOS=windows CGO_ENABLED=1 go build -ldflags "-H=windowsgui" main/phonon.go

test:
	go test -v -count=1 ./...

short-test:
	go test -v -count=1 ./... -short

generate:
	go generate ./...

android-sdk:
	mkdir -p androidSDK/
	cd session && gomobile bind  -target android -o ../androidSDK/phononAndroid.aar

frontend:
	(cd gui/frontend && npm install)
	(cd gui/frontend && npm run build)

release-mac: build
	cp phonon ./release/MacOS/Phonon.app/Contents/MacOS/phonon

checkout-submodules:
	git submodule update --init
