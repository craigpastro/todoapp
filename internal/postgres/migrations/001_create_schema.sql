-- +goose Up
-- create function set_user_id(user_id text) returns void as $$
-- begin
--     perform set_config('crudapp.request.user_id', user_id, false);
-- end;
-- $$ language plpgsql;

create table post (
    id bigint generated always as identity not null,
    user_id text not null,
    post_id text default gen_random_uuid() not null,
    data text not null,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    primary key (user_id, post_id)
);

-- +goose Down
drop table post;