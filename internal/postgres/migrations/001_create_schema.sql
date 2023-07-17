-- +goose Up
create schema todoapp;

create table todoapp.todo (
    id bigint generated always as identity not null,
    user_id text not null,
    todo_id text default gen_random_uuid() not null,
    todo text not null,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    primary key (user_id, todo_id)
);

-- authenticator is a role to log into Postgres
create role authenticator noinherit login password 'password';

-- todoapp_user is for authenticated users
create role todoapp_user nologin;

-- add authenticator to todoapp_user so it can do what todoapp_user can do
grant todoapp_user to authenticator;

grant usage on schema todoapp to todoapp_user;
grant all on todoapp.todo to todoapp_user;


-- +goose Down
drop role todoapp_user;
drop role authenticator;
drop table post;
drop schema todoapp;
