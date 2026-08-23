create table schedule_sections (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    name text not null,
    sort_order int not null default 0,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table schedule_shifts (
    id bigserial primary key,
    branch_id bigint not null references branches(id),
    section_id bigint not null references schedule_sections(id),
    user_id bigint not null references users(id),
    shift_date date not null,
    start_time text,
    code text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (section_id, user_id, shift_date)
);

create table schedule_notes (
    branch_id bigint not null references branches(id),
    week_start_date date not null,
    notes text not null default '',
    updated_at timestamptz not null default now(),
    primary key (branch_id, week_start_date)
);
