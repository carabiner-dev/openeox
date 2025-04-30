.PHONY: gen
gen: ## Genera las librerias desde los protos
	which protoc-go-inject-tag || go install github.com/favadi/protoc-go-inject-tag@latest
	which protoc-gen-go || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	buf generate
	protoc-go-inject-tag -input=./types/tbd/openeox.pb.go
