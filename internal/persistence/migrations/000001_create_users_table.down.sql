create table users
(
    user_id         bigint primary key generated always as identity,
    user_name       text                     not null,
    user_password   text                     not null,
    user_created_at timestamp with time zone not null,
    user_last_login timestamp with time zone not null,
);