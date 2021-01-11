run:
	go run main/main.go

build:
	go build main/phonon.go

windows-build:
	GOOS=windows go build main/phonon.go