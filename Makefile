SHELL=/bin/bash

help:
	@echo Help for Oddmu
	@echo =====================
	@echo
	@echo make run
	@echo "    runs program, offline"
	@echo
	@echo make test
	@echo "    runs the tests"
	@echo
	@echo make upload
	@echo "    this is how I upgrade my server"
	@echo
	@echo go build
	@echo "    just build it"

run:
	go run .

test:
	go test

upload:
	go build
	rsync --itemize-changes --archive oddmu oddmu.service *.html sibirocobombus.root:/home/oddmu/
	ssh sibirocobombus.root "systemctl restart oddmu"
