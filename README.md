# A simple CRUD app

A simple CRUD app to demonstrate concepts.

## Run crudapp

Run crudapp by

```
docker compose up -d
make run
```

## Usage

Create a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Create \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "data": "my first post"}'
{"postId":"01H2KJVPWRSQCMZ3TQMHS1SGSQ", "createdAt":"2023-06-10T21:19:40.446394Z"}
```

Read a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Read \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2KJVPWRSQCMZ3TQMHS1SGSQ"}'
{"userId":"foo", "postId":"01H2KJVPWRSQCMZ3TQMHS1SGSQ", "data":"my first post", "createdAt":"2023-06-10T21:19:40.446394Z", "updatedAt":"2023-06-10T21:19:40.446394Z"}
```

Read all posts:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/ReadAll \
-H 'Content-Type: application/json' \
-d '{"userId": "foo"}'
```

Update a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Upsert \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2KJVPWRSQCMZ3TQMHS1SGSQ", "data": "my first updated post"}'
{"postId":"01H2KJVPWRSQCMZ3TQMHS1SGSQ", "updatedAt":"2023-06-10T21:22:28.057756Z"}%
```

Delete a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Delete \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2KJVPWRSQCMZ3TQMHS1SGSQ"}'
{}
```

## Tests

Run

```
make test
```

## Tracing

If you `docker compose up -d` then you should have
[Jaeger](https://www.jaegertracing.io/) running. You can access Jaeger at
http://localhost:16686.
