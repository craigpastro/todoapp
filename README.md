# A simple CRUD app

A simple CRUD app that uses JWTs for authentication and Postgres for storage.

## Run crudapp

Run crudapp:

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
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Create \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"data": "my first post"}'
{"post":{"userId":"mr_roboto","postId":"01H2VFQYEZ3ESX0604BPSVRVRE","data":"my first post","createdAt":"2023-06-13T22:59:06.852296Z","updatedAt":"2023-06-13T22:59:06.852296Z"}}
```

Read a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Read \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"postId": "01H2VFQYEZ3ESX0604BPSVRVRE"}'
{"post":{"userId":"mr_roboto","postId":"01H2VFQYEZ3ESX0604BPSVRVRE","data":"my first post","createdAt":"2023-06-13T22:59:06.852296Z","updatedAt":"2023-06-13T22:59:06.852296Z"}}
```

Read all posts:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/ReadAll \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{}'
{"posts":[{"userId":"mr_roboto","postId":"01H2VFQYEZ3ESX0604BPSVRVRE","data":"my first post","createdAt":"2023-06-13T22:59:06.852296Z","updatedAt":"2023-06-13T22:59:06.852296Z"}],"lastIndex":"1"}
```

Update a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Upsert \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"postId": "01H2VFQYEZ3ESX0604BPSVRVRE", "data": "my first updated post"}'
{"post":{"userId":"mr_roboto","postId":"01H2VFQYEZ3ESX0604BPSVRVRE","data":"my first updated post","createdAt":"2023-06-13T22:59:06.852296Z","updatedAt":"2023-06-13T23:01:36.365864Z"}}
```

Delete a post:

```
$ curl -XPOST http://localhost:8080/crudapp.v1.CrudAppService/Delete \
-H "Authentication: Bearer $TOKEN" \
-H 'Content-Type: application/json' \
-d '{"postId": "01H2VFQYEZ3ESX0604BPSVRVRE"}'
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
