SHELL=/bin/bash

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
	@echo go build
	@echo "    just build it"
	@echo
	@echo make install
	@echo "    install the files to ~/.local"
	@echo
	@echo make upload
	@echo "    this is how I upgrade my server"

run:
	go run .

test:
	go test -shuffle on .

upload:
	go build
	rsync --itemize-changes --archive oddmu sibirocobombus.root:/home/oddmu/
	ssh sibirocobombus.root "systemctl restart oddmu; systemctl restart alex; systemctl restart claudia; systemctl restart campaignwiki"
	@echo Changes to the template files need careful consideration

docs:
	cd man; make

install:
	make docs
	for n in 1 5 7; do install -D -t $$HOME/.local/share/man/man$$n man/*.$$n; done
	go build
	install -D -t $$HOME/.local/bin oddmu

missing:
	for f in man/*.txt; do grep --quiet "$$f" README.md || echo $$f is not in the README; done
