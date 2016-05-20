build:
	go build -o ry .

run: build
	./ry

clean:
	rm ry

.PHONY: build run clean
