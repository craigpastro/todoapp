-- name: Create :one
insert into crudapp.post (user_id, data)
values ($1, $2)
returning *;

-- name: Read :one
select *
from crudapp.post
where user_id = $1 and post_id = $2;

-- name: ReadPage :many
select *
from crudapp.post
where user_id = $1
and id > $2
order by id asc
limit 100;

-- name: Upsert :one
insert into crudapp.post (user_id, post_id, data, created_at, updated_at)
values ($1, $2, $3, NOW(), NOW())
on conflict (user_id, post_id)
do update set data = $3, updated_at = NOW()
returning *;

-- name: Delete :exec
delete from crudapp.post
where user_id = $1 and post_id = $2;
