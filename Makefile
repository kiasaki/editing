build:
	go build -o ry .

run: build
	./ry

deps:
	go get github.com/gdamore/tcell
	go get github.com/mattn/go-runewidth
	go get github.com/go-errors/errors
	go get github.com/kiasaki/go-rope

clean:
	rm ry

r:
	go build -v -o ry r/r.go && ./ry

.PHONY: build run deps clean r
