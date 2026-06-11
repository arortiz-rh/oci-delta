.PHONY: build clean test test-coverage fmt

COVERDIR ?= $(CURDIR)/.coverdata

VERSION := 0.1.0

build:
	go build

clean:
	rm -f oci-delta
	rm -rf $(COVERDIR)
	rm -rf $(CURDIR)/rpmbuild
	rm -f build/package/rpm/*-vendor.tar.bz2
	rm -f build/package/rpm/oci-delta-*.tar.gz

test: build
	go test ./...
	python3 tests/test-synthetic.py
	tests/integration-test.sh

test-coverage:
	go build -cover -o oci-delta
	rm -rf $(COVERDIR) && mkdir -p $(COVERDIR)/unit $(COVERDIR)/integration
	go test -cover ./... -args -test.gocoverdir=$(COVERDIR)/unit
	GOCOVERDIR=$(COVERDIR)/integration python3 tests/test-synthetic.py
	GOCOVERDIR=$(COVERDIR)/integration tests/integration-test.sh
	go tool covdata percent -i $(COVERDIR)/unit,$(COVERDIR)/integration

fmt:
	go fmt ./...

install:
	go install

#
# RPM packaging
#

RPM_PKGDIR = build/package/rpm
RPM_SPECFILE = $(RPM_PKGDIR)/oci-delta.spec
RPM_TOML = $(RPM_PKGDIR)/go-vendor-tools.toml
RPM_TOPDIR = $(CURDIR)/rpmbuild

.PHONY: vendor-tarball
vendor-tarball:
	git archive --prefix=oci-delta-$(VERSION)/ HEAD | gzip > $(RPM_PKGDIR)/oci-delta-$(VERSION).tar.gz
	go mod vendor
	tar cjf $(RPM_PKGDIR)/oci-delta-$(VERSION)-vendor.tar.bz2 vendor/
	rm -rf vendor/

.PHONY: srpm
srpm: vendor-tarball
	mkdir -p $(RPM_TOPDIR)/SOURCES
	cp $(RPM_PKGDIR)/oci-delta-*.tar.gz $(RPM_PKGDIR)/*-vendor.tar.bz2 $(RPM_TOML) $(RPM_TOPDIR)/SOURCES/
	rpmbuild -bs \
		--define "_topdir $(RPM_TOPDIR)" \
		--with tests \
		$(RPM_SPECFILE)

.PHONY: rpm
rpm: vendor-tarball
	mkdir -p $(RPM_TOPDIR)/SOURCES
	cp $(RPM_PKGDIR)/oci-delta-*.tar.gz $(RPM_PKGDIR)/*-vendor.tar.bz2 $(RPM_TOML) $(RPM_TOPDIR)/SOURCES/
	rpmbuild -bb \
		--define "_topdir $(RPM_TOPDIR)" \
		--with tests \
		$(RPM_SPECFILE)

.PHONY: scratch
scratch: vendor-tarball
	mkdir -p $(RPM_TOPDIR)/SOURCES
	cp $(RPM_PKGDIR)/oci-delta-*.tar.gz $(RPM_PKGDIR)/*-vendor.tar.bz2 $(RPM_TOML) $(RPM_TOPDIR)/SOURCES/
	rpmbuild -bb \
		--define "_topdir $(RPM_TOPDIR)" \
		--without tests \
		--nocheck \
		$(RPM_SPECFILE)
