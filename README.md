# A simple CRUD app

Still lots to do:
- Use grpc-gateway instead of gin
- Tests
- Tracing
- Other storage: dynamodb, ?

## Run the app

Depending on the storage type you want, run one of the following commands.
```
make run  # defaults to memory
make run-postgres
make run-redis
```

You may need the appropriate storage running. If you want to use a container for this purpose you can
```
docker compose up STORAGE_TYPE -d
```
If you are going to run `postgres` the tables will need to be created first; you can `make create-postgres-table` for this purpose. You will need to have `psql` installed.

If everything works fine the service should be listening on `127.0.0.1:8080`.

### Tests

You will need Postgres and Redis running. You can use:
```
docker compose up -d
```
Then
```
make test
```
If you want to bring the containers down:
```
docker compose down
```

## Create

To create a new post for user 1:
```
grpcurl -plaintext -import-path ./api/proto/v1 -proto service.proto -d '{"userId": "1", "data": "a great post"}' 127.0.0.1:8080 api.proto.v1.Service/Create
```

## Read

To get user 1's post 2: 
```
grpcurl -plaintext -import-path ./api/proto/v1 -proto service.proto -d '{"userId": "1", "postId": "2"}' 127.0.0.1:8080 api.proto.v1.Service/Read
```

## ReadAll

To get all user 1's posts:
```
grpcurl -plaintext -import-path ./api/proto/v1 -proto service.proto -d '{"userId": "1"}' 127.0.0.1:8080 api.proto.v1.Service/ReadAll
```

## Update

To update user 1's post 2: 
```
grpcurl -plaintext -import-path ./api/proto/v1 -proto service.proto -d '{"userId": "1", "postId": "2", "data": "update my great post"}' 127.0.0.1:8080 api.proto.v1.Service/Update
```

## Delete

To delete user 1's post 2: 
```
grpcurl -plaintext -import-path ./api/proto/v1 -proto service.proto -d '{"userId": "1", "postId": "2"}' 127.0.0.1:8080 api.proto.v1.Service/Delete
```
