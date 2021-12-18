# A simple CRUD app

## Run the app

Execute `go run main.go` and (by default) the service should be listening on `127.0.0.1:8080`.

## Create

```
curl -XPOST -i 127.0.0.1:8080/v1/users/1/posts \
  -H 'Content-Type: application/json' \
  -d '{"data": "a great post"}'
```

## Read

To get user 1's, post 2: 
```
curl -XGET -i 127.0.0.1:8080/v1/users/1/posts/2
```

## Update

To update user 1's, post 2: 
```
curl -XPATCH -i 127.0.0.1:8080/v1/users/1/posts/2
```

## Delete

To delete user 1's, post 2: 
```
curl -XDELETE -i 127.0.0.1:8080/v1/users/1/posts/2
```

## Run tests

```
go test
```