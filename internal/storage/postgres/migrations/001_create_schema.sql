-- +goose Up
create table post (
    id bigint generated always as identity not null,
    user_id text not null,
    post_id text not null,
    data text not null,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    primary key (user_id, post_id)
);

-- +goose Down
drop table post;