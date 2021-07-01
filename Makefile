vendor: vendor/modules.txt

vendor/modules.txt: go.mod
	go mod vendor

.PHONY: build
build: vendor
	go build