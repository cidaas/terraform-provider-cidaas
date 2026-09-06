TEST?=./...
HOSTNAME=registry.terraform.io
NAMESPACE=Cidaas
NAME=cidaas
BINARY=terraform-provider-${NAME}
VERSION=4.0.0-alpha.1
OS_ARCH?=$$(go env GOOS)_$$(go env GOARCH)

default: build

build:
	go build -ldflags="-X main.version=$(VERSION)" -o ${BINARY} .

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/

test:
	go test $(TEST) $(TESTARGS) -timeout=5m -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

test-ci:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m -parallel=4 -coverprofile .coverage.txt ./...
	go tool cover -func .coverage.txt
	go tool cover -html=.coverage.txt -o coverage.html

fmt:
	gofmt -s -w ./internal ./main.go

generate:
	go generate ./...

.PHONY: default build install test testacc test-ci fmt generate
