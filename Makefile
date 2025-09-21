SHELL=/bin/bash
PREFIX=${HOME}/.local

.PHONY: help build test run upload docs install priv

help:
	@echo Help for Oddmu
	@echo ==============
	@echo make run
	@echo "    runs program, offline"
	@echo make test
	@echo "    runs the tests without log output"
	@echo make docs
	@echo "    create man pages from text files"
	@echo make build
	@echo "    just build it"
	@echo make install
	@echo "    install the files to ~/.local"
	@echo sudo make install PREFIX=/usr/local
	@echo "    install the files to /usr/local"
	@echo make upload
	@echo "    this is how I upgrade my server"
	@echo make dist
	@echo "    cross compile for other systems"
	@echo make clean
	@echo "    remove built files"

build: oddmu

oddmu: *.go
	go build

test:
	rm -rf testdata/*
	go test -shuffle on .

run:
	go run .

upload: build
	rsync --itemize-changes --archive oddmu sibirocobombus.root:/home/oddmu/
	ssh sibirocobombus.root "systemctl restart oddmu; systemctl restart alex; systemctl restart claudia; systemctl restart campaignwiki; systemctl restart community"
	@echo Changes to the template files need careful consideration

docs:
	cd man; make man

install:
	for n in 1 5 7; do install -D -t ${PREFIX}/share/man/man$$n man/*.$$n; done
	install -D -t ${PREFIX}/bin oddmu

clean:
	rm --force oddmu oddmu.exe oddmu-{linux,darwin,windows}-{amd64,arm64}{,.tar.gz}
	cd man && make clean

dist: oddmu-linux-amd64.tar.gz oddmu-linux-arm64.tar.gz oddmu-darwin-amd64.tar.gz oddmu-windows-amd64.tar.gz

oddmu-linux-amd64: *.go
	GOOS=linux GOARCH=amd64 go build -o $@

oddmu-linux-arm64: *.go
	env GOOS=linux GOARCH=arm64 GOARM=5 go build -o $@

oddmu-darwin-amd64: *.go
	GOOS=darwin GOARCH=arm64 go build -o $@

oddmu.exe: *.go
	GOOS=windows GOARCH=amd64 go build -o $@

oddmu-windows-amd64.tar.gz: oddmu.exe
	cd man && make html
	tar --create --file $@ --transform='s/^/oddmu\//' --exclude='*~' \
	  $< *.md man/*.[157].{html,md} themes/

%.tar.gz: %
	tar --create --gzip --file $@ --transform='s/^$</oddmu/' --transform='s/^/oddmu\//' --exclude='*~' \
	  $< *.html Makefile *.socket *.service *.md man/Makefile man/*.[157] themes/

priv:
	sudo setcap 'cap_net_bind_service=+ep' oddmu
