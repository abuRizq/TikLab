.PHONY: build test clean

build:
	go build -o sandbox .

test:
	go test -v ./...

clean:
	rm -f sandbox sandbox.exe
