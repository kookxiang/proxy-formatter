BINARY := proxy-formatter
SERVICE := proxy-formatter.service

PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
SYSTEMD_DIR := /usr/lib/systemd/system

build:
	go build -o $(BINARY) .

install: build
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 $(SERVICE) $(DESTDIR)$(SYSTEMD_DIR)/$(SERVICE)
	systemctl daemon-reload || true

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(SYSTEMD_DIR)/$(SERVICE)
	systemctl daemon-reload || true

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
