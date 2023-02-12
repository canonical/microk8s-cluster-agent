all: cluster-agent

.PHONY = go.fmt go.vet go.lint go.staticcheck go.test

cluster-agent: $(shell find . -name '*.go') go.mod go.sum
	CGO_ENABLED=0 go build -ldflags '-s -w' -o cluster-agent ./main.go

go.fmt:
	go mod tidy
	cd tools && go mod tidy
	go fmt ./...

go.vet:
	go vet ./...

go.lint:
	cd tools && go run golang.org/x/lint/golint -set_exit_status ../...

go.staticcheck:
	cd tools && go install honnef.co/go/tools/cmd/staticcheck
	$(shell go env GOPATH)/bin/staticcheck ./...

go.test:
	go test -v ./...
