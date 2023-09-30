
BUILD_TIME_STAMP := $(shell date +%FT%TZ)

TOOLS	:= \
      github.com/golang/dep/cmd/dep; \
		github.com/go-swagger/go-swagger/cmd/swagger; \
		golang.org/x/lint/golint; \
		github.com/mitchellh/gox; 

GETTOOLS := $(foreach TOOL,$(TOOLS), go get -v $(TOOL))

PACKAGES := $(shell go list ./...)

default: build test
all: gen build test

# Build will always run lint and vet
#build: lint vet
build: 
	GOOS=darwin GOARCH=arm64 go \
		build -o cricsum \
		./cmd/main.go

#dist: lint vet
dist: 
	mkdir -p dist
	cd cmd
	# run gox to build platform dependent binaries
	gox \
	  -os="linux darwin windows"  -arch="arm64 amd64 386" -verbose \
	  -output="./dist/{{.Dir}}_{{.OS}}_{{.Arch}}" ./cmd/

push:
ifndef HOSTS
	$(error "HOSTS is not defined. Please define HOSTS with a space separated list of hosts to deploy to.")
endif
	for host in $(HOSTS) ; do scp -p dist/cmd_linux_amd64 @$$host:/x/cricsum; done


tools:
	$(GETTOOLS)

all: tools build test

clean:
	rm -rf main \
      coverage.out \
      coverage-all.out 

deps:
	dep ensure -v

lint:
	golint -set_exit_status $(shell go list ./... | grep -v assets)

vet:
	go vet $(shell go list ./...)

# This is a workaround for go test -cover with multiple packages
test:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		test -f coverage.out && tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

cover:
	go tool cover -func=coverage-all.out

init:
	go mod init
	go mod tidy


.PHONY: \
	build \
	push \
	restart \
	tools \
	all \
	clean \
	gen \
	deps \
	lint \
	vet \
	test
