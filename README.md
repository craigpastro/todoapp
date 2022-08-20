# A simple CRUD app

A simple CRUD app to demonstrate concepts.

## Things to do

- Switch back to env vars only. go-envconfig looks fine.
- Implement preshared key auth
- Add health check
- Add migrate command
- Implement Redis cache
- Add develop logging

And, of course, I welcome suggestions.

## Run the app

Depending on the storage type you want, run one of the following commands.
```
make run-memory
make run-mongodb
make run-postgres
```

You may need the appropriate storage running. If you want to use a container for this purpose you can
```
docker compose up STORAGE_TYPE -d
```
For `postgres` the tables will need to be created first; you can `make create-local-postgres-table` for this purpose. You will need to have `psql`.

If everything works properly the service should be listening on `localhost:8080`.

## Tests

Run
```
make test
```

## API

### Create

To create a new post for user 1:
```
make USER_ID=1 DATA='create a great post' create
```
which just calls
```
curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Create \
  -H 'Content-Type: application/json' \
  -d '{"userId": "1", "data": "create a great post"}'
```

### Read

To get user 1's post 2: 
```
make USER_ID=1 POST_ID=2 read
```
which just calls
```
curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Read \
	-H 'Content-Type: application/json' \
  -d '{"userId": "1", "postId": "2"}'
```

### ReadAll

ReadAll is a streaming endpoint.

To get all user 1's posts:
```
make USER_ID=1 read-all
```
which just calls
```
grpcurl -plaintext -d '{"userId": "1"}' localhost:8080 crudapp.v1.CrudAppService/ReadAll
```
Note this is a grpcurl command. My attempts to get a streaming response with curl aren't working yet.

### Update

To update user 1's post 2: 
```
make USER_ID=1 POST_ID=2 DATA='update my great post' update
```
which just calls
```
curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Update \
  -H 'Content-Type: application/json' \
  -d '{"userId": "1", "postId": "2", "data": "update my great post"}'
```

### Delete

To delete user 1's post 2: 
```
make USER_ID=1 POST_ID=2 delete
```
which just calls
```
curl -XPOST -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Delete \
  -H 'Content-Type: application/json' \
  -d '{"userId": "1", "postId": "2"}'
```

## Tracing

If you `docker compose up -d` then you should have [Jaeger](https://www.jaegertracing.io/) and [Zipkin](https://zipkin.io/) running. You can access Jaeger at http://localhost:16686 and Zipkin at http://localhost:9411.
