SHELL:=/usr/bin/env bash

build: clean copy_configs copy_web ./build/server
	@echo 'Done'

ebidlocal: cmd/example.go
	go build -o build/ebidlocal cmd/example.go

./build/server: cmd/server.go
	go build -o build/server cmd/server.go

runEbid: ebidlocal
	./build/ebidlocal | tee /tmp/index.html | md5 | tee /tmp/index.html.md5 

run: ./build/server
	./build/server

.PHONY: clean
clean:
	rm -rf ./build

copy_configs: configs
	mkdir -p ./build/configs
	cp -r ./configs/*.json ./build/configs/ || :

copy_web: web
	mkdir -p ./build/web
	cp -r ./web/ ./build/web/ || :

.PHONY: requestNewUser
requestNewUser:
	curl --request POST \
		-X POST \
		--header "Content-Type: application/json" \
	  	--data '{"username":"xyz","password":"xyz"}' \
		localhost:8282/user
