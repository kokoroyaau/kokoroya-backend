create table time_entries (
    id bigserial primary key,
    user_id bigint not null references users(id) on delete cascade,
    branch_id bigint not null references branches(id) on delete cascade,
    clock_in_at timestamptz not null,
    clock_out_at timestamptz,
    created_at timestamptz not null default now()
);

create index time_entries_user_open_idx on time_entries (user_id) where clock_out_at is null;
create index time_entries_branch_time_idx on time_entries (branch_id, clock_in_at);
