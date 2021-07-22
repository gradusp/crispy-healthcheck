export GOSUMDB=off
export GO111MODULE=on

export GOBIN=$(CURDIR)/bin
GOLANGCI_BIN:=$(GOBIN)/golangci-lint
GOLANGCI_REPO=https://github.com/golangci/golangci-lint
GOLANGCI_LATEST_VERSION:= $(shell git ls-remote --tags --refs --sort='v:refname' $(GOLANGCI_REPO)|tail -1|egrep -E -o "v\d+\.\d+\..*")

ifneq ($(wildcard $(GOLANGCI_BIN)),)
	GOLANGCI_CUR_VERSION:=v$(shell $(GOLANGCI_BIN) --version|sed -E 's/.* version (.*) built from .* on .*/\1/g')
else
	GOLANGCI_CUR_VERSION:=
endif

# install linter tool
.PHONY: install-linter
install-linter:
ifneq ($(GOLANGCI_CUR_VERSION), $(GOLANGCI_LATEST_VERSION))
	$(info Installing GOLANGCI-LINT $(GOLANGCI_LATEST_VERSION)...)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LATEST_VERSION)
else
	@echo "GOLANGCI-LINT is need not install"
endif

# run full lint like in pipeline
.PHONY: lint
lint: install-linter
	$(GOLANGCI_BIN) cache clean && \
	$(GOLANGCI_BIN) run --config=.golangci.yaml -v ./...

# install project dependencies
.PHONY: go-deps
go-deps:
	$(info Install dependencies...)
	@go mod tidy && go mod vendor && go mod verify

.PHONY: bin-tools
bin-tools:
	@echo "Install bin tools"
ifeq ($(wildcard $(GOBIN)/protoc-gen-grpc-gateway),)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
endif
ifeq ($(wildcard $(GOBIN)/protoc-gen-openapiv2),)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
endif
ifeq ($(wildcard $(GOBIN)/protoc-gen-go),)
	go install google.golang.org/protobuf/cmd/protoc-gen-go
endif
ifeq ($(wildcard $(GOBIN)/protoc-gen-go-grpc),)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
endif
	@echo "" > /dev/null



.PHONY: generate
generate: bin-tools
	@echo "Generate API from proto"
	@PATH=$(PATH):$(GOBIN) && \
	protoc -I $(CURDIR)/vendor/github.com/grpc-ecosystem/grpc-gateway/v2/ \
		-I $(CURDIR)/3d-party \
		--go_out $(CURDIR)/pkg \
		--go-grpc_out $(CURDIR)/pkg \
		--proto_path=$(CURDIR)/api \
		--grpc-gateway_out $(CURDIR)/pkg \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt standalone=false \
		healthcheck/healthchecker.proto && \
	protoc -I $(CURDIR)/vendor/github.com/grpc-ecosystem/grpc-gateway/v2/ \
		-I $(CURDIR)/3d-party \
		--proto_path=$(CURDIR)/api \
		--openapiv2_out $(CURDIR)/internal \
		--openapiv2_opt logtostderr=true \
		healthcheck/healthchecker.proto

.PHONY: test
test:
	$(info Running tests...)
	@go test -count=1 ./...



