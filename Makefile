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
		
create-local-postgres-table:
	psql postgres://postgres:password@localhost:5432/postgres -c 'CREATE TABLE IF NOT EXISTS post (user_id TEXT NOT NULL, post_id TEXT NOT NULL, data TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, post_id));'

create-all-local-tables: create-local-postgres-table create-local-dynamodb-table

test: build-protos
	go test ./...

build-protos:
	protoc \
		-I ./protos \
		--go_out=./protos --go_opt=paths=source_relative \
		--go-grpc_out=./protos --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./protos \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=logtostderr=true \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--openapiv2_out=./protos \
		./protos/api/v1/service.proto

build: build-protos
	go build -o ./bin/crudapp main.go

run: build
	./bin/crudapp

run-cockroachdb: build
	STORAGE_TYPE=cockroachdb ./bin/crudapp

run-dynamodb: build
	STORAGE_TYPE=dynamodb ./bin/crudapp

run-local-dynamodb:
	DYNAMODB_LOCAL=true make run-dynamodb

run-mongodb: build
	STORAGE_TYPE=mongodb ./bin/crudapp

run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

run-redis: build
	STORAGE_TYPE=redis ./bin/crudapp

create:
	curl -XPOST -i 127.0.0.1:8080/v1/users/${USER_ID}/posts \
        -H 'Content-Type: application/json' \
        -d '{"data": "${DATA}"}'

read:
	curl -XGET -i 127.0.0.1:8080/v1/users/${USER_ID}/posts/${POST_ID}

read-all:
	curl -XGET -i 127.0.0.1:8080/v1/users/${USER_ID}/posts

update:
	curl -XPATCH -i 127.0.0.1:8080/v1/users/${USER_ID}/posts/${POST_ID} \
		-H 'Content-Type: application/json' \
		-d '{"data": "${DATA}"}'

delete:
	curl -XDELETE -i 127.0.0.1:8080/v1/users/${USER_ID}/posts/${POST_ID}
