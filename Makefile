.PHONY: build clean

build:
	go build -o bin/lw ./cmd/lw

clean:
	rm -rf bin/
