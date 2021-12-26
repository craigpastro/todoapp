# A simple CRUD app

Still lots to do.
- Tests
- Tracing
- Other storage

## Run the app

Depending on the storage type you want, run one of the following commands. If everything works fine the should be listening on `127.0.0.1:8080`.

```
make run  # defaults to memory
make run-postgres
```

### Postgres

```
docker run --rm --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres:14.1
```

## Create

```
curl -XPOST -i 127.0.0.1:8080/v1/users/1/posts \
  -H 'Content-Type: application/json' \
  -d '{"data": "a great post"}'
```

## Read

To get user 1's post 2: 
```
curl -XGET -i 127.0.0.1:8080/v1/users/1/posts/2
```

## ReadAll

To get all user 1's posts: 
```
curl -XGET -i 127.0.0.1:8080/v1/users/1/posts
```

## Update

To update user 1's post 2: 
```
curl -XPATCH -i 127.0.0.1:8080/v1/users/1/posts/2 \
  -H 'Content-Type: application/json' \
  -d '{"data": "update my great post"}'
```

## Delete

To delete user 1's post 2: 
```
curl -XDELETE -i 127.0.0.1:8080/v1/users/1/posts/2
```

## Run tests

```
go test ./...
```
