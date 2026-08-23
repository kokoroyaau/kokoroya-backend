create table if not exists users
(
    id            bigserial
        primary key,
    name          text                                   not null,
    email         text                                   not null
        unique,
    password_hash text                                   not null,
    role          text                                   not null,
    phone         text,
    is_active     boolean                  default true  not null,
    rate_weekday  numeric(10, 2),
    rate_weekend  numeric(10, 2),
    created_at    timestamp with time zone default now() not null,
    updated_at    timestamp with time zone default now() not null
);
