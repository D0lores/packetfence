include ../config.mk

.PHONY: pfhttpd-sha
pfhttpd-sha:
	make NAME="pfhttpd" _pfhttpd

.PHONY: pfhttpd
pfhttpd:
	make NAME="pfhttpd" _pfhttpd

.PHONY: _pfhttpd
_pfhttpd:
	cd caddy/caddy/caddy && \
	go build && \
	cd - && \
	mv caddy/caddy/caddy/caddy ${NAME}

.PHONY: pfhttpd-race
pfhttpd-race:
	cd caddy/caddy/caddy && \
	go build -race && \
	cd - && \
	mv caddy/caddy/caddy/caddy pfhttpd-race

.PHONY: pfdhcp
pfdhcp:
	cd dhcp && \
	go build && \
	cd - && \
	mv dhcp/dhcp pfdhcp

.PHONY: pfdns
pfdns:
	cd coredns && \
	go build && \
	cd - && \
	mv coredns/coredns pfdns

.PHONY: pfstats
pfstats:
	cd stats && \
        go build && \
        cd - && \
        mv stats/stats pfstats

.PHONY: pfdetect
pfdetect:
	cd detect && \
        go build && \
        cd - && \
        mv detect/detect pfdetect

.PHONY: galera-autofix
galera-autofix:
	cd galeraautofix && \
        go build && \
        cd - && \
        mv galeraautofix/galeraautofix galera-autofix

.PHONY: pfacct
pfacct:
	cd acct && \
        go build && \
        cd - && \
        mv acct/acct pfacct



.PHONY: clean-caddy-src
clean-caddy-src:
	find caddy/ -type f -exec sed -i.bak "s'github.com/mholt/caddy'github.com/inverse-inc/packetfence/go/caddy/caddy'g" {} \; ; find . -name '*.bak' -delete
	find caddy/ -type f -exec sed -i.bak "s'github.com/caddyserver/caddy'github.com/inverse-inc/packetfence/go/caddy/caddy'g" {} \; ; find . -name '*.bak' -delete
	find caddy/caddy/ -name '*_test.go' -delete

.PHONY: go-env
go-env:
	GOVERSION=$(GOVERSION) ../addons/dev-helpers/setup-go-env.sh

.PHONY: test
test:
	/usr/local/pf/t/pfconfig-test ;\
	PFCONFIG_TESTING=y go test -count=1 ./...

.PHONY: all
all: $(GOBINARIES)

.PHONY: copy
copy:
	cp -f $(GOBINARIES) $(DESTDIR)$(SBINDIR)

.PHONY: clean
clean:
	rm -f $(GOBINARIES)

.PHONY: clean-coredns-src
clean-coredns-src:
	find coredns/ -type f -exec sed -i.bak "s'github.com/mholt/caddy'github.com/inverse-inc/packetfence/go/caddy/caddy'g" {} \; ; find . -name '*.bak' -delete
	find coredns/ -type f -exec sed -i.bak "s'github.com/coredns/coredns'github.com/inverse-inc/packetfence/go/coredns'g" {} \; ; find . -name '*.bak' -delete
	find coredns/ -name '*_test.go' -delete
