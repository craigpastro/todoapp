# A simple CRUD app

## Help

If anything looks weird or can be better, please help me out! Create an issue or PR.

## Things to do

- Switch back to env vars only
- Add retry to storage startups (pings)
- Implement preshared key auth
- Do streaming with read all
- Add health check

And, of course, I welcome suggestions.

## Run the app

Depending on the storage type you want, run one of the following commands.
```
make run  # defaults to memory
make run-dynamodb
make run-mongodb
make run-postgres
make run-redis
```

You may need the appropriate storage running. If you want to use a container for this purpose you can
```
docker compose up STORAGE_TYPE -d
```
For `dynamodb` and `postgres` the the tables will need to be created first; you can `create-local-dynamodb-table` or `create-local-postgres-table` respectively for this purpose. You will need to have `psql` or the aws cli installed.

If everything works properly the service should be listening on `127.0.0.1:8080`.

## Tests

Use
```
docker compose up -d
```
to get all the required services running. Then
```
make test
```
Don't forget to
```
docker compose down
```

## API

### Create

To create a new post for user 1:
```
make USER_ID=1 DATA='update my great post' create
```
which just calls
```
curl -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Create \
  -H 'Content-Type: application/json' \
  -d '{"userId": "${USER_ID}", "data": "${DATA}"}'
```

### Read

To get user 1's post 2: 
```
make USER_ID=1 POST_ID=2 read
```
which just calls
```
curl -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Read \
	-H 'Content-Type: application/json' \
  -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'```
```

### ReadAll

To get all user 1's posts:
```
make USER_ID=1 read-all
```
which just calls
```
curl -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/ReadAll \
  -H 'Content-Type: application/json' \
  -d '{"userId": "${USER_ID}"}'```
```

### Update

To update user 1's post 2: 
```
make USER_ID=1 POST_ID=2 DATA='update my great post' update
```
which just calls
```
curl -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Update \
  -H 'Content-Type: application/json' \
  -d '{"userId": "${USER_ID}", "postId": "${POST_ID}", "data": "${DATA}"}'
```

### Delete

To delete user 1's post 2: 
```
make USER_ID=1 POST_ID=2 delete
```
which just calls
```
curl -i http://127.0.0.1:8080/crudapp.v1.CrudAppService/Update \
  -H 'Content-Type: application/json' \
  -d '{"userId": "${USER_ID}", "postId": "${POST_ID}"}'```
```

## Tracing

If you `docker compose up -d`ed then you should have [Jaeger](https://www.jaegertracing.io/) and [Zipkin](https://zipkin.io/) running. You can access Jaeger at http://localhost:16686 and Zipkin at http://localhost:9411.
