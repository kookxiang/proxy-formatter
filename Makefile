PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
WORKDIR := /var/lib/proxy-formatter
SYSTEMD_DIR := /usr/lib/systemd/system

build: bin geosite

bin:
	go build -trimpath -o proxy-formatter -v .

geosite:
	curl -L https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat -o geosite.dat

install:
	install -Dm755 proxy-formatter $(DESTDIR)$(BINDIR)/proxy-formatter
	install -Dm644 proxy-formatter.service $(DESTDIR)$(SYSTEMD_DIR)/proxy-formatter.service
	install -Dm644 geosite.dat $(DESTDIR)$(WORKDIR)/geosite.dat
	systemctl daemon-reload || true

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/proxy-formatter
	rm -f $(DESTDIR)$(SYSTEMD_DIR)/proxy-formatter.service
	rm -rf $(DESTDIR)$(WORKDIR)/geosite.dat
	systemctl daemon-reload || true

clean:
	rm -f proxy-formatter

.PHONY: build bin geosite install uninstall clean
