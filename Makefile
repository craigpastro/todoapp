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
	go build -o ./bin/crudapp main.go

.PHONY: create-local-dynamodb-table
create-local-dynamodb-table:
	aws dynamodb create-table \
		--table-name Posts \
		--attribute-definitions \
			AttributeName=UserID,AttributeType=S \
			AttributeName=PostID,AttributeType=S \
		--key-schema \
			AttributeName=UserID,KeyType=HASH \
			AttributeName=PostID,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST \
		--endpoint-url http://localhost:8000

.PHONY: create-local-postgres-table
create-local-postgres-table:
	psql postgres://postgres:password@localhost:5432/postgres -c 'CREATE TABLE IF NOT EXISTS post (user_id TEXT NOT NULL, post_id TEXT NOT NULL, data TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, post_id));'

.PHONY: create-all-local-tables
create-all-local-tables: create-local-postgres-table create-local-dynamodb-table

.PHONY: run
run: build
	./bin/crudapp

.PHONY: run-dynamodb
run-dynamodb: build
	STORAGE_TYPE=dynamodb ./bin/crudapp

.PHONY: run-local-dynamodb
run-local-dynamodb:
	DYNAMODB_LOCAL=true make run-dynamodb

.PHONY: run-mongodb
run-mongodb: build
	STORAGE_TYPE=mongodb ./bin/crudapp

.PHONY: run-postgres
run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

.PHONY: run-redis
run-redis: build
	STORAGE_TYPE=redis ./bin/crudapp

.PHONY: create
create:
	curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Create \
	  -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "data": "${DATA}"}'

.PHONY: read
read:
	curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Read \
	  -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'

.PHONY: read-all
read-all:
	curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/ReadAll \
	  -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}"}'

.PHONY: update
update:
	curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Update \
      -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}", "data": "${DATA}"}'

.PHONY: delete
delete:
	curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Update \
      -H 'Content-Type: application/json' \
      -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'
