# A simple CRUD app

Still lots to do:
- Use grpc-gateway instead of gin
- Tests
- Tracing
- Other storage: redis, dynamodb, ...?

## Run the app

Depending on the storage type you want, run one of the following commands. If everything works fine the should be listening on `127.0.0.1:8080`.

```
make run  # defaults to memory
make run-postgres
```

### Tests

You will need Postgres running. You can use:
```
docker compose up -d
```
Then
```
make test
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
