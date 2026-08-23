create table weekly_labour_rates (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    week_start_date date not null,
    weekday_rate numeric(10,2) not null default 0,
    weekend_rate numeric(10,2) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (branch_id, week_start_date)
);
