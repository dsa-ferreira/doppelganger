build:
	go build
run:
	go run .
package:
	GOBIN=$(pwd)/bin go install
install:
	GOBIN=/usr/bin/ go install

# Python targets
install-py:
	pip install --user ./python

install-py-pipx:
	pipx install ./python

uninstall-py:
	pip uninstall -y doppelganger-py

uninstall-py-pipx:
	pipx uninstall doppelganger-py
