test:
	go test ./...

build:
	go build -o ./bin/crudapp main.go

run: build
	./bin/crudapp

run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

run-redis: build
	STORAGE_TYPE=redis ./bin/crudapp

create-postgres-table:
	psql postgres://postgres:password@localhost:5432/postgres -c 'CREATE TABLE IF NOT EXISTS post (user_id TEXT NOT NULL, post_id TEXT NOT NULL, data TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, PRIMARY KEY (user_id, post_id));'
