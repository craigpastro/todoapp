.PHONY: download
download:
	@cd tools && go mod download

.PHONY: install-tools
install-tools: download
	@cd tools && go list -f '{{range .Imports}}{{.}} {{end}}' tools.go | xargs go install

.PHONY: buf-mod-update
buf-mod-update: install-tools
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
test: buf-generate
	go test -v ./...

.PHONY: build
build: buf-generate
	sqlc generate
	go build -o ./bin/crudapp ./cmd/crudapp

.PHONY: create-local-postgres-table
create-local-postgres-table:
	psql postgres://postgres:password@localhost:5432/postgres -c 'CREATE TABLE IF NOT EXISTS post (user_id TEXT NOT NULL, post_id TEXT NOT NULL, data TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, post_id));'

.PHONY: run-memory
run-memory: build
	./bin/crudapp

.PHONY: run-postgres
run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

.PHONY: create
create:
	curl -XPOST -i http://localhost:8080/crudapp.v1.CrudAppService/Create \
	  -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "data": "${DATA}"}'

.PHONY: read
read:
	curl -XPOST -i http://localhost:8080/crudapp.v1.CrudAppService/Read \
	  -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'

.PHONY: read-all
read-all:
	grpcurl -plaintext -d '{"userId": "${USER_ID}"}' localhost:8080 crudapp.v1.CrudAppService/ReadAll

.PHONY: upsert
upsert:
	curl -XPOST -i http://localhost:8080/crudapp.v1.CrudAppService/Upsert \
      -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}", "data": "${DATA}"}'

.PHONY: delete
delete:
	curl -XPOST -i http://localhost:8080/crudapp.v1.CrudAppService/Delete \
      -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'
