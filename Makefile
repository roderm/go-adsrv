proto:
	find . -type f -not \( -path "./vendor/*" -prune \) -name *.proto -exec \
		protoc \
			--proto_path=${GOPATH}/src:. \
			--go_out=plugins=grpc:${GOPATH}/src/ \
		{} \;