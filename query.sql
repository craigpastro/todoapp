-- name: Create :one
insert into post (user_id, post_id, data, created_at, updated_at)
values ($1, $2, $3, NOW(), NOW())
returning *;

-- name: Read :one
select *
from post
where user_id = $1 and post_id = $2;

-- name: ReadPage :many
select *
from post
where user_id = $1
limit $2
offset $3;

-- name: Upsert :one
insert into post (user_id, post_id, data, created_at, updated_at)
values ($1, $2, $3, NOW(), NOW())
on conflict (user_id, post_id)
do update set data = $3, updated_at = NOW()
returning *;

-- name: Delete :exec
delete from post
where user_id = $1 and post_id = $2;
