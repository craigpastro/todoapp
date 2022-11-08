create table post (
    user_id text not null,
    post_id text not null,
    data text not null,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    primary key (user_id, post_id)
);
