
all:
	go build .

linux:
	CGO_ENABLED=0 GOOS=linux go build .

image: linux
	docker build . -t "kubewait"

.PHONY: all
