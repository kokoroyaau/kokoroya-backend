create table branches (
    id bigserial primary key,
    name text not null,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table user_branches (
    user_id bigint not null references users(id) on delete cascade,
    branch_id bigint not null references branches(id) on delete cascade,
    primary key (user_id, branch_id)
);

alter table users add column if not exists tfn text;
