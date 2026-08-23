create table labour_hour_entries (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    user_id bigint not null references users(id) on delete cascade,
    entry_date date not null,
    total_hours numeric(10,2) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (user_id, entry_date)
);
create index labour_hour_entries_branch_date_idx on labour_hour_entries (branch_id, entry_date);
