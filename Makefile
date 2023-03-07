B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)

build: 
	go build -o dist/rpid -v -mod=vendor

info:
	- @echo "revision $(REV)"

deploy:
	make build
	sudo systemctl stop rpid.service
	sudo cp dist/rpid /usr/bin/
	sudo cp dist/rpid.service /etc/systemd/system/
	sudo mkdir -p /etc/rpid
	sudo cp config/config.yml /etc/rpid/
	sudo systemctl daemon-reload
	sudo systemctl enable rpid.service
	sudo systemctl start rpid.service

status:
	sudo systemctl status rpid.service

remove:
	sudo systemctl stop rpid.service
	sudo rm /usr/bin/rpid
	sudo rm /etc/rpid/config.yml
	sudo rm /etc/rpid -rf
	sudo rm /etc/systemd/system/rpid.service

.PHONY: build info deploy status remove
