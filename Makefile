.PHONY: all
all: openapi peta; $(info $(M)...Build all of binary.) @ ## Build all of binary.

openapi: ;$(info $(M)...Begin to openapi.)  @ ## OpenAPI.
	go run ./tools/cmd/doc-gen/main.go

peta: ;$(info $(M)...Begin to build peta binary.) @ ## Build peta.
	hack/gobuild.sh ./