# A simple CRUD app

What else to do? I welcome suggestions.
- Add validators to gRPC
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

## Tests

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

## API

### Create

To create a new post for user 1:
```
curl -XPOST -i 127.0.0.1:8080/v1/users/1/posts \
  -H 'Content-Type: application/json' \
  -d '{"data": "a great post"}'
```

### Read

To get user 1's post 2: 
```
curl -XGET -i 127.0.0.1:8080/v1/users/1/posts/2
```

### ReadAll

To get all user 1's posts:
```
curl -XGET -i 127.0.0.1:8080/v1/users/1/posts
```

### Update

To update user 1's post 2: 
```
curl -XPATCH -i 127.0.0.1:8080/v1/users/1/posts/2 \
  -H 'Content-Type: application/json' \
  -d '{"data": "update my great post"}'
```

### Delete

To delete user 1's post 2: 
```
curl -XDELETE -i 127.0.0.1:8080/v1/users/1/posts/2
```
