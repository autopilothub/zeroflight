.PHONY: test build build-pi install-service

BINARY=zeroflight
PKG=./cmd/zeroflight

test:
	go test ./...

build:
	go build -o $(BINARY) $(PKG)

build-pi:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY) $(PKG)

install-service:
	sudo cp deploy/zeroflight.service /etc/systemd/system/
	sudo systemctl daemon-reload

lint:
	go vet ./...
