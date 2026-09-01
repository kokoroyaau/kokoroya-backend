create table labour_pay_splits (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    user_id bigint not null references users(id) on delete cascade,
    week_start_date date not null,
    weekday_hours numeric(6, 2) not null default 0,
    weekend_hours numeric(6, 2) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (branch_id, user_id, week_start_date)
);
