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
{"post":{"userId":"foo", "postId":"01H2R9YTEHM1K1YMNWSKT5G21Y", "data":"my first post", "createdAt":"2023-06-12T17:20:17.366576Z", "updatedAt":"2023-06-12T17:20:17.366576Z"}}
```

Read a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Read \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2R9YTEHM1K1YMNWSKT5G21Y"}'                          
{"post":{"userId":"foo", "postId":"01H2R9YTEHM1K1YMNWSKT5G21Y", "data":"my first post", "createdAt":"2023-06-12T17:20:17.366576Z", "updatedAt":"2023-06-12T17:20:17.366576Z"}}
```

Read all posts:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/ReadAll \
-H 'Content-Type: application/json' \
-d '{"userId": "foo"}'
{"posts":[{"userId":"foo", "postId":"01H2R9YTEHM1K1YMNWSKT5G21Y", "data":"my first post", "createdAt":"2023-06-12T17:20:17.366576Z", "updatedAt":"2023-06-12T17:20:17.366576Z"}], "lastIndex":"1"}
```

Update a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Upsert \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2R9YTEHM1K1YMNWSKT5G21Y", "data": "my first updated post"}'
{"post":{"userId":"foo", "postId":"01H2R9YTEHM1K1YMNWSKT5G21Y", "data":"my first updated post", "createdAt":"2023-06-12T17:20:17.366576Z", "updatedAt":"2023-06-12T17:22:27.323652Z"}}
```

Delete a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Delete \
-H 'Content-Type: application/json' \
-d '{"userId": "foo", "postId": "01H2R9YTEHM1K1YMNWSKT5G21Y"}'
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
