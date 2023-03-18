deploy:
	- sudo systemctl stop rpid.service || true
	sudo cp rpid /usr/bin/
	sudo cp rpid.service /etc/systemd/system/
	sudo mkdir -p /etc/rpid
	sudo cp config.yml /etc/rpid/
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

.PHONY: deploy status remove