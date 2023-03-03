B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)

build: 
	go build -o rpid -v -mod=vendor

info:
	- @echo "revision $(REV)"

service-deploy:
	make build
	sudo systemctl stop rpid.service
	sudo cp rpid /usr/bin/
	sudo cp rpid.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable rpid.service
	sudo systemctl start rpid.service

service-status:
	sudo systemctl status rpid.service

.PHONY: build info service-deploy service-status
