create table users
(
    user_id         uuid primary key,

    user_name       text unique              not null,
    user_password   text                     not null,
    user_created_at timestamp with time zone not null,
    user_last_login timestamp with time zone not null
);