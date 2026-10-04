default: build

build:
	go build -o terraform-provider-volcano .

install: build
	go install .

lint:
	go vet ./...
	gofmt -l .

test:
	go test ./... -v $(TESTARGS) -timeout=120s

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout=120m

docs:
	go tool tfplugindocs generate

fmt:
	gofmt -w .

.PHONY: build install lint test testacc docs fmt
