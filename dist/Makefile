deploy:
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

.PHONY: deploy status remove