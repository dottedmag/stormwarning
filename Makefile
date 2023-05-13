APP=stormwarning
BIN=$(APP).linux-amd64
GB=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
H=tea

build:
	$(GB) -o $(BIN) .

deploy: build
	ssh $H mkdir -p $(APP)
	rsync $(BIN) $H:$(APP)/$(APP)
	rsync $(APP).service $(APP).timer $H:.config/systemd/user
	ssh $H systemctl --user daemon-reload
	ssh $H systemctl --user enable $(APP).service $(APP).timer
	ssh $H systemctl --user restart $(APP).service $(APP).timer

clean:
	rm -f $(BIN)
