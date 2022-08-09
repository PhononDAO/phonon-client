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

release-mac: generate frontend
	CGO_ENABLED=1 CC="clang -target arm64v8-apple-darwin-macho" GOOS=darwin GOARCH=arm64 go build -o phonon_arm64 main/phonon.go
	CGO_ENABLED=1 CC="clang -target x86_64-apple-darwin-macho" GOOS=darwin GOARCH=amd64 go build -o phonon_x86_64 main/phonon.go
	cp phonon_arm64 ./release/MacOS/Phonon.app/Contents/MacOS/phonon_arm64
	cp phonon_x86_64 ./release/MacOS/Phonon.app/Contents/MacOS/phonon_x86_64
	create-dmg \
		--app-drop-link 100 300 \
		--icon "Phonon.app" 100 100\
		--volname "Phonon Installer" \
		--hide-extension "Phonon.app" \
		--window-size 1200 600 \
		--background "./release/MacOS/background.png" \
		phonon.dmg \
		./release/MacOS/Phonon.app

checkout-submodules:
	git submodule update --init
