.PHONY: buf-mod-update
buf-mod-update:
	@test -s ./proto/buf.lock || buf mod update proto

.PHONY: buf-lint
buf-lint: buf-mod-update
	buf lint

.PHONY: buf-generate
buf-generate: buf-lint
	buf generate

.PHONY: lint
lint: buf-generate
	golangci-lint run

.PHONY: test
test:
	go test -race ./...

.PHONY: build
build: buf-generate
	sqlc generate
	go build -o ./crudapp ./cmd/crudapp

.PHONY: run
run: build
	POSTGRES_CONN_STRING=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

.PHONY: read-all
read-all:
	grpcurl -plaintext -d '{"userId": "${USER_ID}"}' localhost:8080 crudapp.v1.CrudAppService/ReadAll
