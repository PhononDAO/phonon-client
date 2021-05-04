run:
	go run main/phonon.go

build:
	go build main/phonon.go

windows-build:
	GOOS=windows go build main/phonon.go

test:
	go test -v -count=1 ./...