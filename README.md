# A simple TODO app

A simple TODO app that uses Buf Connect, JWTs for authentication, OTEL for
tracing, and sqlc and Postgres for storage.

## Run todoapp

```
docker compose up -d
make run
```

## Usage

Create a JWT with the `sub` claim set to a user id, or just use

```
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJtcl9yb2JvdG8ifQ.oUD_0r5Q1H_akjeJFWYAxbcr2fckBEb7M25wVJw432Y"
```

with the default secret.

Create a post:

```
$ curl -XPOST http://localhost:8080/todoapp.v1.TodoAppService/Create \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"todo": "buy some carrots"}'
{"post":{"userId":"mr_roboto","todoId":"6086008b-4706-4245-8f4e-58ed3eba43d7","todo":"buy some carrots","createdAt":"2023-06-15T18:20:56.235695Z","updatedAt":"2023-06-15T18:20:56.235695Z"}}
```

Read a post:

```
$ curl -XPOST http://localhost:8080/todoapp.v1.TodoAppService/Read \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"todoId": "6086008b-4706-4245-8f4e-58ed3eba43d7"}'
{"post":{"userId":"mr_roboto","todoId":"6086008b-4706-4245-8f4e-58ed3eba43d7","todo":"buy some carrots","createdAt":"2023-06-15T18:20:56.235695Z","updatedAt":"2023-06-15T18:20:56.235695Z"}}
```

Read all posts:

```
$ curl -XPOST http://localhost:8080/todoapp.v1.TodoAppService/ReadAll \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{}'
{"posts":[{"userId":"mr_roboto","todoId":"6086008b-4706-4245-8f4e-58ed3eba43d7","todo":"buy some carrots","createdAt":"2023-06-15T18:20:56.235695Z","updatedAt":"2023-06-15T18:20:56.235695Z"}],"lastIndex":"1"}
```

Update a post:

```
$ curl -XPOST http://localhost:8080/todoapp.v1.TodoAppService/Update \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"todoId": "6086008b-4706-4245-8f4e-58ed3eba43d7", "todo": "buy onions"}'
{"post":{"userId":"mr_roboto","todoId":"6086008b-4706-4245-8f4e-58ed3eba43d7","data":"buy onions","createdAt":"2023-06-15T18:20:56.235695Z","updatedAt":"2023-06-15T18:22:18.689477Z"}}
```

Delete a post:

```
$ curl -XPOST http://localhost:8080/todoapp.v1.TodoAppService/Delete \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"todoId": "6086008b-4706-4245-8f4e-58ed3eba43d7"}'
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
