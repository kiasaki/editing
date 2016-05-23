build:
	go build -o ry-lisp-repl ./cmd/ry-lisp-repl
	go build -o ry .

run: build
	./ry

clean:
	rm ry

.PHONY: build run clean
