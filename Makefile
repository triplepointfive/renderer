OUT := bin/renderer

build:
	go build -i -o $(OUT)

run: build
	$(OUT)
