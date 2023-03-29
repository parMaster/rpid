B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d)

build:
	go build -o dist/rpid -v

test:
	go test ./...

info:
	- @echo "revision $(REV)"

deploy:
	make build
	sudo systemctl stop rpid.service || true
	sudo cp dist/rpid /usr/bin/
	sudo cp dist/rpid.service /etc/systemd/system/
	sudo mkdir -p /etc/rpid
	sudo chown pi:pi /etc/rpid
	cp config/config.yml /etc/rpid/
	touch /etc/rpid/data.db
	chmod 0755 /etc/rpid/data.db
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

release:
	cp config/config_example.yml dist/config.yml
	cp LICENSE dist/LICENSE
	GOOS=linux GOARCH=arm64 go build -o dist/rpid
	cp -r dist rpid-$(GITREV)-arm64
	tar -czvf rpid-$(GITREV)-arm64.tar.gz rpid-$(GITREV)-arm64/*
	rm -rf rpid-$(GITREV)-arm64
	rm dist/rpid
	GOOS=linux GOARCH=arm go build -o dist/rpid
	cp -r dist rpid-$(GITREV)-arm
	tar -czvf rpid-$(GITREV)-arm.tar.gz rpid-$(GITREV)-arm/*
	rm -rf rpid-$(GITREV)-arm
	rm dist/rpid
	rm dist/config.yml

.PHONY: build info deploy status remove release test
