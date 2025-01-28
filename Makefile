BINARY_NAME=go_printful_api

build:
	go build  -o dist/go_printful_api.exe ./src/main.go

run: build
	dist/${BINARY_NAME}.exe

clean:
	go clean
