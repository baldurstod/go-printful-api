BINARY_NAME=go_printful_api

build:
	go build  -o dist/go_printful_api.exe ./src/main.go

build-test:
	go build -ldflags="-X go_printful_api/src/server.ReleaseMode=false" -o build/${BINARY_NAME}.exe ./src/main.go

run-test: build-test
	build/${BINARY_NAME}.exe

clean:
	go clean
