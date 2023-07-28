.PHONY: buf-mod-update
buf-mod-update:
	@test -s ./proto/buf.lock || buf mod update proto

.PHONY: buf-lint
buf-lint: buf-mod-update
	buf lint

.PHONY: generate
generate: buf-lint
	buf format -w
	buf generate
	sqlc generate

.PHONY: lint
lint: generate
	golangci-lint run

.PHONY: test
test:
	go test -race

.PHONY: build
build: generate
	go build -o ./todoapp ./cmd/todoapp

.PHONY: run
run: build
	POSTGRES_CONN_STRING=postgres://postgres:password@127.0.0.1:5432/postgres ./todoapp
