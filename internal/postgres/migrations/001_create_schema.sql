-- +goose Up
create schema crudapp;

create table crudapp.post (
    id bigint generated always as identity not null,
    user_id text default current_setting('crudapp.request.user_id')::text not null,
    post_id text default gen_random_uuid() not null,
    data text not null,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    primary key (user_id, post_id)
);

-- authenticator is a role to log into Postgres
create role authenticator noinherit login password 'password';

-- crudapp_user is for authenticated users
create role crudapp_user nologin;

-- add authenticator to crudapp_user so it can do what crudapp_user can do
grant crudapp_user to authenticator;

grant usage on schema crudapp to crudapp_user;
grant all on crudapp.post to crudapp_user;


-- +goose Down
drop role crudapp_user;
drop role authenticator;
drop table post;
drop schema crudapp;
