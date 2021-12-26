test:
	go test ./...

build:
	go build -o ./bin/crudapp main.go

run: build
	./bin/crudapp

run-postgres: build
	STORAGE_TYPE=postgres POSTGRES_URI=postgres://postgres:password@127.0.0.1:5432/postgres ./bin/crudapp

