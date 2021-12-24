test:
	go test ./...

build:
	go build -o ./bin/crudapp main.go

run: build
	./bin/crudapp
