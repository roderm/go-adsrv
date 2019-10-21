GOPATH:=$(shell echo $(GOPATH))

.PHONY: proto build

proto: 
	find . -type f -name *.proto \
	-exec protoc --proto_path=${GOPATH}/src:. --go_out=plugins=grpc:${GOPATH}/src/ {} \;

build:
	# TODO: docker