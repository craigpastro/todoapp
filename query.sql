-- name: Create :one
insert into todoapp.todo (user_id, todo)
values ($1, $2)
returning *;

-- name: Read :one
select *
from todoapp.todo
where user_id = $1 and todo_id = $2;

-- name: ReadPage :many
select *
from todoapp.todo
where user_id = $1
and id > $2
order by id asc
limit 100;

-- name: Update :one
update todoapp.todo
set todo = $1, updated_at = NOW()
where user_id  = $2 AND todo_id = $3
returning *;

-- name: Delete :exec
delete from todoapp.todo
where user_id = $1 and todo_id = $2;
