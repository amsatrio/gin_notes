
.SILENT: build
.SILENT: run
.SILENT: swag_init
.SILENT: dev
.SILENT: air

BINARY_NAME=gin_notes

build:
	mkdir -p bin
	#GOARCH=amd64 GOOS=darwin go build -o out/bin/${BINARY_NAME}-darwin main.go
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux main.go
	#GOARCH=amd64 GOOS=windows go build -o out/bin/${BINARY_NAME}-windows main.go

run: build
	./bin/${BINARY_NAME}-linux

clean:
	go clean
	rm -rf ${BINARY_NAME}
	rm -rf bin/${BINARY_NAME}-darwin
	rm -rf bin/${BINARY_NAME}-linux
	rm -rf bin/${BINARY_NAME}-windows
	rm -rf output/go
	rm -rf output/java
	rm -rf output/csv

swag_init:
	/home/mos/go/bin/swag init
	sed -i 's/LeftDelim:        "{{",//g' docs/docs.go
	sed -i 's/RightDelim:       "}}",//g' docs/docs.go
dev: swag_init
	/home/mos/go/bin/CompileDaemon -command="./${BINARY_NAME}" -pattern="(.+\.go|.+\.c|.+\.html|.+\.css|.+\.js)"

# hot reload by air
air:
	swag_init
	/home/mos/go/bin/air server --port 8080

migrate:
	go run migrate/migrate.go

test_dev:
	go test ./tests/... -v > ./docs/test.out

test_coverage:
	go test ./tests/... -v -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet

lint:
	golangci-lint run --enable-all

# web
web_run:
	 /home/mos/go/bin/fyne serve

