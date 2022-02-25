# A simple CRUD app

## Help

If anything looks weird or can be better, please help me out! Create an issue or PR.

## Things to do

- Add validators to gRPC. Or perhaps switch to [Twirp](https://github.com/twitchtv/twirp) and find out what they do there.
- Use mTLS for auth (no auth atm)
- Add pagination to read all

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

## Tracing

If you `docker compose up -d`ed then you should have [Jaeger](https://www.jaegertracing.io/) and [Zipkin](https://zipkin.io/) running. You can access Jaeger at http://localhost:16686 and Zipkin at http://localhost:9411.
