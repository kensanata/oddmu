SHELL=/bin/bash
PREFIX=${HOME}/.local

.PHONY: help build test run upload docs install missing

help:
	@echo Help for Oddmu
	@echo =====================
	@echo
	@echo make run
	@echo "    runs program, offline"
	@echo
	@echo make test
	@echo "    runs the tests without log output"
	@echo
	@echo make docs
	@echo "    create man pages from text files"
	@echo
	@echo make build
	@echo "    just build it"
	@echo
	@echo make install
	@echo "    install the files to ~/.local"
	@echo
	@echo make upload
	@echo "    this is how I upgrade my server"

build: oddmu

oddmu: *.go
	go build

test:
	go test -shuffle on .

run:
	go run .

upload: build
	rsync --itemize-changes --archive oddmu sibirocobombus.root:/home/oddmu/
	ssh sibirocobombus.root "systemctl restart oddmu; systemctl restart alex; systemctl restart claudia; systemctl restart campaignwiki; systemctl restart community"
	@echo Changes to the template files need careful consideration

docs:
	cd man; make

install:
	for n in 1 5 7; do install -D -t ${PREFIX}/share/man/man$$n man/*.$$n; done
	install -D -t ${PREFIX}/.local/bin oddmu

# More could be added, of course!
dist: oddmu-linux-amd64.tar.gz

oddmu-linux-amd64: *.go
	GOOS=linux GOARCH=amd64 go build -o $@

%.tar.gz: %
	tar czf $@ --transform='s/^$</oddmu/' --transform='s/^/oddmu\//' --exclude='*~' \
	  $< Makefile *.socket *.service *.md man/Makefile man/*.1 man/*.5 man/*.7 themes/
