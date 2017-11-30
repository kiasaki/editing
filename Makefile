build:
	go build -i -v -o ry .

run: build
	./ry

deps:
	go get github.com/tools/godep
	godep restore

install:
	install ./ry /usr/local/bin/ry

clean:
	rm ry

.PHONY: build run deps clean
