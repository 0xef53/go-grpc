CWD := $(shell pwd)

DOCKER_PB_ARGS := \
    -w /root/pkg \
    -v $(CWD):/root/pkg

protofiles_grpc = \
    field_options.proto

.PHONY: protobufs

all: protobufs

protobufs:
	docker run --rm -it $(DOCKER_PB_ARGS) 0xef53/go-proto-compiler:v3.18 \
		--proto_path options \
		--go_opt "plugins=grpc,paths=source_relative" \
		--go_out ./options \
		$(protofiles_grpc)
