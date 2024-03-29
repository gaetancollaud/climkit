
PLATFORM=local
MIGRATIONDIR := pkg/postgres/migrations
MIGRATIONS :=  $(wildcard ${MIGRATIONDIR}/*.sql)
TOOLS := ${GOPATH}/bin/go-bindata

# This is all the tools required to compile, test and handle protobufs
#tools: ${TOOLS}

# Build all files.
build: ${MIGRATIONDIR}/bindata.go
	@echo "==> Building ./dist/sdm"
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/climkit-amd64 ./main.go
.PHONY: build

build-arm:
	@echo "==> Building ./dist/sdm"
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -o dist/climkit-arm ./main.go
.PHONY: build


${GOPATH}/bin/go-bindata:
	GO111MODULE=off go get -u github.com/go-bindata/go-bindata/...

${MIGRATIONDIR}/bindata.go: ${MIGRATIONS}
	# Building bindata
	go-bindata -o ${MIGRATIONDIR}/bindata.go -prefix ${MIGRATIONDIR} -pkg migrations ${MIGRATIONDIR}/*.sql

# Install from source.
install:
	@echo "==> Installing climkit ${GOPATH}/bin/climkit"
	go install ./...
.PHONY: install

# Run all tests.
test:
	go test -timeout 2m ./... && echo "\n==>\033[32m Ok\033[m\n"
.PHONY: test

.PHONY: docker
docker:
	@docker build . --target bin \
	--output bin/ \
	--platform ${PLATFORM}

# Clean.
clean:
	@rm -fr \
		dist \
.PHONY: clean
