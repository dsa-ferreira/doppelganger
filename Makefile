build:
	go build
run:
	go run .
package:
	GOBIN=$(pwd)/bin go install
install:
	GOBIN=/usr/bin/ go install
