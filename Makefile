test:
	go test ./...

build:
	go build -o ./bin/crudapp main.go

build-protos:
	protoc \
		-I ./protos \
		--go_out=./protos --go_opt=paths=source_relative \
		--go-grpc_out=./protos --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./protos \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=logtostderr=true \
		--grpc-gateway_opt=generate_unbound_methods=true \
		./protos/api/v1/service.proto

create-postgres-table:
	psql postgres://postgres:password@localhost:5432/postgres -c 'CREATE TABLE IF NOT EXISTS post (user_id TEXT NOT NULL, post_id TEXT NOT NULL, data TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, post_id));'

run: build
	./bin/crudapp

run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

run-redis: build
	STORAGE_TYPE=redis ./bin/crudapp



