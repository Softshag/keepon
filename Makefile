PREFIX ?= /usr/local
SRC = main.go runner.go runner-cli.go
BUILD_NAME=keepon


export GOPATH=$(CURDIR)/Godeps/_workspace


$(BUILD_NAME):
	@mkdir -p build
	go build -o build/$(BUILD_NAME) $(SRC)

install: $(BUILD_NAME)
	
	install -m 0755 build/$(BUILD_NAME) $(PREFIX)/bin/$(BUILD_NAME)

uninstall:
	rm -f $(PREFIX)/bin/$(BUILD_NAME)

clean:
	@rm -rf build